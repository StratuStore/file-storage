package fileio

import (
	"github.com/google/uuid"
	"io"
	"os"
)

type FileSystem interface {
	ReadOpen(name string) (FsFile, error)
	Stat(name string) (os.FileInfo, error)
	FSDelete(name string) error
	CreateOrWriteOpen(name string) (FsFile, error)
}

type FsFile interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	Stat() (os.FileInfo, error)
}

type StorageController interface {
	FileSystem
	AllocateStorage(size int) error
	ReleaseStorage(size int) error
	AllocateAll() (n int, err error)
	AddFile(id uuid.UUID) (File, error)
	DeleteFile(id uuid.UUID) error
	File(id uuid.UUID) (File, error)
}
