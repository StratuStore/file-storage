package usecases

import (
	"github.com/StratuStore/file-storage/internal/app/fileio"
	"github.com/google/uuid"
	"io"
)

type UseCases struct {
	FilesConnector    Connector[fileio.File]
	ReadersConnector  Connector[Reader]
	StorageController StorageController
	MaxBufferSize     int
	MinBufferSize     int
}

type Connector[V Closer] interface {
	OpenConnection(value V) (uuid.UUID, error)
	Connection(id uuid.UUID) (V, error)
}

type Closer interface {
	Closed() bool
	io.Closer
}

type File interface {
	io.Writer
	io.WriterAt
	io.Closer
	Delete() error
}

type Reader interface {
	io.ReadSeekCloser
	Closer
}

type StorageController interface {
	AddFile(id uuid.UUID) (fileio.File, error)
	DeleteFile(id uuid.UUID) error
	File(id uuid.UUID) (fileio.File, error)
}
