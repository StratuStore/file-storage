package usecases

import (
	"context"
	"github.com/google/uuid"
)

// OpenFile supposed to be a request from FileSystem Manager via Kafka
func (u *UseCases) OpenFile(ctx context.Context, fileID uuid.UUID) (connectionID uuid.UUID, err error) {
	file, err := u.StorageController.File(fileID)
	if err != nil {
		return connectionID, err
	}

	bufferSize := max(min(u.MaxBufferSize, file.Size()/10), u.MinBufferSize)

	reader, err := file.Reader(bufferSize)
	if err != nil {
		return connectionID, err
	}

	return u.ReadersConnector.OpenConnection(reader)
}
