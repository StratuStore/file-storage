package fileio

type StorageController interface {
	AllocateStorage(size int) error
	ReleaseStorage(size int) error
	AllocateAll() (n int, err error)
}
