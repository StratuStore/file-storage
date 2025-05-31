package queue

import (
	"errors"
	"github.com/google/uuid"
)

type RequestType int

const (
	CreateType RequestType = iota
	UpdateType
	OpenType
	DeleteType
)

type Request struct {
	ID     uuid.UUID
	Host   string
	Type   RequestType
	FileID uuid.UUID
	Size   uint
}

type Response struct {
	ID           uuid.UUID // must be equal to request.ID
	Host         string
	ConnectionID uuid.UUID
	Err          string
}

func (r *Response) ToReturn() (string, string, error) {
	var err error
	if r.Err != "" {
		err = errors.New(r.Err)
	}

	return r.Host, r.ConnectionID.String(), err
}
