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
	MaxSize     int64
	CurrentSize atomic.Int64
	Files       map[uuid.UUID]fileio.File
	path        string
	mx          *sync.RWMutex
}

func NewController(path string, maxSize int64) (*Controller, error) {
	controller := &Controller{
		FileSystem:  osFs{},
		MaxSize:     maxSize,
		CurrentSize: atomic.Int64{},
		Files:       nil,
		path:        path,
		mx:          &sync.RWMutex{},
	}

	files, err := controller.FileSystem.ListDir(path)
	if err != nil {
		return nil, err
	}
	err = controller.parseStorage(path, files)
	if err != nil {
		return nil, err
	}
	if controller.CurrentSize.Load() > maxSize {
		return nil, ErrMaxSizeExceeded
	}

	return controller, nil
}

func (c *Controller) AllocateStorage(size int64) error {
	if c.CurrentSize.Load()+size > c.MaxSize {
		return ErrMaxSizeExceeded
	}

	c.CurrentSize.Add(size)

	return nil
}

func (c *Controller) ReleaseStorage(size int64) error {
	c.CurrentSize.Add(-size)

	return nil
}

func (c *Controller) AllocateAll() (n int, err error) {
	if result := c.MaxSize - c.CurrentSize.Load(); result > 0 {
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
	defer c.mx.Unlock()
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

func (c *Controller) parseStorage(path string, files map[string]int64) (globalErr error) {
	c.Files = make(map[uuid.UUID]fileio.File, len(files))

	for filename, size := range files {
		id, err := uuid.Parse(filename)
		if err != nil {
			errors.Join(globalErr, err)
			continue
		}

		file, err := fileio.NewFile(path, id, c)
		if err != nil {
			errors.Join(globalErr, err)
			continue
		}

		c.Files[id] = file
		c.CurrentSize.Add(size)
	}

	return globalErr
}
