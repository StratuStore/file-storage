package app

import (
	"context"
	"github.com/StratuStore/file-storage/internal/app/connector"
	"github.com/StratuStore/file-storage/internal/app/controller"
	"github.com/StratuStore/file-storage/internal/app/fileio"
	"github.com/StratuStore/file-storage/internal/app/handlers/rest"
	"github.com/StratuStore/file-storage/internal/app/usecases"
	"github.com/StratuStore/file-storage/internal/libs/config"
	"github.com/StratuStore/file-storage/internal/libs/log"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	l, _ := log.New(cfg)

	filesConnector := connector.NewConnector[fileio.File]()
	readersConnector := connector.NewConnector[usecases.Reader]()

	filesController, err := controller.NewController(cfg.StoragePath, cfg.StorageSize)
	if err != nil {
		panic(err)
	}

	useCases := usecases.NewUseCases(filesConnector, readersConnector, filesController, l, cfg.MinBufferSize, cfg.MaxBufferSize)
	handler := rest.NewHandler(useCases, l, cfg)

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	l.Info("running server", slog.String("url", cfg.URL))
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return handler.Start(ctx)
	})

	g.Go(func() error {
		<-gCtx.Done()
		return handler.Stop(ctx)
	})

	if err := g.Wait(); err != nil {
		l.Error("server shutdown", slog.String("err", err.Error()))
	}
}
