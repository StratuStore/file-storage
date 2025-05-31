package connector

import (
	"errors"
	"github.com/StratuStore/file-storage/internal/libs/syncmap"
	"github.com/google/uuid"
	"io"
	"os"
	"time"
)

type Closeder interface {
	Closed() bool
}

// Connector saves opened disposable objects.
type Connector[V Closeder] struct {
	m syncmap.Map[uuid.UUID, Connection[V]]
}

func NewConnector[V Closeder]() *Connector[V] {
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

func (c *Connector[V]) StartDisposalRoutine(sleep time.Duration, timeout time.Duration) {
	go func() {
		for {
			time.Sleep(sleep)
			c.dispose(timeout)
		}
	}()
}

func (c *Connector[V]) dispose(timeout time.Duration) (err error) {
	for id, connection := range c.m.All() {
		if connection.ActivityTime.Add(timeout).Before(time.Now()) || connection.Value.Closed() {
			if closer, ok := any(connection.Value).(io.Closer); ok && !connection.Value.Closed() {
				closer.Close()
			}

			err = errors.Join(err, c.m.Delete(id))
		}
	}

	return err
}

type Connection[V any] struct {
	ID           uuid.UUID
	ActivityTime time.Time
	Value        V
}
