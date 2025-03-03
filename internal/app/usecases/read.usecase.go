package usecases

import (
	"context"
	"errors"
	"github.com/google/uuid"
)

// Read supposed to be a request from user directly
func (u *UseCases) Read(ctx context.Context, connectionID uuid.UUID) (reader Reader, err error) {
	if connection, err := u.ReadersConnector.Connection(connectionID); err != nil {
		return nil, err
	} else if connection.Closed() {
		return nil, errors.New("connection closed")
	} else {
		return connection, nil
	}
}
