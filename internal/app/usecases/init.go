package usecases

import (
	"github.com/StratuStore/file-storage/internal/app/fileio"
	"github.com/google/uuid"
	"io"
	"log/slog"
)

type UseCases struct {
	FilesConnector    Connector[fileio.File]
	ReadersConnector  Connector[Reader]
	StorageController StorageController
	MaxBufferSize     int
	MinBufferSize     int
	l                 *slog.Logger
}

func NewUseCases(
	filesConnector Connector[fileio.File],
	readersConnector Connector[Reader],
	storageController StorageController,
	maxBufferSize int,
	minBufferSize int,
	logger *slog.Logger,
) *UseCases {
	return &UseCases{
		FilesConnector:    filesConnector,
		ReadersConnector:  readersConnector,
		StorageController: storageController,
		MaxBufferSize:     maxBufferSize,
		MinBufferSize:     minBufferSize,
		l:                 logger.With(slog.String("op", "internal.app.usecases.UseCases")),
	}
}

type Connector[V Closeder] interface {
	OpenConnection(value V) (uuid.UUID, error)
	Connection(id uuid.UUID) (V, error)
}

type Closeder interface {
	Closed() bool
}

type File interface {
	io.Writer
	io.WriterAt
	io.Closer
	Delete() error
}

type Reader interface {
	io.ReadSeekCloser
	Closeder
}

type StorageController interface {
	AddFile(id uuid.UUID) (fileio.File, error)
	DeleteFile(id uuid.UUID) error
	File(id uuid.UUID) (fileio.File, error)
}
