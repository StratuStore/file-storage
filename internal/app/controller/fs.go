package controller

import (
	"github.com/StratuStore/file-storage/internal/app/fileio"
	"os"
)

type FileSystem interface {
	ReadOpen(name string) (fileio.FsFile, error)
	Stat(name string) (os.FileInfo, error)
	FSDelete(name string) error
	CreateOrWriteOpen(name string) (fileio.FsFile, error)
}

// osFs implements fileSystem using the local drive.
type osFs struct{}

func (osFs) ReadOpen(name string) (fileio.FsFile, error) { return os.Open(name) }

func (osFs) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

func (osFs) FSDelete(name string) error {
	return os.Remove(name)
}

func (osFs) CreateOrWriteOpen(name string) (fileio.FsFile, error) {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0o666)

	return file, err
}
