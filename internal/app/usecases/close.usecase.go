package usecases

import (
	"context"
	"github.com/google/uuid"
)

// Close supposed to be a request from user directly
func (u *UseCases) Close(ctx context.Context, connectionID uuid.UUID) (err error) {
	reader, err := u.ReadersConnector.Connection(connectionID)
	if err != nil {
		return err
	}

	err = reader.Close()
	if err != nil {
		return err
	}

	return nil
}
