package gopool_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/karrick/gopool"
)

// EXPERIMENT

type SemaphorePool struct {
	free       []interface{}
	getc, putc *sync.Cond // NOTE: both uses same underlying mutex
	count, max int
}

func NewSemaphorePool(items []interface{}) (*SemaphorePool, error) {
	m := new(sync.Mutex)

	s := &SemaphorePool{
		free:  make([]interface{}, len(items)),
		getc:  &sync.Cond{L: m},
		putc:  &sync.Cond{L: m},
		count: len(items),
		max:   len(items),
	}
	if n := copy(s.free, items); n < len(items) {
		return nil, fmt.Errorf("only copied %d out of %d items", n, len(items))
	}
	return s, nil
}

func (p *SemaphorePool) Close() error {
	p.getc.L.Lock()
	p.free = nil

	// Set count and max such that both Get() and Put() would not block
	p.count = 1
	p.max = 2

	// Unlock all calls to Get() and Put() to prevent go routine leak
	p.getc.L.Unlock()
	p.getc.Broadcast()
	p.putc.Broadcast()

	return nil
}

const inplace = false

func (p *SemaphorePool) Get() (item interface{}) {
	p.getc.L.Lock()
	for p.count == 0 {
		p.getc.Wait()
	}
	if p.free == nil {
		panic("Get() after Close()")
	}

	p.count--
	if inplace {
		item = p.free[p.count]
	} else {
		item, p.free = p.free[0], p.free[1:]
	}

	p.getc.L.Unlock()
	p.putc.Signal()

	return item
}

func (p *SemaphorePool) Put(item interface{}) {
	p.putc.L.Lock()
	for p.count == p.max {
		p.putc.Wait()
	}
	if p.free == nil {
		panic("Put() after Close()")
	}

	if inplace {
		p.free[p.count] = item
	} else {
		p.free = append(p.free, item)
	}
	p.count++

	p.putc.L.Unlock()
	p.getc.Signal()
}

// TESTING

func newSemaphorePool(tb testing.TB, count int) gopool.Pool {
	items := make([]interface{}, count)
	for i := 0; i < count; i++ {
		items[i] = bytes.NewBuffer(make([]byte, defaultBufSize))
	}

	pool, err := NewSemaphorePool(items)
	if err != nil {
		tb.Fatal(err)
	}

	return pool
}

func TestSemaphorePool(t *testing.T) {
	test(t, newSemaphorePool(t, lowCap))
}

func BenchmarkSemaphoreLowConcurrency(b *testing.B) {
	bench(b, newSemaphorePool(b, lowCap), lowConcurrency)
}

func BenchmarkSemaphoreMediumConcurrency(b *testing.B) {
	bench(b, newSemaphorePool(b, medCap), medConcurrency)
}

func BenchmarkSemaphoreHighConcurrency(b *testing.B) {
	bench(b, newSemaphorePool(b, largeCap), highConcurrency)
}
