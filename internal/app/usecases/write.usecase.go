package usecases

import (
	"context"
	"github.com/google/uuid"
	"io"
)

// Write supposed to be a request from user directly
func (u *UseCases) Write(ctx context.Context, connectionID uuid.UUID, reader io.Reader) (err error) {
	file, err := u.FilesConnector.Connection(connectionID)

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
}
