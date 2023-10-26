package resource

import "errors"

// Pool is a resource pool that instantiates a fixed number of elements that can be retrieved later
type Pool[T any] struct {
	producer  func() (T, error)
	resources chan T
	done      chan struct{}
}

// New creates and returns a new Pool struct
func New[T any](maxCount int, producer func() (T, error)) *Pool[T] {
	p := &Pool[T]{
		producer:  producer,
		resources: make(chan T, maxCount),
		done:      make(chan struct{}),
	}

	go func() {
		for {
			select {
			case <-p.done:
				return
			default:
			}

			r, err := p.producer()
			if err != nil {
				// do something with the error(?)
			}

			select {
			case <-p.done:
				return
			case p.resources <- r:
			}
		}
	}()

	return p
}

func (p *Pool[T]) Claim() T {
	return <-p.resources
}

func (p *Pool[T]) TryClaim() (T, error) {
	select {
	case r := <-p.resources:
		return r, nil
	default:
		var t T
		return t, errors.New("resource pool empty")
	}
}
