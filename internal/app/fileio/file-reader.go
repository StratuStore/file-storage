package fileio

import (
	"bufio"
	"errors"
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

	oldPosition, err := r.position()
	if err != nil {
		return oldPosition, err
	}

	newOffset, err := r.osFile.Seek(offset, whence)
	if err != nil {
		return newOffset, err
	}

	if diff := newOffset - int64(r.buffer.Buffered()) - oldPosition; diff >= 0 {
		_, err = r.buffer.Discard(int(diff))
		off, err2 := r.osFile.Seek(int64(r.buffer.Buffered()), 1)

		return off - int64(r.buffer.Buffered()), errors.Join(err, err2)
	}
	r.buffer.Reset(r.osFile)

	return newOffset, nil
}

func (r *FileReader) Closed() bool {
	return r.closed || r.file.closed
}

func (r *FileReader) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.closed = true
	r.buffer = nil

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

	return n, err
}

// mutexes must be triggered before using this func
func (r *FileReader) position() (int64, error) {
	ret, err := r.osFile.Seek(0, 1)

	buffSize := r.buffer.Buffered()

	return ret - int64(buffSize), err
}
