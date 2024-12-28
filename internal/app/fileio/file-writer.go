package fileio

import (
	"os"
	"path"
	"sync"
)

// io.WriteCloser
type FileWriter struct {
	file   *File
	osFile *os.File
	mx     sync.Mutex
	closed bool
}

func NewFileWriter(f *File) (*FileWriter, error) {
	file, err := createOrOpenFile(path.Join(f.Path, f.ID.String()))
	if err != nil {
		return nil, err
	}

	err = f.controller.ReleaseStorage(f.Size)
	if err != nil {
		return nil, err
	}
	f.Size = 0

	return &FileWriter{
		file:   f,
		osFile: file,
		mx:     sync.Mutex{},
	}, nil
}

func (f *FileWriter) Write(b []byte) (n int, err error) {
	if f.file.closed || f.closed {
		return 0, os.ErrClosed
	}

	f.mx.Lock()
	defer f.mx.Unlock()

	err = f.file.controller.AllocateStorage(len(b))
	if err != nil {
		return 0, err
	}

	return f.osFile.Write(b)
}

func (f *FileWriter) Close() error {
	f.mx.Lock()
	defer f.mx.Unlock()

	err := f.osFile.Close()
	f.closed = true

	f.file.mx.Unlock()

	return err
}
