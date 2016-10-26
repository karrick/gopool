package gopool

import (
	"errors"
	"strings"
)

// ChanPool implements the Pool interface, maintaining a pool of resources.
type ChanPool struct {
	ch chan interface{}
	pc config
}

// New creates a new Pool. The factory method used to create new items for the Pool must be
// specified using the gopool.Factory method. Optionally, the pool size and a reset function can be
// specified.
//
//	package main
//
//	import (
//		"bytes"
//		"log"
//		"sync"
//
//		"github.com/karrick/gopool"
//	)
//
//	func main() {
//		makeBuffer := func() (interface{}, error) {
//			return new(bytes.Buffer), nil
//		}
//
//		resetBuffer := func(item interface{}) {
//			item.(*bytes.Buffer).Reset()
//		}
//
//		bp, err := gopool.New(gopool.Size(25), gopool.Factory(makeBuffer), gopool.Reset(resetBuffer))
//		if err != nil {
//			log.Fatal(err)
//		}
//		var wg sync.WaitGroup
//		wg.Add(100)
//		for i := 0; i < 100; i++ {
//			go func() {
//				for j := 0; j < 1000; j++ {
//					bb := bp.Get().(*bytes.Buffer)
//					for k := 0; k < 4096; k++ {
//						bb.WriteByte(byte(k % 256))
//					}
//					bp.Put(bb)
//				}
//				wg.Done()
//			}()
//		}
//		wg.Wait()
//	}
func New(setters ...Configurator) (Pool, error) {
	pc := &config{
		size: DefaultSize,
	}
	for _, setter := range setters {
		if err := setter(pc); err != nil {
			return nil, err
		}
	}
	if pc.factory == nil {
		return nil, errors.New("ought to specify factory method")
	}
	pool := &ChanPool{
		ch: make(chan interface{}, pc.size),
		pc: *pc,
	}
	for i := 0; i < pool.pc.size; i++ {
		item, err := pool.pc.factory()
		if err != nil {
			return nil, err
		}
		pool.ch <- item
	}
	return pool, nil
}

// Get acquires and returns an item from the pool of resources.
func (pool *ChanPool) Get() interface{} {
	return <-pool.ch
}

// Put will release a resource back to the pool.  If the Pool was initialized with a Reset function,
// it will be invoked with the resource as its sole argument, prior to the resource being added back
// to the pool.
func (pool *ChanPool) Put(item interface{}) {
	if pool.pc.reset != nil {
		pool.pc.reset(item)
	}
	pool.ch <- item
}

// Close is called when the Pool is no longer needed, and the resources in the Pool ought to be
// released.  If a Pool has a close function, it will be invoked one time for each resource, with
// that resource as its sole argument.
func (pool *ChanPool) Close() error {
	var errs []error
	for {
		select {
		case item := <-pool.ch:
			if pool.pc.close != nil {
				if err := pool.pc.close(item); err != nil {
					errs = append(errs, err)
				}
			}
		default:
			if len(errs) == 0 {
				return nil
			}
			var messages []string
			for _, err := range errs {
				messages = append(messages, err.Error())
			}
			return errors.New(strings.Join(messages, ", "))
		}
	}
}
