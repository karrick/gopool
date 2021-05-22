package gopool

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type minipool struct {
	pc         config
	items      []interface{}
	getc, putc *sync.Cond // NOTE: both uses same underlying mutex
}

func NewMiniPool(setters ...Configurator) (*minipool, error) {
	pc := config{
		maxsize: DefaultSize,
	}
	for _, setter := range setters {
		if err := setter(&pc); err != nil {
			return nil, err
		}
	}
	if pc.maxsize <= 0 {
		return nil, fmt.Errorf("cannot create with size less than or equal to zero: %d", pc.maxsize)
	}
	if pc.minsize > 0 {
		if pc.minsize > pc.maxsize {
			return nil, fmt.Errorf("cannot create when minimum size is greater than size: %d > %d", pc.minsize, pc.maxsize)
		}
		if pc.factory == nil {
			return nil, errors.New("cannot create with non zero minimum size and without specifying a factory method")
		}
	}

	mutex := new(sync.Mutex)

	pool := &minipool{
		items: make([]interface{}, 0, pc.maxsize),
		getc:  &sync.Cond{L: mutex},
		putc:  &sync.Cond{L: mutex},
		pc:    pc,
	}

	for i := 0; i < pc.minsize; i++ {
		item, err := pool.pc.factory()
		if err != nil {
			if pool.pc.close != nil {
				_ = pool.Close() // ignore error; want user to get error from factory call
			}
			return nil, err
		}
		pool.items = append(pool.items, item)
	}

	return pool, nil
}

// Close releases resources consumed by the pool by invoking the Close function
// with each item in the pool. It also unblocks any outstanding Get() or Put()
// operations that are currently blocked. Any outstanding Get() or Put()
// invocations will panic, so be sure that go routines are not blocked in those
// calls before calling this method.
func (pool *minipool) Close() error {
	pool.getc.L.Lock()

	var errs []error
	var err error
	var i int

	if pool.pc.close != nil {
		for _, item := range pool.items {
			if err := pool.pc.close(item); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// Prevent use of pool after Close.
	pool.items = nil

	// Unlock all calls to Get() and Put() to prevent go routine leak.
	pool.getc.L.Unlock()
	pool.getc.Broadcast()
	pool.putc.Broadcast()

	if len(errs) == 0 {
		return nil
	}

	messages := make([]string, len(errs))
	for i, err = range errs {
		messages[i] = err.Error()
	}
	return errors.New(strings.Join(messages, ", "))
}

func (pool *minipool) Get() (item interface{}) {
	pool.getc.L.Lock()
	for {
		if len(pool.items) > 0 {
			// fmt.Fprintf(os.Stderr, "Mini Get() %d of %d\n", len(pool.items), pool.pc.maxsize)
			item, pool.items = pool.items[0], pool.items[1:]

			pool.getc.L.Unlock()
			pool.putc.Signal()

			return item
		}
		if pool.items == nil {
			panic("Get() after Close()")
		}
		// fmt.Fprintf(os.Stderr, "Mini Get() %d of %d wait\n", len(pool.items), pool.pc.maxsize)
		pool.getc.Wait()
	}
}

func (pool *minipool) Put(item interface{}) {
	if pool.pc.reset != nil {
		pool.pc.reset(item)
	}

	pool.putc.L.Lock()
	for {
		if len(pool.items) < pool.pc.maxsize {
			// fmt.Fprintf(os.Stderr, "Mini Put() %d of %d\n", len(pool.items), pool.pc.maxsize)
			pool.items = append(pool.items, item)

			pool.putc.L.Unlock()
			pool.getc.Signal()

			return
		}
		if pool.items == nil {
			panic("Put() after Close()")
		}
		// fmt.Fprintf(os.Stderr, "Mini Put() %d of %d wait\n", len(pool.items), pool.pc.maxsize)
		pool.putc.Wait()
	}
}
