package config

import (
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Handler struct {
	URL        string `env:"URL"`
	CORSOrigin string `env:"CORS_ORIGIN"`
}

type Storage struct {
	FSType        string `env:"FS_TYPE" env-default:"local"`
	StoragePath   string `env:"STORAGE_PATH"`
	StorageSize   int64  `env:"STORAGE_SIZE"`
	MinBufferSize int    `env:"MIN_BUFFER_SIZE" env-default:"65536"`
	MaxBufferSize int    `env:"MAX_BUFFER_SIZE" env-default:"5242880"`
}

type Logger struct {
	Level string `env:"LOGGER_LEVEL" env-default:"INFO"`
}

type Config struct {
	Logger
	Handler
	Storage
	Env string `env:"ENV" env-default:"dev"`
}

const filepath = "./.env"

func New() (*Config, error) {
	var c Config

	err := cleanenv.ReadConfig(filepath, &c)
	if errors.Is(err, os.ErrNotExist) {
		err = cleanenv.ReadEnv(&c)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to read config: %w", err)
	}

	return &c, nil
}
