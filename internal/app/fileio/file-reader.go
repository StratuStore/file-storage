package fileio

import (
	"bufio"
	"errors"
	"io"
	"os"
	"sync"
	"weak"
)

type Reader interface {
	io.ReadSeekCloser
	Closed() bool
}

type reader struct {
	file       File
	osFile     FsFile
	buffer     weak.Pointer[bufio.Reader]
	bufferSize int
	mx         sync.Mutex
	v          int
	closed     bool
	pos        int64
}

func newFileReader(f File, bufferSize int) (Reader, error) {
	osFile, err := f.openForReading()
	if err != nil {
		return nil, err
	}

	buffer := weak.Make(bufio.NewReaderSize(osFile, bufferSize))

	return &reader{
		file:       f,
		osFile:     osFile,
		buffer:     buffer,
		bufferSize: bufferSize,
		mx:         sync.Mutex{},
		v:          f.version(),
	}, nil
}

func (r *reader) Seek(offset int64, whence int) (n int64, err error) {
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

	buffer := r.buffer.Value()
	if buffer == nil {
		n, err = r.osFile.Seek(offset, whence)
		r.pos = n

		return n, err
	}

	newOffset, err := r.osFile.Seek(offset, whence)
	if err != nil {
		r.pos = newOffset
		return newOffset, err
	}

	if diff := newOffset - r.pos; diff >= 0 && diff < int64(buffer.Buffered()) {
		_, err = buffer.Discard(int(diff))
		off, err2 := r.osFile.Seek(int64(buffer.Buffered()), io.SeekCurrent)
		r.pos = off - int64(buffer.Buffered())

		return r.pos, errors.Join(err, err2)
	}
	buffer.Reset(r.osFile)
	r.pos = newOffset

	return r.pos, nil
}

func (r *reader) Closed() bool {
	return r.closed || r.file.Closed()
}

func (r *reader) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.closed = true

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

	buffer := r.buffer.Value()
	if buffer == nil {
		_, err = r.osFile.Seek(r.pos, io.SeekStart)
		buffer = bufio.NewReaderSize(r.osFile, r.bufferSize)
		r.buffer = weak.Make(buffer)
	}

	n, err = buffer.Read(p)
	r.pos += int64(n)

	return n, err
}
