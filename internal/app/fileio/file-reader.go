package fileio

import (
	"bufio"
	"os"
	"path"
)

type FileReader struct {
	file   *File
	osFile *os.File
	buffer *bufio.Reader
}

func NewFileReader(f *File) (*FileReader, error) {
	osFile, err := os.Open(path.Join(f.Path, f.ID.String()))
	if err != nil {
		return nil, err
	}

	buffer := bufio.NewReader(osFile)

	return &FileReader{
		file:   f,
		osFile: osFile,
		buffer: buffer,
	}, nil
}

func (r *FileReader) Seek(offset int64, whence int) (int64, error) {
	if r.file.closed {
		return 0, os.ErrClosed
	}

	r.file.mx.RLock()
	defer r.file.mx.RUnlock()

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

	return r.buffer.Read(buffer)
}
