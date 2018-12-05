package gopool

import (
	"errors"
	"strings"
)

// ChanPool implements the Pool interface, maintaining a pool of resources.
type ChanPool struct {
	ch  chan interface{}
	cfg config
}

// New creates a new Pool. The factory method used to create new items for the Pool must be
// specified using the gopool.Factory method. Optionally, the pool size and a reset function can be
// specified.
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
//		bp, err := gopool.New(gopool.Size(poolSize), gopool.Factory(makeBuffer), gopool.Reset(resetBuffer))
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
		ch:  make(chan interface{}, pc.size),
		cfg: *pc,
	}
	for i := 0; i < pool.cfg.size; i++ {
		item, err := pool.cfg.factory()
		if err != nil {
			return nil, err
		}
		pool.ch <- item
	}
	return pool, nil
}

// Get acquires and returns an item from the pool of resources. When the Pool's Wait is set, Get
// blocks while there are no items in the pool. Otherwise, Get creates amd returns a new instance.
func (p *ChanPool) Get() interface{} {
	for {
		select {
		case item := <-p.ch:
			return item
		default:
			if !p.cfg.wait {
				if item, err := p.cfg.factory(); err == nil {
					return item
				}
			}
		}
	}
}

// Put will release a resource back to the pool. When the Pool's Wait is set, Put blocks if pool
// already full. Otherwise, Put discards the resource after optionally calling the Close method with
// the resource. If the Pool was initialized with a Reset function, it will be invoked with the
// resource as its sole argument, prior to the resource being added back to the pool. If Put is
// called when adding the resource to the pool _would_ result in having more elements in the pool
// than the pool size, the resource is effectively dropped on the floor after calling any optional
// Reset and Close methods on the resource.
func (p *ChanPool) Put(item interface{}) {
	if p.cfg.reset != nil {
		p.cfg.reset(item)
	}
	for {
		select {
		case p.ch <- item:
			return
		default:
			if !p.cfg.wait {
				if p.cfg.close != nil {
					_ = p.cfg.close(item) // ignore close errors
				}
				return
			}
		}
	}
}

// Close is called when the Pool is no longer needed, and the resources in the Pool ought to be
// released.  If a Pool has a close function, it will be invoked one time for each resource, with
// that resource as its sole argument.
func (p *ChanPool) Close() error {
	var errs []error
	for {
		select {
		case item := <-p.ch:
			if p.cfg.close != nil {
				if err := p.cfg.close(item); err != nil {
					errs = append(errs, err)
				}
			}
		default:
			close(p.ch)
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
