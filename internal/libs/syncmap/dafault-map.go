package syncmap

import (
	"iter"
	"maps"
	"os"
	"sync"
)

// DefaultMap currently uses default map as an implementation (synced by sync.RWMutex). Could be changed to sync.Map any time.
type DefaultMap[K comparable, V any] struct {
	m  map[K]V
	mx *sync.RWMutex
}

func NewMap[K comparable, V any]() *DefaultMap[K, V] {
	return &DefaultMap[K, V]{
		m:  map[K]V{},
		mx: &sync.RWMutex{},
	}
}

func (m *DefaultMap[K, V]) Get(key K) (V, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	v, ok := m.m[key]

	return v, ok
}

func (m *DefaultMap[K, V]) Set(key K, value V) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.m[key] = value

	return nil
}

func (m *DefaultMap[K, V]) Delete(key K) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	if _, ok := m.m[key]; ok {
		return os.ErrExist
	}

	delete(m.m, key)

	return nil
}

// All creates a copy of the underlying map and returns an iterator over it. Use it wisely.
func (m *DefaultMap[K, V]) All() iter.Seq2[K, V] {
	m.mx.RLock()
	defer m.mx.RUnlock()

	mp := make(map[K]V)
	maps.Copy(mp, m.m)

	return maps.All(mp)
}
