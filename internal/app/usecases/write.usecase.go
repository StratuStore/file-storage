package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/url"
)

const fsmPath = "/file/"

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

	if size <= 0 {
		return fmt.Errorf("request is empty: %w", u.handleWriteError(context.Background(), file.Host, file.File.ID()))
	}

	n, err := io.CopyN(writer, &contextReader{reader, ctx}, size)
	if err != nil || n != size {
		return fmt.Errorf("unable to write full file: %w", errors.Join(err, u.handleWriteError(context.Background(), file.Host, file.File.ID())))
	}

	return nil
}

func (u *UseCases) handleWriteError(ctx context.Context, host string, fileID uuid.UUID) error {
	link, err := url.JoinPath(host, fsmPath, fileID.String())
	if err != nil {
		return fmt.Errorf("failed to join path during handling /write error: %w", err)
	}

	_, err = u.client.R().
		SetContext(ctx).
		SetAuthScheme("Bearer").
		SetAuthToken(u.serviceToken).
		Delete(link)
	if err != nil {
		return fmt.Errorf("failed to delete file during handling /write error: %w", err)
	}

	if err := u.DeleteFile(ctx, fileID); err != nil {
		return fmt.Errorf("failed to delete file after write error: %w", err)
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
