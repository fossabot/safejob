package safejob

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCloseChannel(t *testing.T) {
	ch := make(chan struct{}, 1000)
	sj := New(func() {
		close(ch)
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sj.Run(ctx)

	// prepare jobs
	var errCount uint64
	jobCount := 1000
	steps := 1000
	start := make(chan struct{})
	wg := new(sync.WaitGroup)
	wg.Add(jobCount)
	for n := 0; n < jobCount; n++ {
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < steps; i++ {
				if err := sj.Do(func() error {
					ch <- struct{}{}
					return nil
				}); err != nil {
					atomic.AddUint64(&errCount, 1)
				}
			}
		}()
	}
	close(start)

	// random terminate
	n := 0
	rand.Seed(time.Now().Unix())
	stop := rand.Intn(jobCount * steps / 2)
	for range ch {
		n++
		if n == stop {
			cancel()
		}
	}
	wg.Wait()
	assert.EqualValues(t, jobCount*steps, n+int(errCount))
}

func TestWithContext(t *testing.T) {
	ch := make(chan struct{}, 1000)
	sj := New(func() {
		close(ch)
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sj.Run(ctx)

	// prepare jobs
	var errCount uint64
	jobCount := 1000
	steps := 1000
	start := make(chan struct{})
	wg := new(sync.WaitGroup)
	wg.Add(jobCount)
	for n := 0; n < jobCount; n++ {
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < steps; i++ {
				if err := sj.DoWithContext(func(c context.Context) error {
					select {
					case <-c.Done():
						return c.Err()
					default:
					}
					ch <- struct{}{}
					<-c.Done()
					return nil
				}); err != nil {
					atomic.AddUint64(&errCount, 1)
				}
			}
		}()
	}
	close(start)

	// random terminate
	n := 0
	rand.Seed(time.Now().Unix())
	stop := rand.Intn(jobCount)
	if stop == 0 {
		stop = 1
	}
	for range ch {
		n++
		if n == stop {
			cancel()
		}
	}
	wg.Wait()
	assert.EqualValues(t, jobCount*steps, n+int(errCount))
	assert.True(t, n <= jobCount+1)
}
