package fileio

import (
	"github.com/google/uuid"
	"io"
	"os"
	"path"
	"sync"
)

//go:generate go run github.com/vektra/mockery/v2@v2.52.3 --name=File --structname=MockFile --filename=mock_file.go --inpackage
type File interface {
	Sync(controller StorageController) error
	Reader(bufferSize int) (Reader, error)
	Writer() (io.WriteCloser, error)
	Close() error
	Delete() error

	ID() uuid.UUID
	FullPath() string
	Size() int
	Closed() bool
	version() int
	rwMx() *sync.RWMutex

	allocate(size int) error
	readOpen() (FsFile, error)
	writeOpen() (FsFile, error)
}

type file struct {
	id         uuid.UUID
	path       string
	size       int
	controller StorageController
	mx         *sync.RWMutex
	closed     bool
	v          int
}

func NewFile(filePath string, id uuid.UUID, controller StorageController) (File, error) {
	f, err := controller.CreateOrWriteOpen(path.Join(filePath, id.String()))
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()
	f.Close()

	return &file{
		id:         id,
		path:       filePath,
		size:       int(size),
		controller: controller,
		mx:         &sync.RWMutex{},
	}, nil
}

// Sync is used when file has been imported from DB and has some missing unexported fields
func (f *file) Sync(controller StorageController) error {
	if f.closed {
		return os.ErrClosed
	}
	file, err := controller.CreateOrWriteOpen(f.FullPath())
	if err != nil {
		return err
	}
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	f.size = int(stat.Size())
	file.Close()

	f.controller = controller
	f.mx = &sync.RWMutex{}

	return nil
}

func (f *file) Reader(bufferSize int) (Reader, error) {
	if f.closed {
		return nil, os.ErrClosed
	}
	f.mx.RLock()
	defer f.mx.RUnlock()

	reader, err := newFileReader(f, bufferSize)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (f *file) Writer() (io.WriteCloser, error) {
	if f.closed {
		return nil, os.ErrClosed
	}
	f.mx.Lock()

	f.v++

	size := f.size
	err := f.controller.ReleaseStorage(f.size)
	if err != nil {
		return nil, err
	}
	f.size = 0

	writer, err := newFileWriter(f)
	if err != nil {
		f.v--
		f.size = size
		f.mx.Unlock()
	}

	return writer, err
}

func (f *file) Close() error {
	if f.closed {
		return os.ErrClosed
	}
	f.mx.Lock()
	f.closed = true
	f.mx.Unlock()

	return nil
}

func (f *file) Delete() error {
	if f.closed {
		return os.ErrClosed
	}
	f.Close()
	f.mx.Lock()
	defer f.mx.Unlock()

	return f.controller.FSDelete(f.FullPath())
}

func (f *file) Closed() bool {
	return f.closed
}

func (f *file) FullPath() string {
	return path.Join(f.path, f.id.String())
}

func (f *file) ID() uuid.UUID {
	return f.id
}
func (f *file) Size() int {
	return f.size
}

func (f *file) allocate(size int) error {
	return f.controller.AllocateStorage(size)
}

func (f *file) rwMx() *sync.RWMutex {
	return f.mx
}

func (f *file) readOpen() (FsFile, error) {
	return f.controller.ReadOpen(f.FullPath())
}

func (f *file) writeOpen() (FsFile, error) {
	return f.controller.CreateOrWriteOpen(f.FullPath())
}

func (f *file) version() int {
	return f.v
}
