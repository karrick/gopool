package gopool

import "fmt"

// DefaultSize is the default number of items that will be maintained in the
// pool.
const DefaultSize = 10

// Pool is the interface implemented by an object that acts as a free-list
// resource pool.
type Pool interface {
	Close() error
	Get() interface{}
	Put(interface{})
}

type config struct {
	close   func(interface{}) error     // close is an optional function to call with each item in the pool to clean up item resources.
	factory func() (interface{}, error) // factory is an optional function to call when need to create a new item for the pool.
	reset   func(interface{})           // reset is an optional function to call when returning an item to the pool in Put.
	minsize int                         // minsize is the minimum number of items to keep in the pool.
	maxsize int                         // maxsize is the maximum number of items to keep in the pool.
}

// Configurator is a function that modifies a pool configuration structure.
type Configurator func(*config) error

// Close specifies the optional function to be called once for each resource
// when the Pool is closed.
func Close(close func(interface{}) error) Configurator {
	return func(pc *config) error {
		pc.close = close
		return nil
	}
}

// Factory specifies the function used to make new elements for the pool.  The
// factory function is called to fill the pool N times during initialization,
// for a pool size of N.
func Factory(factory func() (interface{}, error)) Configurator {
	return func(pc *config) error {
		pc.factory = factory
		return nil
	}
}

// MinSize instructs pool to maintain a certain minimum size. Returns an error
// when the specified size is less than 0. The Factory must also be specified
// when the minimum size is greater than 0.
func MinSize(minsize int) Configurator {
	return func(pc *config) error {
		if minsize < 0 {
			return fmt.Errorf("cannot create a pool with a negative minimum size: %d", minsize)
		}
		pc.minsize = minsize
		return nil
	}
}

// Reset specifies the optional function to be called on resources when released
// back to the pool.  If a reset function is not specified, then resources are
// returned to the pool without any reset step.  For instance, if maintaining a
// Pool of buffers, a library may choose to have the reset function invoke the
// buffer's Reset method to free resources prior to returning the buffer to the
// Pool.
func Reset(reset func(interface{})) Configurator {
	return func(pc *config) error {
		pc.reset = reset
		return nil
	}
}

// Size specifies the maximum number of items to maintain in the pool.
func Size(maxsize int) Configurator {
	return func(pc *config) error {
		if maxsize <= 0 {
			return fmt.Errorf("cannot create a pool without space for at least one item: %d", maxsize)
		}
		pc.maxsize = maxsize
		return nil
	}
}
