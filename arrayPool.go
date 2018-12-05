package gopool

import (
	"errors"
	"strings"
	"sync"
)

const (
	putBlocks = iota
	getBocks
	neitherBlocks
)

// ArrayPool implements the Pool interface, maintaining a pool of resources.
type ArrayPool struct {
	cond    *sync.Cond
	blocked int // putBlocks | getBlocks | neitherBlocks
	cfg     config
	gi      int // index of next Get
	pi      int // index of next Put
	items   []interface{}
}

// NewArrayPool creates a new Pool. The factory method used to create new items for the Pool must be
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
//		bp, err := gopool.NewArrayPool(gopool.Size(poolSize), gopool.Factory(makeBuffer), gopool.Reset(resetBuffer))
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
func NewArrayPool(setters ...Configurator) (Pool, error) {
	pc := config{
		size: DefaultSize,
	}
	for _, setter := range setters {
		if err := setter(&pc); err != nil {
			return nil, err
		}
	}
	if pc.factory == nil {
		return nil, errors.New("cannot create pool without specifying a factory method")
	}
	pool := &ArrayPool{
		blocked: putBlocks,
		cond:    &sync.Cond{L: &sync.Mutex{}},
		items:   make([]interface{}, pc.size),
		cfg:     pc,
	}
	for i := 0; i < pool.cfg.size; i++ {
		item, err := pool.cfg.factory()
		if err != nil {
			if pool.cfg.close != nil {
				_ = pool.Close() // ignore error; want user to get error from factory call
			}
			return nil, err
		}
		pool.items[i] = item
	}
	return pool, nil
}

// Get acquires and returns an item from the pool of resources. Get blocks while there are no items
// in the pool.
func (p *ArrayPool) Get() interface{} {
	if !p.cfg.wait {
		p.cond.L.RLock()
		if p.blocked == getBocks {
			p.cond.L.RUnlock()
			if item, err := p.cfg.factory(); err == nil {
				return item
			}
		}
	}

	// Get blocks when attempt to Get made at location next Put goes to
	p.cond.L.Lock()
	for p.blocked == getBocks {
		p.cond.Wait()
	}
	item := p.items[p.gi]

	p.gi = (p.gi + 1) % p.cfg.size
	if p.gi == p.pi {
		p.blocked = getBocks
	} else {
		p.blocked = neitherBlocks
	}

	p.cond.L.Unlock()
	p.cond.Signal()
	return item
}

// Put will release a resource back to the pool. Put blocks if pool already full. If the Pool was
// initialized with a Reset function, it will be invoked with the resource as its sole argument,
// prior to the resource being added back to the pool. If Put is called when adding the resource to
// the pool _would_ result in having more elements in the pool than the pool size, the resource is
// effectively dropped on the floor after calling any optional Reset and Close methods on the
// resource.
func (p *ArrayPool) Put(item interface{}) {
	if p.cfg.reset != nil {
		p.cfg.reset(item)
	}

	// Put blocks when attempt to Put made at location next Get comes from
	p.cond.L.Lock()
	for p.blocked == putBlocks {
		p.cond.Wait()
	}
	p.items[p.pi] = item

	p.pi = (p.pi + 1) % p.cfg.size
	if p.gi == p.pi {
		p.blocked = putBlocks
	} else {
		p.blocked = neitherBlocks
	}

	p.cond.L.Unlock()
	p.cond.Signal()
}

// Close is called when the Pool is no longer needed, and the resources in the Pool ought to be
// released.  If a Pool has a close function, it will be invoked one time for each resource, with
// that resource as its sole argument.
func (p *ArrayPool) Close() error {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	var errs []error
	if p.cfg.close != nil {
		for _, item := range p.items {
			if err := p.cfg.close(item); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// prevent use of pool after Close
	p.items = nil
	p.gi = 0
	p.pi = 0
	p.blocked = getBocks

	if len(errs) == 0 {
		return nil
	}
	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return errors.New(strings.Join(messages, ", "))
}
