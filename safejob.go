package safejob

import (
	"context"
	"errors"
	"io"
	"sync"
)

var ErrClosed = errors.New("unsuccessful attempt to perform work after stop")

type safejob struct {
	closer io.Closer
	tokens chan func()
}

func New(c io.Closer) *safejob {
	return &safejob{
		closer: c,
		tokens: make(chan func()),
	}
}

func (s *safejob) Do(f func()) error {
	token := <-s.tokens
	if token == nil {
		return ErrClosed
	}
	defer token()
	f()
	return nil
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
			s.closer.Close()
			return
		case s.tokens <- wg.Done:
			wg.Add(1)
		}
	}
}
