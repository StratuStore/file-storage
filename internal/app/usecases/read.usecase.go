package usecases

import (
	"context"
	"github.com/google/uuid"
)

// Read supposed to be a request from user directly
func (u *UseCases) Read(ctx context.Context, connectionID uuid.UUID) (reader Reader, err error) {
	return u.ReadersConnector.Connection(connectionID)
}
