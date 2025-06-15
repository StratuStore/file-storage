package usecases

import (
	"context"
	"errors"
	"github.com/StratuStore/file-storage/internal/app/fileio"
	"github.com/google/uuid"
)

// Read supposed to be a request from user directly
func (u *UseCases) Read(ctx context.Context, connectionID uuid.UUID) (reader Reader, err error) {
	if reader, err = u.ReadersConnector.Connection(connectionID); errors.Is(err, fileio.ErrBusy) {
		return nil, newErrorWithMessage("file is busy")
	} else if err != nil {
		return nil, err
	} else {
		return reader, nil
	}
}
