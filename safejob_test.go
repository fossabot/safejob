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

type testChan struct {
	c chan struct{}
}

func (t *testChan) Close() error {
	close(t.c)
	return nil
}

func TestCloseChannel(t *testing.T) {
	c := &testChan{
		c: make(chan struct{}, 1000),
	}
	sj := New(c)
	ctx, cancel := context.WithCancel(context.Background())
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
				if err := sj.Do(func() {
					c.c <- struct{}{}
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
	for _ = range c.c {
		n++
		if n == stop {
			cancel()
		}
	}
	wg.Wait()
	assert.EqualValues(t, jobCount*steps, n+int(errCount))
}
