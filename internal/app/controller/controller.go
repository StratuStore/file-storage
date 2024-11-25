package controller

import (
	"errors"
	"github.com/StratuStore/file-storage/internal/app/fileio"
	"github.com/google/uuid"
	"os"
	"sync"
	"sync/atomic"
)

var ErrMaxSizeExceeded = errors.New("max size exceeded")

type Controller struct {
	MaxSize     int
	CurrentSize atomic.Int64
	Files       map[uuid.UUID]*fileio.File
	mx          *sync.RWMutex
}

func NewController(filePath string, maxSize int) (*Controller, error) {
	controller := &Controller{
		MaxSize:     maxSize,
		CurrentSize: atomic.Int64{},
		Files:       nil,
		mx:          &sync.RWMutex{},
	}

	files, currentSize, err := parseStorage(filePath, controller)
	if err != nil {
		return nil, err
	}
	if currentSize > maxSize {
		return nil, ErrMaxSizeExceeded
	}

	controller.Files = files
	controller.CurrentSize.Add(int64(currentSize))

	return controller, nil
}

func (c *Controller) AllocateStorage(size int) error {
	if int(c.CurrentSize.Load())+size > c.MaxSize {
		return ErrMaxSizeExceeded
	}

	c.CurrentSize.Add(int64(size))

	return nil
}

func (c *Controller) ReleaseStorage(size int) error {
	c.CurrentSize.Add(-int64(size))

	return nil
}

func (c *Controller) AllocateAll() (n int, err error) {
	if result := int64(c.MaxSize) - c.CurrentSize.Load(); result > 0 {
		return int(result), err
	}

	return 0, ErrMaxSizeExceeded
}

// Don't forget to sync map with mutex!!!
func AddFile(id uuid.UUID) (*fileio.File, error)

func DeleteFile(id uuid.UUID) error

func File(id uuid.UUID) (*fileio.File, error)

func parseStorage(filePath string, controller *Controller) (files map[uuid.UUID]*fileio.File, currentSize int, err error) {
	dir, err := os.ReadDir(filePath)
	if err != nil {
		return nil, 0, err
	}

	files = map[uuid.UUID]*fileio.File{}
	for _, file := range dir {
		if file.IsDir() {
			continue
		}

		stat, err2 := file.Info()
		if err2 != nil {
			errors.Join(err, err2)
			continue
		}
		currentSize += int(stat.Size())

		id, err2 := uuid.Parse(stat.Name())
		if err2 != nil {
			errors.Join(err, err2)
			continue
		}

		file, err2 := fileio.NewFile(filePath, id, controller)
		if err2 != nil {
			errors.Join(err, err2)
			continue
		}

		files[id] = file
	}

	return files, currentSize, err
}
