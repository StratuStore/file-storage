package fileio

import (
	"bufio"
	"errors"
	"io"
	"os"
	"sync"
)

type Reader interface {
	io.ReadSeekCloser
	Closed() bool
}

type reader struct {
	file   File
	osFile FsFile
	buffer *bufio.Reader
	mx     sync.Mutex
	v      int
	closed bool
}

func newFileReader(f File, bufferSize, version int) (Reader, error) {
	osFile, err := f.readOpen()
	if err != nil {
		return nil, err
	}

	buffer := bufio.NewReaderSize(osFile, bufferSize)

	return &reader{
		file:   f,
		osFile: osFile,
		buffer: buffer,
		mx:     sync.Mutex{},
		v:      version,
	}, nil
}

func (r *reader) Seek(offset int64, whence int) (_ int64, err error) {
	if r.file.Closed() || r.closed {
		return 0, os.ErrClosed
	}
	if r.file.version() != r.v {
		r.Close()
		return 0, os.ErrClosed
	}

	r.file.rwMx().RLock()
	defer r.file.rwMx().RUnlock()
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

func (r *reader) Closed() bool {
	return r.closed || r.file.Closed()
}

func (r *reader) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.closed = true
	r.buffer = nil

	return r.osFile.Close()
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.file.Closed() || r.closed {
		return 0, os.ErrClosed
	}
	if r.file.version() != r.v {
		r.Close()
		return 0, os.ErrClosed
	}

	r.file.rwMx().RLock()
	defer r.file.rwMx().RUnlock()
	r.mx.Lock()
	defer r.mx.Unlock()

	n, err = r.buffer.Read(p)

	return n, err
}

// mutexes must be triggered before using this func
func (r *reader) position() (int64, error) {
	ret, err := r.osFile.Seek(0, 1)

	buffSize := r.buffer.Buffered()

	return ret - int64(buffSize), err
}
