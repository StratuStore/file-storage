package log

import (
	"fmt"
	"github.com/StratuStore/file-storage/internal/libs/config"
	"log/slog"
	"os"
)

func New(cfg *config.Config) (*slog.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("unable to parse log level: %w", err)
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})

	return slog.New(handler), nil
}
