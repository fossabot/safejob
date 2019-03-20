package safejob

import (
	"context"
	"errors"
	"sync"
)

var ErrClosed = errors.New("unsuccessful attempt to perform work after stop")

type safejob struct {
	closer func()
	tokens chan func()
}

func New(closer func()) *safejob {
	return &safejob{
		closer: closer,
		tokens: make(chan func()),
	}
}

func (s *safejob) Do(f func() error) error {
	token := <-s.tokens
	if token == nil {
		return ErrClosed
	}
	defer token()
	return f()
}

func (s *safejob) Run(ctx context.Context) {
	wg := new(sync.WaitGroup)
	wg.Add(1)
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
		case s.tokens <- wg.Done:
			wg.Add(1)
		}
	}
}
