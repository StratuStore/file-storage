package usecases

import (
	"context"
	"github.com/google/uuid"
)

// DeleteFile supposed to be a request from FileSystem Manager via Kafka
func (u *UseCases) DeleteFile(ctx context.Context, fileID uuid.UUID) error {
	return u.StorageController.DeleteFile(fileID)
}
