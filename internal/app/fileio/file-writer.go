package fileio

import (
	"io"
	"os"
	"sync"
)

type writer struct {
	ownFile File
	osFile  FsFile
	mx      sync.Mutex
	closed  bool
}

func newFileWriter(f *file) (io.WriteCloser, error) {
	file, err := f.writeOpen()
	if err != nil {
		return nil, err
	}

	return &writer{
		ownFile: f,
		osFile:  file,
		mx:      sync.Mutex{},
	}, nil
}

func (f *writer) Write(b []byte) (n int, err error) {
	if f.ownFile.Closed() || f.closed {
		return 0, os.ErrClosed
	}

	f.mx.Lock()
	defer f.mx.Unlock()

	err = f.ownFile.allocate(len(b))
	if err != nil {
		return 0, err
	}

	return f.osFile.Write(b)
}

func (f *writer) Close() error {
	f.mx.Lock()
	defer f.mx.Unlock()

	err := f.osFile.Close()
	f.closed = true

	f.ownFile.rwMx().Unlock()

	return err
}
