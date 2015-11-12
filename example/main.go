package main

import (
	"bytes"
	"log"

	"github.com/karrick/gopool"
)

func main() {
	makeBuffer := func() (interface{}, error) {
		return new(bytes.Buffer), nil
	}

	resetBuffer := func(item interface{}) {
		item.(*bytes.Buffer).Reset()
	}

	bp, err := gopool.New(gopool.Size(25),
		gopool.Factory(makeBuffer),
		gopool.Reset(resetBuffer))
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				bb := bp.Get().(*bytes.Buffer)
				for k := 0; k < 4096; k++ {
					bb.WriteByte(byte(k % 256))
				}
				bp.Put(bb)
			}
		}()
	}
}
