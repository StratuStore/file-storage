package syncmap

import "iter"

type Map[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V) error
	Delete(key K) error
	All() iter.Seq2[K, V]
}
