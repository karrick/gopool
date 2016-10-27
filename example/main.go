package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/karrick/gopool"
)

const (
	bufSize  = 64 * 1024
	poolSize = 25
)

func main() {
	const iterationCount = 1000
	const parallelCount = 100

	makeBuffer := func() (interface{}, error) {
		return bytes.NewBuffer(make([]byte, 0, bufSize)), nil
	}

	resetBuffer := func(item interface{}) {
		item.(*bytes.Buffer).Reset()
	}

	bp, err := gopool.New(gopool.Size(poolSize), gopool.Factory(makeBuffer), gopool.Reset(resetBuffer))
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(parallelCount)

	for i := 0; i < parallelCount; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < iterationCount; j++ {
				if err := grabBufferAndUseIt(bp); err != nil {
					fmt.Println(err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func grabBufferAndUseIt(pool gopool.Pool) error {
	// WARNING: Must ensure resource returns to pool otherwise gopool will deadlock once all
	// resources used.
	bb := pool.Get().(*bytes.Buffer)
	defer pool.Put(bb) // IMPORTANT: defer here to ensure invoked even when subsequent code bails

	for k := 0; k < bufSize; k++ {
		if rand.Intn(100000000) == 1 {
			return errors.New("random error to illustrate need to return resource to pool")
		}
		bb.WriteByte(byte(k % 256))
	}
	return nil
}
