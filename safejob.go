package safejob

import (
	"context"
	"errors"
	"sync"
)

// ErrClosed returns when safejob is closed
var ErrClosed = errors.New("unsuccessful attempt to perform work after stop")

type (
	safejob struct {
		closer func()
		tokens chan *token
	}

	token struct {
		context context.Context
		done    func()
	}
)

// New create new instance safejob
func New(closer func()) *safejob {
	return &safejob{
		closer: closer,
		tokens: make(chan *token),
	}
}

// Do safe job
func (s *safejob) Do(f func() error) error {
	token := <-s.tokens
	if token == nil {
		return ErrClosed
	}
	defer token.done()
	return f()
}

// Do safe job with context
func (s *safejob) DoWithContext(f func(context.Context) error) error {
	token := <-s.tokens
	if token == nil {
		return ErrClosed
	}
	defer token.done()
	return f(token.context)
}

// Run worker
func (s *safejob) Run(ctx context.Context) {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	t := &token{
		context: ctx,
		done:    wg.Done,
	}
	for {
		select {
		case <-ctx.Done():
			close(s.tokens)
			wg.Done()
			wg.Wait()
			if s.closer != nil {
				s.closer()
			}
			return
		case s.tokens <- t:
			wg.Add(1)
		}
	}
}
