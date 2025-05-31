package usecases

import (
	"context"
	"github.com/google/uuid"
)

// UpdateFile supposed to be a request from FileSystem Manager via Kafka
func (u *UseCases) UpdateFile(ctx context.Context, fileID uuid.UUID) (connectionID uuid.UUID, err error) {
	file, err := u.StorageController.File(fileID)
	if err != nil {
		return connectionID, err
	}

	return u.FilesConnector.OpenConnection(file)
}
