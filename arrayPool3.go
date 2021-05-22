package gopool

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// ArrayPool3 implements the Pool interface, maintaining a pool of resources.
type ArrayPool3 struct {
	pc         config
	items      []interface{}
	getc, putc sync.Cond // Both use same underlying mutex

	// index points to where a Put puts. Get will block when index is 0, Put blocks
	// when index is size.
	//
	// Empty:
	//    0 1 2 3 4 5 6 7
	//
	//    ^
	// Full:
	//    0 1 2 3 4 5 6 7
	//    x x x x x x x x
	//                    ^
	index int
}

// NewArrayPool3 creates a new Pool. The factory method used to create new items
// for the Pool must be specified using the gopool.Factory method. Optionally,
// the pool size and a reset function can be specified.
//
//	package main
//
//	import (
//		"bytes"
//		"errors"
//		"fmt"
//		"log"
//		"math/rand"
//		"sync"
//
//		"github.com/karrick/gopool"
//	)
//
//	const (
//		bufSize  = 64 * 1024
//		poolSize = 25
//	)
//
//	func main() {
//		const iterationCount = 1000
//		const parallelCount = 100
//
//		makeBuffer := func() (interface{}, error) {
//			return bytes.NewBuffer(make([]byte, 0, bufSize)), nil
//		}
//
//		resetBuffer := func(item interface{}) {
//			item.(*bytes.Buffer).Reset()
//		}
//
//		bp, err := gopool.NewArrayPool3(gopool.Size(poolSize), gopool.Factory(makeBuffer), gopool.Reset(resetBuffer))
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		var wg sync.WaitGroup
//		wg.Add(parallelCount)
//
//		for i := 0; i < parallelCount; i++ {
//			go func() {
//				defer wg.Done()
//
//				for j := 0; j < iterationCount; j++ {
//					if err := grabBufferAndUseIt(bp); err != nil {
//						fmt.Println(err)
//						return
//					}
//				}
//			}()
//		}
//		wg.Wait()
//	}
//
//	func grabBufferAndUseIt(pool gopool.Pool) error {
//		// WARNING: Must ensure resource returns to pool otherwise gopool will deadlock once all
//		// resources used.
//		bb := pool.Get().(*bytes.Buffer)
//		defer pool.Put(bb) // IMPORTANT: defer here to ensure invoked even when subsequent code bails
//
//		for k := 0; k < bufSize; k++ {
//			if rand.Intn(100000000) == 1 {
//				return errors.New("random error to illustrate need to return resource to pool")
//			}
//			bb.WriteByte(byte(k % 256))
//		}
//		return nil
//	}
func NewArrayPool3(setters ...Configurator) (Pool, error) {
	var pc config
	var err error
	for _, setter := range setters {
		if err = setter(&pc); err != nil {
			return nil, err
		}
	}

	if pc.maxsize == 0 {
		if pc.minsize == 0 {
			return nil, errors.New("cannot create without specifying maximum or minimum size")
		}
		// Allow specifying minsize without maxsize. Is this reasonable?
		pc.maxsize = pc.minsize
	} else if pc.minsize > pc.maxsize {
		return nil, fmt.Errorf("cannot create when minimum size is greater than maximum size: %d > %d", pc.minsize, pc.maxsize)
	}

	mutex := new(sync.Mutex)

	pool := &ArrayPool3{
		getc:  sync.Cond{L: mutex},
		putc:  sync.Cond{L: mutex},
		items: make([]interface{}, pc.maxsize),
		pc:    pc,
	}

	if pc.minsize > 0 {
		if pc.factory == nil {
			return nil, errors.New("cannot create with non zero minimum size and without specifying a factory method")
		}
		/* pool.index = 0 */ // pool.already already has zero value from initialization
		for ; pool.index < pool.pc.minsize; pool.index++ {
			pool.items[pool.index], err = pool.pc.factory()
			if err != nil {
				if pool.pc.close != nil {
					_ = pool.Close() // ignore close error; rather want user to get error from factory failure above
				}
				return nil, err
			}
		}
	}

	return pool, nil
}

// Close is called when the Pool is no longer needed, and the resources in the
// Pool ought to be released.  If a Pool has a close function, it will be
// invoked one time for each resource, with that resource as its sole argument.
func (pool *ArrayPool3) Close() error {
	pool.getc.L.Lock()

	var errs []error
	var err error
	var i int

	if pool.pc.close != nil {
		for i = 0; i < pool.index; i++ {
			if err = pool.pc.close(pool.items[i]); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// Prevent use of pool after Close.
	pool.items = nil
	pool.index = 0

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

// Get returns an item from the free list, after possibly blocking if no items
// are available.
func (pool *ArrayPool3) Get() interface{} {
	pool.getc.L.Lock()
	for {
		if pool.index > 0 {
			pool.index--
			item := pool.items[pool.index]

			pool.getc.L.Unlock()
			pool.putc.Signal()
			return item
		}
		if pool.items == nil {
			panic("Get() after Close()")
		}
		pool.getc.Wait()
	}
}

// Put will release a resource back to the pool. Put blocks if pool already
// full. If the Pool was initialized with a Reset function, it will be invoked
// with the resource as its sole argument, prior to the resource being added
// back to the pool. If Put is called when adding the resource to the pool
// _would_ result in having more elements in the pool than the pool size, the
// resource is effectively dropped on the floor after calling any optional Reset
// and Close methods on the resource.
func (pool *ArrayPool3) Put(item interface{}) {
	if pool.pc.reset != nil {
		pool.pc.reset(item)
	}

	pool.putc.L.Lock()
	for {
		if pool.index < pool.pc.maxsize {
			pool.items[pool.index] = item
			pool.index++

			pool.putc.L.Unlock()
			pool.getc.Signal()

			return
		}
		if pool.items == nil {
			panic("Put() after Close()")
		}
		pool.putc.Wait()
	}
}
