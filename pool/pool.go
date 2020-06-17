package pool

import (
	"errors"
	"io"
	"log"
	"sync"
)

type Pool struct {
	m       sync.Mutex
	res     chan io.Closer
	factory func() (io.Closer, error)
	closed  bool
}

func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("size is too small")
	}

	return &Pool{
		factory: fn,
		res:     make(chan io.Closer, size),
	}, nil
}

func (p *Pool) Acquire() (io.Closer, error) {
	select {
	case r, ok := <-p.res:
		log.Println("Acquire: 共享")
		if !ok {
			return nil, errors.New("size is too small")
		}
		return r, nil
	default:
		log.Println("Acquire: new source")
		return p.factory()
	}
}

func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}

	p.closed = true
	close(p.res)

	for r := range p.res {
		r.Close()
	}
}

func (p *Pool) Release(r io.Closer) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		r.Close()
		return
	}

	select {
	case p.res <- r:
		log.Println("source release success")
	default:
		log.Println("the pool is full")
		r.Close()
	}
}
