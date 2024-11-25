package connector

import (
	"github.com/StratuStore/file-storage/internal/libs/syncmap"
	"github.com/google/uuid"
	"io"
	"os"
	"time"
)

// Connector saves an opened disposable objects. V should be a pointer or an interface
type Connector[V io.Closer] struct {
	m syncmap.Map[uuid.UUID, Connection[V]]
}

func NewConnector[V io.Closer]() *Connector[V] {
	return &Connector[V]{m: syncmap.NewMap[uuid.UUID, Connection[V]]()}
}

func (c *Connector[V]) OpenConnection(value V) (uuid.UUID, error) {
	id := uuid.New()

	err := c.m.Set(id, Connection[V]{
		ID:           id,
		ActivityTime: time.Now(),
		Value:        value,
	})

	return id, err
}

func (c *Connector[V]) Connection(id uuid.UUID) (V, error) {
	var value V

	connection, ok := c.m.Get(id)
	if !ok {
		return value, os.ErrNotExist
	}
	value = connection.Value

	err := c.m.Set(id, Connection[V]{
		ID:           id,
		ActivityTime: time.Now(),
		Value:        value,
	})

	return value, err
}

func StartDisposalRoutine(d time.Duration)

type Connection[V any] struct {
	ID           uuid.UUID
	ActivityTime time.Time
	Value        V
}
