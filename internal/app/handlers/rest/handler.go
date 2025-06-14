package rest

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/StratuStore/file-storage/internal/app/usecases"
	"github.com/StratuStore/file-storage/internal/libs/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log/slog"
	"net/http"
	"strings"
)

type Handler struct {
	useCases *usecases.UseCases
	l        *slog.Logger
	r        chi.Router
	cfg      *config.Config
	server   *http.Server
}

func NewHandler(useCases *usecases.UseCases, logger *slog.Logger, cfg *config.Config) *Handler {
	return &Handler{
		useCases: useCases,
		l:        logger.With(slog.String("op", "internal.app.handlers.rest.Handler")),
		cfg:      cfg,
		r:        chi.NewRouter(),
	}
}

func (h *Handler) Register() {
	r := h.r

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	//r.Use(middleware.Timeout(60 * time.Second))

	if h.cfg.Env == "dev" {
		r.Use(cors.Handler(cors.Options{
			// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
			AllowedOrigins: []string{"https://*", "http://*"},
			// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		}))
	} else {
		r.Use(cors.Handler(cors.Options{
			// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
			AllowedOrigins: []string{"https://" + h.cfg.CORSOrigin, "http://*" + h.cfg.CORSOrigin},
			// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		}))
	}

	r.Route("/files", func(r chi.Router) {
		r.Get("/read", h.ReadFile)
		r.Post("/write", h.WriteFile)
		r.Post("/close", h.CloseFile)
	})
}

func (h *Handler) Start(ctx context.Context) error {
	h.Register()
	// Init server
	h.server = &http.Server{
		Addr:        h.cfg.URL,
		Handler:     h.r,
		IdleTimeout: h.cfg.IdleTimeout,
	}

	return h.server.ListenAndServe()
}

func (h *Handler) Stop(ctx context.Context) error {
	return h.server.Shutdown(context.Background())
}

func (h *Handler) handleError(w http.ResponseWriter, status int, err error, messages ...string) error {
	var errWithMessage usecases.ErrorWithMessage
	if errors.As(err, &errWithMessage) {
		messages = append(messages, errWithMessage.Message())
	}
	message := strings.Join(messages, ", ")

	w.WriteHeader(status)
	response := ResponseError{Err: message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.l.Error("unable to encode response", slog.String("err", err.Error()))
		return err
	}

	return nil
}

type ResponseError struct {
	Err string `json:"error"`
}

func (e *ResponseError) Error() string {
	return e.Err
}
