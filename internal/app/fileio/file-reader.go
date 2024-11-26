package fileio

import (
	"bufio"
	"os"
	"path"
	"sync"
)

type FileReader struct {
	file   *File
	osFile *os.File
	buffer *bufio.Reader
	mx     *sync.Mutex
}

func NewFileReader(f *File, size int) (*FileReader, error) {
	osFile, err := os.Open(path.Join(f.Path, f.ID.String()))
	if err != nil {
		return nil, err
	}

	buffer := bufio.NewReaderSize(osFile, size)

	return &FileReader{
		file:   f,
		osFile: osFile,
		buffer: buffer,
		mx:     &sync.Mutex{},
	}, nil
}

func (r *FileReader) Seek(offset int64, whence int) (int64, error) {
	if r.file.closed {
		return 0, os.ErrClosed
	}

	r.file.mx.RLock()
	defer r.file.mx.RUnlock()
	r.mx.Lock()
	defer r.mx.Unlock()

	ret, err := r.osFile.Seek(offset, whence)
	r.buffer.Reset(r.osFile)

	return ret, err
}

func (r *FileReader) Close() error {
	return r.osFile.Close()
}

func (r *FileReader) ReadAt(buffer []byte, off int64) (n int, err error) {
	if r.file.closed {
		return 0, os.ErrClosed
	}

	r.file.mx.RLock()
	defer r.file.mx.RUnlock()

	_, err = r.Seek(off, 0)
	r.mx.Lock()
	defer r.mx.Unlock()

	if err != nil {
		return 0, err
	}

	return r.buffer.Read(buffer)
}

func (r *FileReader) Read(buffer []byte) (n int, err error) {
	if r.file.closed {
		return 0, os.ErrClosed
	}

	r.file.mx.RLock()
	defer r.file.mx.RUnlock()
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.buffer.Read(buffer)
}
