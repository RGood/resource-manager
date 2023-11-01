package resource

import (
	"errors"
)

// Pool is a resource pool that instantiates a fixed number of elements that can be retrieved later
type Pool[T any] struct {
	producer  func() T
	resources chan T
	done      chan struct{}
}

var (
	// ErrDisposed represents the return value is the Pool is already disposed
	ErrDisposed = errors.New("resource pool disposed")

	// ErrPoolDrained is returned if you call TryClaim on an empty pool
	ErrPoolDrained = errors.New("resource pool empty")
)

// New creates and returns a new Pool struct
func New[T any](maxCount int, producer func() T) *Pool[T] {
	p := &Pool[T]{
		producer:  producer,
		resources: make(chan T, maxCount),
		done:      make(chan struct{}),
	}

	for i := 0; i < maxCount; i++ {
		go p.addResource()
	}

	return p
}

func (p *Pool[T]) addResource() {
	r := p.producer()

	select {
	case <-p.done:
		return
	default:
	}

	select {
	case <-p.done:
		return
	case p.resources <- r:
		return
	}
}

func (p *Pool[T]) Claim() (T, error) {
	r, ok := <-p.resources
	if !ok {
		var t T
		return t, ErrDisposed
	}

	go p.addResource()
	return r, nil
}

func (p *Pool[T]) TryClaim() (T, error) {
	select {
	case r, ok := <-p.resources:
		if ok {
			go p.addResource()
			return r, nil
		}
		var t T
		return t, ErrDisposed
	default:
		var t T
		return t, ErrPoolDrained
	}
}

func (p *Pool[T]) Dispose() error {
	close(p.resources)

	// Empty the pool
	for _, ok := <-p.resources; ok; _, ok = <-p.resources {
	}

	return nil
}
