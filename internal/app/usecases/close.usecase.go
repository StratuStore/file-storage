package usecases

import (
	"context"
	"github.com/google/uuid"
)

// Close supposed to be a request from user directly
func (u *UseCases) Close(ctx context.Context, connectionID uuid.UUID) (err error) {
	file, err := u.FilesConnector.Connection(connectionID)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}
