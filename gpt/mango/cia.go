package mango

import (
	"log"
	"time"
)

type (
	// CIA is like a Guava loading cache
	CIA[K comparable, V any] interface {
		ComputeIfAbsent(in K) Future[V]
		ReloadValue(in K) Future[V]
		Handle(error)
	}

	loadingLRU[K comparable, V any] struct {
		fp  futureProvider[K, V]
		lru Cache[K, V]

		fail  func(error)
		stats CacheLoadStats
	}
)

func (cia *loadingLRU[K, V]) Handle(err error) {
	if err != nil {
		cia.fail(err)
	}
}

func (cia *loadingLRU[K, V]) ComputeIfAbsent(in K) Future[V] {
	if out, ok := cia.Get(in); ok {
		return &tolerant[V]{nil, out}
	}
	out := &deferred[V]{}
	pipe := make(chan *V)
	start := time.Now()
	go func(hole chan<- *V) {
		defer close(hole)
		cia.Handle(cia.fp(hole, in))
	}(pipe)
	go func(hole <-chan *V) {
		pv := <-hole
		end := time.Now()
		nanos := uint64(end.Sub(start).Nanoseconds())
		cia.stats.InNanos.Count(nanos)
		if pv != nil {
			cia.stats.Success.Count(1)
			out.cached = &tolerant[V]{nil, pv}
			cia.Put(in, *pv)
		} else {
			cia.stats.Failure.Count(1)
			out.cached = &tolerant[V]{nil, nil}
		}
	}(pipe)
	return out
}

func (cia *loadingLRU[K, V]) ReloadValue(in K) Future[V] {
	cia.Evict(in)
	return cia.ComputeIfAbsent(in)
}

func NewLoadingCache[K comparable, V any](my CacheOptions, f func(K) V) CIA[K, V] {
	return newCIA(my, func(out chan<- *V, in K) error {
		v := f(in)
		out <- &v
		return nil
	})
}

func newCIA[K comparable, V any](my CacheOptions, f futureProvider[K, V]) CIA[K, V] {
	cache := NewLRU[K, V](my)
	return &loadingLRU[K, V]{
		fp:  f,
		lru: cache,

		// TODO: work out how failures and panic in loader works
		fail: func(err error) {
			log.Panicln(err)
		},
	}
}

func (cia *loadingLRU[K, V]) Clear() {
	cia.lru.Clear()
}

func (cia *loadingLRU[K, V]) Evict(in K) (*V, bool) {
	return cia.lru.Evict(in)
}

func (cia *loadingLRU[K, V]) Stats() CacheStats {
	return cia.lru.Stats()
}

func (cia *loadingLRU[K, V]) Get(in K) (*V, bool) {
	return cia.lru.Get(in)
}

func (cia *loadingLRU[K, V]) Put(key K, value V) {
	cia.lru.Put(key, value)
}

func (cia *loadingLRU[K, V]) Len() int {
	return cia.lru.Len()
}
