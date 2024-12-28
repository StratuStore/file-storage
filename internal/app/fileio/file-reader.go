package fileio

import (
	"bufio"
	"io"
	"os"
	"path"
	"sync"
)

type Reader interface {
	io.ReadSeekCloser
	Closed() bool
}

type FileReader struct {
	file   *File
	osFile *os.File
	buffer *bufio.Reader
	mx     sync.Mutex
	pos    int64
	v      int
	closed bool
}

func NewFileReader(f *File, bufferSize, v int) (*FileReader, error) {
	osFile, err := os.Open(path.Join(f.Path, f.ID.String()))
	if err != nil {
		return nil, err
	}

	buffer := bufio.NewReaderSize(osFile, bufferSize)

	return &FileReader{
		file:   f,
		osFile: osFile,
		buffer: buffer,
		mx:     sync.Mutex{},
		v:      v, // version of the file
	}, nil
}

func (r *FileReader) Seek(offset int64, whence int) (_ int64, err error) {
	if r.file.closed || r.closed {
		return 0, os.ErrClosed
	}
	if r.file.v != r.v {
		r.Close()
		return 0, os.ErrClosed
	}

	r.file.mx.RLock()
	defer r.file.mx.RUnlock()
	r.mx.Lock()
	defer r.mx.Unlock()

	oldGlobOffset := r.pos
	r.pos, err = r.osFile.Seek(offset, whence)
	if err != nil {
		return r.pos, err
	}

	if off := r.pos - oldGlobOffset; off >= 0 {
		_, err := r.buffer.Discard(int(off))

		return r.pos, err
	}
	r.buffer.Reset(r.osFile)

	return r.pos, nil
}

func (r *FileReader) Closed() bool {
	return r.closed || r.file.closed
}

func (r *FileReader) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.closed = true

	return r.osFile.Close()
}

func (r *FileReader) Read(p []byte) (n int, err error) {
	if r.file.closed || r.closed {
		return 0, os.ErrClosed
	}
	if r.file.v != r.v {
		r.Close()
		return 0, os.ErrClosed
	}

	r.file.mx.RLock()
	defer r.file.mx.RUnlock()
	r.mx.Lock()
	defer r.mx.Unlock()

	n, err = r.buffer.Read(p)

	r.pos, _ = r.osFile.Seek(0, 1)

	return n, err
}
