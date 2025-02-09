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
	FileSystem
	MaxSize     int
	CurrentSize atomic.Int64
	Files       map[uuid.UUID]fileio.File
	path        string
	mx          *sync.RWMutex
}

func NewController(filePath string, maxSize int) (*Controller, error) {
	controller := &Controller{
		FileSystem:  osFs{},
		MaxSize:     maxSize,
		CurrentSize: atomic.Int64{},
		Files:       nil,
		path:        filePath,
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

func (c *Controller) AddFile(id uuid.UUID) (fileio.File, error) {
	file, err := fileio.NewFile(c.path, id, c)
	if err != nil {
		return nil, err
	}

	c.mx.Lock()
	defer c.mx.RUnlock()
	if _, ok := c.Files[id]; ok {
		return nil, os.ErrExist
	}

	c.Files[id] = file

	return file, nil
}

func (c *Controller) DeleteFile(id uuid.UUID) error {
	c.mx.Lock()

	file, ok := c.Files[id]
	if !ok {
		return os.ErrNotExist
	}

	delete(c.Files, id)

	c.mx.Unlock()

	if err := file.Delete(); err != nil {
		return err
	}

	return nil
}

func (c *Controller) File(id uuid.UUID) (fileio.File, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	if file, ok := c.Files[id]; ok {
		return file, nil
	}

	return nil, os.ErrNotExist
}

func parseStorage(filePath string, controller *Controller) (files map[uuid.UUID]fileio.File, currentSize int, err error) {
	dir, err := os.ReadDir(filePath)
	if err != nil {
		return nil, 0, err
	}

	files = map[uuid.UUID]fileio.File{}
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
