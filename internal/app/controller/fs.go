package controller

import (
	"errors"
	"github.com/StratuStore/file-storage/internal/app/fileio"
	"os"
)

type FileSystem interface {
	OpenForReading(name string) (fileio.FsFile, error)
	Stat(name string) (os.FileInfo, error)
	FSDelete(name string) error
	CreateOrOpenForWriting(name string) (fileio.FsFile, error)
	ListDir(path string) (files map[string]int64, err error)
}

// osFs implements fileSystem using the local drive.
type osFs struct{}

func (osFs) OpenForReading(name string) (fileio.FsFile, error) { return os.Open(name) }

func (osFs) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

func (osFs) FSDelete(name string) error {
	return os.Remove(name)
}

func (osFs) CreateOrOpenForWriting(name string) (fileio.FsFile, error) {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0o666)

	return file, err
}

func (osFs) ListDir(path string) (files map[string]int64, err error) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range dir {
		if file.IsDir() {
			continue
		}

		stat, err2 := file.Info()
		if err2 != nil {
			errors.Join(err, err2)
			continue
		}

		files[stat.Name()] = stat.Size()
	}

	return files, err
}
