package mango

import (
	"context"
	"log"
	"sync"

	"golang.org/x/sync/errgroup"
)

type (
	// deferred is resolved to a result (value/error) that is cache-able
	deferred[Value any] struct {
		cached *tolerant[Value]
		source <-chan *Value

		once sync.Once
	}

	// tolerant captured a single result (value or error); implements Future
	tolerant[Value any] struct {
		failure error
		success *Value
	}

	// Future captures the result of a computation that may occur now/later
	Future[Value any] interface {
		Await(context.Context) (*Value, error)
	}

	// futureProvider returns nil if "it" was procured (given some in/out)
	futureProvider[K comparable, It any] func(out chan<- *It, in K) error
)

func awaitMap[K comparable, V any](ctx context.Context, vmap map[K]Future[V]) (map[K]*V, error) {
	keys := make([]K, 0, len(vmap))
	for key := range vmap {
		keys = append(keys, key)
	}
	in := make([]Future[V], len(keys))
	for index, key := range keys {
		in[index] = vmap[key]
	}
	out := make(map[K]*V, len(keys))
	temp, err := awaitAll(ctx, in...)
	for index, key := range keys {
		out[key] = temp[index]
	}
	return out, err
}

func awaitAll[V any](ctx context.Context, vs ...Future[V]) ([]*V, error) {
	// FIXME: something isn't quite right here (see test)
	ops, dtx := errgroup.WithContext(ctx)
	//ops.SetLimit(len(vs)) // use ctx?
	resolved := make([]*V, len(vs))
	errors := make(chan error)
	go func(e chan<- error) {
		defer close(e)
		for i, v := range vs {
			index, value := i, v
			ops.Go(func() error {
				rv, err := value.Await(dtx)
				resolved[index] = rv
				return err
			})
		}
		if err := ops.Wait(); err != nil {
			e <- err
		}
	}(errors)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errors:
		return resolved, err
	}
}

func loading[K comparable, V any](f func(in K) (*V, error)) futureProvider[K, V] {
	return func(out chan<- *V, in K) error {
		log.Println("loading...")
		one, err := f(in)
		log.Println("loaded, may cache:", one, "or fail:", err)
		out <- one
		return err
	}
}

func (v *tolerant[V]) Await(_ context.Context) (*V, error) {
	return v.success, v.failure
}

func (v *deferred[V]) Await(ctx context.Context) (*V, error) {
	v.once.Do(func() {
		select {
		case <-ctx.Done():
			v.cached = &tolerant[V]{ctx.Err(), nil}
		case out := <-v.source:
			v.cached = &tolerant[V]{nil, out}
		}
	})
	return v.cached.Await(ctx)
}

func emits[Value any](fn func(context.Context) (*Value, error)) func(context.Context) Future[Value] {
	return func(ctx context.Context) Future[Value] {
		out := &deferred[Value]{}
		go out.once.Do(func() {
			v, err := fn(ctx)
			out.cached = &tolerant[Value]{err, v}
		})
		return out
	}
}

func failed[Value any](err error) Future[Value] {
	return &tolerant[Value]{err, nil}
}

func loaded[Value any](ok Value) Future[Value] {
	return &tolerant[Value]{nil, &ok}
}
