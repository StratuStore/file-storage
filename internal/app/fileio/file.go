package fileio

import (
	"github.com/google/uuid"
	"io"
	"os"
	"path"
	"sync"
)

type File struct {
	ID         uuid.UUID
	Path       string
	Size       int
	controller StorageController
	mx         *sync.RWMutex
	closed     bool
	v          int
}

func NewFile(filePath string, id uuid.UUID, controller StorageController) (*File, error) {
	file, err := createOrOpenFile(path.Join(filePath, id.String()))
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()
	file.Close()

	return &File{
		ID:         id,
		Path:       filePath,
		Size:       int(size),
		controller: controller,
		mx:         &sync.RWMutex{},
	}, nil
}

// Sync is used when File has been imported from DB and has missing unexported fields
func (f *File) Sync(controller StorageController) error {
	if f.closed {
		return os.ErrClosed
	}
	file, err := createOrOpenFile(path.Join(f.Path, f.ID.String()))
	if err != nil {
		return err
	}
	file.Close()

	f.controller = controller
	f.mx = &sync.RWMutex{}

	return nil
}

func (f *File) Reader(bufferSize int) (Reader, error) {
	if f.closed {
		return nil, os.ErrClosed
	}
	f.mx.RLock()
	defer f.mx.RUnlock()

	reader, err := NewFileReader(f, bufferSize, f.v)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (f *File) Writer() (io.WriteCloser, error) {
	if f.closed {
		return nil, os.ErrClosed
	}
	f.mx.Lock()

	f.v++
	writer, err := NewFileWriter(f)
	if err != nil {
		f.mx.Unlock()
		f.v--
	}

	return writer, err
}

func (f *File) Closed() bool {
	return f.closed
}

func (f *File) Close() error {
	if f.closed {
		return os.ErrClosed
	}
	f.mx.Lock()
	f.closed = true
	f.mx.Unlock()

	return nil
}

func (f *File) Delete() error {
	if f.closed {
		return os.ErrClosed
	}
	f.Close()
	f.mx.Lock()
	defer f.mx.Unlock()

	return deleteFile(f.Path, f.ID.String())
}

func deleteFile(filePath string, name string) error {
	return os.Remove(path.Join(filePath, name))
}

func createOrOpenFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)

	return file, err
}
