package bufpool

import "fmt"

// DefaultPoolSize is the default number of buffers that the free-list will maintain.
const DefaultPoolSize = 100

// Pool represents a data structure that maintains a free-list of buffers, accesible via Get and
// Put methods.
type Pool interface {
	Get() interface{}
	Put(interface{})
}

type poolConfig struct {
	poolSize int
	factory  func() (interface{}, error)
	reset    func(interface{})
}

// Configurator is a function that modifies a pool configuration structure.
type Configurator func(*poolConfig) error

// PoolSize specifies the number of buffers to maintain in the pool.  This option has no effect,
// however, on free-lists created with NewSyncPool, because the Go runtime dynamically maintains the
// size of pools created using sync.Pool.
//
//        package main
//
//        import (
//        	"log"
//        	"github.com/karrick/bufpool"
//        )
//
//        func main() {
//        	bp, err := bufpool.NewChanPool(bufpool.PoolSize(25))
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
func PoolSize(size int) Configurator {
	return func(pc *poolConfig) error {
		if size <= 0 {
			return fmt.Errorf("pool size must be greater than 0: %d", size)
		}
		pc.poolSize = size
		return nil
	}
}

func Factory(factory func() (interface{}, error)) Configurator {
	return func(pc *poolConfig) error {
		pc.factory = factory
		return nil
	}
}

func Reset(reset func(interface{})) Configurator {
	return func(pc *poolConfig) error {
		pc.reset = reset
		return nil
	}
}
