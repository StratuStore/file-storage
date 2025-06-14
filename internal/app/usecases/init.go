package usecases

import (
	"github.com/StratuStore/file-storage/internal/app/fileio"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"io"
	"log/slog"
)

type UseCases struct {
	FilesConnector    Connector[*FileWithHost]
	ReadersConnector  Connector[Reader]
	StorageController StorageController
	MaxBufferSize     int
	MinBufferSize     int
	l                 *slog.Logger
	serviceToken      string
	client            *resty.Client
}

func NewUseCases(
	filesConnector Connector[*FileWithHost],
	readersConnector Connector[Reader],
	storageController StorageController,
	logger *slog.Logger,
	minBufferSize int,
	maxBufferSize int,
	serviceToken string,
) *UseCases {
	return &UseCases{
		FilesConnector:    filesConnector,
		ReadersConnector:  readersConnector,
		StorageController: storageController,
		MinBufferSize:     minBufferSize,
		MaxBufferSize:     maxBufferSize,
		serviceToken:      serviceToken,
		client:            resty.New(),
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

type Reader interface {
	io.ReadSeekCloser
	Closeder
}

type FileWithHost struct {
	File fileio.File
	Host string
}

func (f *FileWithHost) Writer() (io.WriteCloser, error) {
	return f.File.Writer()
}

func (f *FileWithHost) Closed() bool {
	return f.File.Closed()
}

type StorageController interface {
	AddFile(id uuid.UUID) (fileio.File, error)
	DeleteFile(id uuid.UUID) error
	File(id uuid.UUID) (fileio.File, error)
}

type ErrorWithMessage interface {
	error
	Message() string
}

type errorWithMessage struct {
	message string
}

func newErrorWithMessage(message string) ErrorWithMessage {
	return &errorWithMessage{message: message}
}

func (e *errorWithMessage) Error() string {
	return e.message
}

func (e *errorWithMessage) Message() string {
	return e.message
}
