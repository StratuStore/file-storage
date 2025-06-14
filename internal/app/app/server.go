package app

import (
	"context"
	"errors"
	"github.com/StratuStore/file-storage/internal/app/connector"
	"github.com/StratuStore/file-storage/internal/app/controller"
	"github.com/StratuStore/file-storage/internal/app/handlers/queue"
	"github.com/StratuStore/file-storage/internal/app/handlers/rest"
	"github.com/StratuStore/file-storage/internal/app/usecases"
	"github.com/StratuStore/file-storage/internal/libs/config"
	"github.com/StratuStore/file-storage/internal/libs/log"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	l, _ := log.New(cfg)

	filesConnector := connector.NewConnector[*usecases.FileWithHost]()
	readersConnector := connector.NewConnector[usecases.Reader]()

	filesController, err := controller.NewController(cfg.StoragePath, cfg.StorageSize)
	if err != nil {
		panic(err)
	}

	useCases := usecases.NewUseCases(filesConnector, readersConnector, filesController, l, cfg.MinBufferSize, cfg.MaxBufferSize, cfg.Token)
	handler := rest.NewHandler(useCases, l, cfg)
	queueHandler, err := queue.New(l, cfg, useCases, filesController)
	if err != nil {
		panic(err)
	}

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	l.Info("running server", slog.String("url", cfg.URL))
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return handler.Start(ctx)
	})

	g.Go(func() error {
		filesConnector.StartDisposalRoutine(time.Duration(cfg.GC.SleepInMinutes)*time.Minute, time.Duration(cfg.GC.KeepAliveInMinutes)*time.Minute)
		readersConnector.StartDisposalRoutine(time.Duration(cfg.GC.SleepInMinutes)*time.Minute, time.Duration(cfg.GC.KeepAliveInMinutes)*time.Minute)

		return nil
	})

	g.Go(func() error {
		return queueHandler.Start(ctx)
	})

	g.Go(func() error {
		<-gCtx.Done()

		err := queueHandler.Stop(ctx)
		err = errors.Join(err, handler.Stop(ctx))
		return err
	})

	if err := g.Wait(); err != nil {
		l.Error("server shutdown", slog.String("err", err.Error()))
	}
}
