package usecases

import (
	"context"
	"github.com/google/uuid"
	"io"
)

// Write supposed to be a request from user directly
func (u *UseCases) Write(ctx context.Context, connectionID uuid.UUID, reader io.Reader, size int64) (err error) {
	file, err := u.FilesConnector.Connection(connectionID)
	if err != nil {
		return err
	}

	writer, err := file.Writer()
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.CopyN(writer, &contextReader{reader, ctx}, size)
	if err != nil {
		return err
	}

	return nil
}

type contextReader struct {
	io.Reader
	ctx context.Context
}

func (c *contextReader) Read(p []byte) (n int, err error) {
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	default:
		return c.Reader.Read(p)
	}
}
