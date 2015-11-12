package gopool

import (
	"bytes"
	"sync"
	"testing"
)

const defaultBufSize = 1024
const defaultMaxKeep = 1024 * 128

func testC(bp Pool, concurrency, loops int) {
	const byteCount = defaultBufSize / 2

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for c := 0; c < concurrency; c++ {
		go func() {
			for j := 0; j < loops; j++ {
				bb := bp.Get().(*bytes.Buffer)
				max := byteCount
				if j%8 == 0 {
					max = 2 * defaultMaxKeep
				}
				for k := 0; k < max; k++ {
					bb.WriteByte(byte(k % 256))
				}
				bp.Put(bb)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func test(t *testing.T, bp Pool) {
	const concurrency = 128
	const loops = 128
	testC(bp, concurrency, loops)
}

const lowConcurrency = 16
const medConcurrency = 128
const highConcurrency = 1024

func bench(b *testing.B, bp Pool, concurrency int) {
	const loops = 256
	for i := 0; i < b.N; i++ {
		testC(bp, concurrency, loops)
	}
}
