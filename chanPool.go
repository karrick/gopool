package gopool

import "errors"

// ChanPool maintains a free-list of buffers.
type ChanPool struct {
	ch chan interface{}
	pc poolConfig
}

// NewChanPool creates a new Pool. The pool size, size of new buffers, and max size of buffers
// to keep when returned to the pool can all be customized.
//
//        package main
//
//        import (
//        	"log"
//
//        	"github.com/karrick/bufpool"
//        )
//
//        func main() {
//        	bp, err := bufpool.NewChanPool()
//        	if err != nil {
//        		log.Fatal(err)
//        	}
//        	for i := 0; i < 4*bufpool.DefaultPoolSize; i++ {
//        		go func() {
//        			for j := 0; j < 1000; j++ {
//        				bb := bp.Get()
//        				for k := 0; k < 3*bufpool.DefaultBufSize; k++ {
//        					bb.WriteByte(byte(k % 256))
//        				}
//        				bp.Put(bb)
//        			}
//        		}()
//        	}
//        }
func NewChanPool(setters ...Configurator) (Pool, error) {
	pc := &poolConfig{
		poolSize: DefaultPoolSize,
		factory: func() (interface{}, error) {
			return nil, errors.New("ought to specify factory method")
		},
	}
	for _, setter := range setters {
		if err := setter(pc); err != nil {
			return nil, err
		}
	}
	bp := &ChanPool{
		ch: make(chan interface{}, pc.poolSize),
		pc: *pc,
	}
	for i := 0; i < bp.pc.poolSize; i++ {
		item, err := bp.pc.factory()
		if err != nil {
			return nil, err
		}
		bp.ch <- item
	}
	return bp, nil
}

// Get returns an initialized buffer from the free-list.
func (bp *ChanPool) Get() interface{} {
	return <-bp.ch
}

// Put will return a used buffer back to the free-list. If the capacity of the used buffer grew
// beyond the max buffer size, it will be discarded and its memory returned to the runtime.
func (bp *ChanPool) Put(item interface{}) {
	if bp.pc.reset != nil {
		bp.pc.reset(item)
	}
	bp.ch <- item
}
