package mango

import (
	"fmt"
	"sync"
)

type (
	CacheOptions struct {
		MaxSize uint64
	}

	Cache[K comparable, V any] interface {
		Get(key K) (*V, bool)
		Put(key K, value V)

		Evict(key K) (*V, bool)
		Clear()
		Stats() CacheStats

		Len() int
	}

	lru[K comparable, V any] struct {
		lock sync.RWMutex

		maxSize int
		nodes   map[K]*lruNode[K, V]
		stats   CacheStats

		head, tail lruNode[K, V]
	}

	lruNode[K comparable, V any] struct {
		key   K
		value V

		lhs, rhs *lruNode[K, V]
	}
)

func NewLRU[K comparable, V any](options CacheOptions) Cache[K, V] {
	limit := 1000
	if m := options.MaxSize; m > 0 {
		limit = int(m)
	}
	cache := &lru[K, V]{
		maxSize: limit,
	}
	cache.Clear()
	return cache
}

func (cache *lru[K, V]) Len() int {
	return len(cache.nodes)
}

func newNode[K comparable, V any](key K, value V) *lruNode[K, V] {
	return &lruNode[K, V]{
		key:   key,
		value: value,
	}
}

func (n *lruNode[K, V]) String() string {
	return fmt.Sprintf("{%v:%v}", n.key, n.value)
}
