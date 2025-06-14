package usecases

import (
	"context"
	"github.com/google/uuid"
)

// CreateFile supposed to be a request from FileSystem Manager via Kafka
func (u *UseCases) CreateFile(ctx context.Context, host string, fileID uuid.UUID) (connectionID uuid.UUID, err error) {
	file, err := u.StorageController.AddFile(fileID)
	if err != nil {
		return connectionID, err
	}

	return u.FilesConnector.OpenConnection(&FileWithHost{
		File: file,
		Host: host,
	})
}
