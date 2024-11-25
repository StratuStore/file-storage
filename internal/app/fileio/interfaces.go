package fileio

import "io"

type Writer interface {
	io.Writer
	io.WriterAt
	io.Closer
}

type Reader interface {
	io.ReadSeekCloser
	io.ReaderAt
}
