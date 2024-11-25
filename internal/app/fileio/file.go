package fileio

import (
	"github.com/google/uuid"
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
	file, err := createOrOpenFile(path.Join(f.Path, f.ID.String()))
	if err != nil {
		return err
	}
	file.Close()

	f.controller = controller
	f.mx = &sync.RWMutex{}

	return nil
}

func (f *File) Reader() (Reader, error) {
	if f.closed {
		return nil, os.ErrClosed
	}
	f.mx.RLock()
	defer f.mx.RUnlock()

	reader, err := NewFileReader(f)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (f *File) Write(b []byte) (n int, err error) {
	return f.WriteAt(b, 0)
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	if f.closed {
		return 0, os.ErrClosed
	}

	f.mx.Lock()
	defer f.mx.Unlock()

	file, err := createOrOpenFile(path.Join(f.Path, f.ID.String()))
	if err != nil {
		return 0, err
	}
	defer file.Close()

	sizeChange := len(b) - f.Size
	if sizeChange > 0 {
		err = f.controller.AllocateStorage(sizeChange)
		if err != nil {
			return 0, err
		}
	}

	n, err = file.WriteAt(b, off)

	if sizeChange < 0 {
		err = f.controller.ReleaseStorage(-sizeChange)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (f *File) Close() error {
	f.mx.Lock()
	f.closed = true
	f.mx.Unlock()

	return nil
}

func (f *File) Delete() error {
	f.Close()
	f.mx.Lock()
	defer f.mx.Unlock()

	return deleteFile(f.Path, f.ID.String())
}

func deleteFile(filePath string, name string) error {
	return os.Remove(path.Join(filePath, name))
}

func createOrOpenFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0o666)

	return file, err
}
