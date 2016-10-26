# gopool

Gopool offers a way to maintain a free-list, or a pool of resources in
Go programs.

## Description

It is often the case that resource setup and teardown can be quite
demanding, and it is desirable to reuse resources rather than close
them and create new ones when needed. Two such examples are network
sockets to a given peer, and large byte buffers for building query
responses.

Although most customizations are optional, it does require
specification of a customized setup function to create new resources.
Optional resources include specifying the size of the resource pool,
specifying a per-use reset function, and specifying a close function
to be called when the pool is no longer needed. The close function is
called one time for each resource in the pool, with that resource as
the close function's sole argument.

### Usage

Documentation is available via
[![GoDoc](https://godoc.org/github.com/karrick/gopool?status.svg)](https://godoc.org/github.com/karrick/gopool).

### Example

The most basic example is creating a buffer pool and using it.

```Go
	package main
	
	import (
		"bytes"
		"log"
		"sync"
	
		"github.com/karrick/gopool"
	)
	
	func main() {
		makeBuffer := func() (interface{}, error) {
			return new(bytes.Buffer), nil
		}
	
		resetBuffer := func(item interface{}) {
			item.(*bytes.Buffer).Reset()
		}
	
		bp, err := gopool.New(gopool.Size(25), gopool.Factory(makeBuffer), gopool.Reset(resetBuffer))
		if err != nil {
			log.Fatal(err)
		}
		var wg sync.WaitGroup
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					bb := bp.Get().(*bytes.Buffer)
					for k := 0; k < 4096; k++ {
						bb.WriteByte(byte(k % 256))
					}
					bp.Put(bb)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
```
