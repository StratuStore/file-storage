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
	file, err := f.openForWriting()
	if err != nil {
		return nil, err
	}

	return &writer{
		ownFile: f,
		osFile:  file,
		mx:      sync.Mutex{},
	}, nil
}

func (w *writer) Write(b []byte) (n int, err error) {
	if w.ownFile.Closed() || w.closed {
		return 0, os.ErrClosed
	}

	w.mx.Lock()
	defer w.mx.Unlock()

	err = w.ownFile.allocate(len(b))
	if err != nil {
		return 0, err
	}

	return w.osFile.Write(b)
}

func (w *writer) Close() error {
	w.mx.Lock()
	defer w.mx.Unlock()

	err := w.osFile.Close()
	w.closed = true

	w.ownFile.rwMx().Unlock()

	return err
}
