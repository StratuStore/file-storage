package queue

import (
	"context"
	"errors"
	"fmt"
	"github.com/StratuStore/file-storage/internal/app/controller"
	"github.com/StratuStore/file-storage/internal/app/usecases"
	"github.com/StratuStore/file-storage/internal/libs/config"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"net/http"
	"sync"
)

type Handler struct {
	l        *slog.Logger
	sub      *amqp.Subscriber
	useCases *usecases.UseCases
	ctrl     *controller.Controller
	client   *resty.Client
	topic    string
	host     string
	token    string
	g        GobMarshaler
}

func New(l *slog.Logger, cfg *config.Config, uc *usecases.UseCases, ctrl *controller.Controller) (*Handler, error) {
	sub, err := amqp.NewSubscriber(
		amqp.NewNonDurablePubSubConfig(cfg.RabbitMQ.URN, func(topic string) string { return cfg.QueueName }),
		watermill.NewSlogLogger(l.With(slog.String("module", "watermill-ampq"))),
	)
	if err != nil {
		return nil, err
	}

	return &Handler{
		l:        l.With("internal.app.handlers.queue"),
		sub:      sub,
		useCases: uc,
		ctrl:     ctrl,
		client:   resty.New(),
		topic:    cfg.RabbitMQ.Topic,
		host:     cfg.RabbitMQ.Host,
		token:    cfg.Token,
		g:        GobMarshaler{},
	}, nil
}

func (h *Handler) Start(ctx context.Context) error {
	ch, err := h.sub.Subscribe(ctx, h.topic)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for msg := range ch {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err2 := h.handle(msg); err2 != nil {
				err = errors.Join(err, err2)
			}
		}()
	}

	wg.Wait()

	return err
}

func (h *Handler) Stop(ctx context.Context) error {
	return h.sub.Close()
}

func (h *Handler) handle(msg *message.Message) error {
	l := h.l.With(slog.String("op", "handle"))

	var request Request
	if err := h.g.Unmarshal(msg.Payload, &request); err != nil {
		l.Error("unable to unmarshal queue payload using gob", slog.String("err", err.Error()))

		msg.Nack()
		return fmt.Errorf("unable to unmarshal queue payload using gob: %w", err)
	}

	msg.Ack()

	ctx := context.Background()
	response, revert, err := h.processRequest(ctx, &request)
	if err != nil {
		return err
	}
	responseBody, err := h.g.Marshal(response)
	if err != nil {
		l.Error("unable to marshal responseBody", slog.String("err", err.Error()))

		return err
	}

	result, err := h.client.R().
		SetContext(ctx).
		SetAuthScheme("Bearer").
		SetAuthToken(h.token).
		SetBody(responseBody).
		Post(request.Host)
	if err != nil {
		l.Error("unable to make request to fsm", slog.String("err", err.Error()))
		revert()

		return err
	}
	if result.StatusCode() == http.StatusResetContent {
		revert()
	}

	return nil
}

func (h *Handler) processRequest(ctx context.Context, r *Request) (response *Response, revert func(), _ error) {
	l := h.l.With(slog.String("op", "processRequest"))

	switch r.Type {
	case CreateType:
		if err := h.ctrl.TryAllocateStorage(int64(r.Size)); err != nil {
			l.Warn("we are full!")
			return nil, nil, err
		}

		connectionID, err := h.useCases.CreateFile(ctx, r.FileID)
		if err != nil {
			l.Error("unable to create file", slog.String("err", err.Error()))

			return nil, nil, err
		}
		response = &Response{
			ID:           r.ID,
			Host:         h.host,
			ConnectionID: connectionID,
			Err:          "",
		}
		revert = func() {
			if err := h.useCases.DeleteFile(ctx, r.FileID); err != nil {
				l.Error("unable to delete file", slog.String("err", err.Error()))
			}
		}
	case UpdateType:
		if err := h.ctrl.TryAllocateStorage(int64(r.Size)); err != nil {
			l.Warn("we are full!")
			return nil, nil, err
		}
		if _, err := h.ctrl.File(r.FileID); err != nil {
			l.Info("we have no such file", slog.String("err", err.Error()))

			return nil, nil, err
		}

		connectionID, err := h.useCases.UpdateFile(ctx, r.FileID)
		var errString string
		if err != nil {
			l.Error("unable to update file", slog.String("err", err.Error()))

			errString = err.Error()
		}
		response = &Response{
			ID:           r.ID,
			Host:         h.host,
			ConnectionID: connectionID,
			Err:          errString,
		}
		revert = func() {}
	case OpenType:
		if _, err := h.ctrl.File(r.FileID); err != nil {
			l.Info("we have no such file", slog.String("err", err.Error()))

			return nil, nil, err
		}

		connectionID, err := h.useCases.OpenFile(ctx, r.FileID)
		var errString string
		if err != nil {
			l.Error("unable to open file", slog.String("err", err.Error()))

			errString = err.Error()
		}
		response = &Response{
			ID:           r.ID,
			Host:         h.host,
			ConnectionID: connectionID,
			Err:          errString,
		}
		revert = func() {
			if err := h.useCases.Close(ctx, connectionID); err != nil {
				l.Error("unable to close connection", slog.String("err", err.Error()))
			}
		}
	case DeleteType:
		if _, err := h.ctrl.File(r.FileID); err != nil {
			l.Info("we have no such file", slog.String("err", err.Error()))

			return nil, nil, err
		}

		err := h.useCases.DeleteFile(ctx, r.FileID)
		var errString string
		if err != nil {
			l.Error("unable to update file", slog.String("err", err.Error()))

			errString = err.Error()
		}
		response = &Response{
			ID:   r.ID,
			Host: h.host,
			Err:  errString,
		}
		revert = func() {}
	default:
		return nil, nil, fmt.Errorf("wrong request type")
	}

	return response, revert, nil
}
