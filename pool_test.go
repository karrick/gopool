package gopool_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/karrick/gopool"
)

////////////////////////////////////////

func ensureError(tb testing.TB, err error, contains ...string) {
	tb.Helper()
	if len(contains) == 0 || (len(contains) == 1 && contains[0] == "") {
		if err != nil {
			tb.Fatalf("GOT: %v; WANT: %v", err, contains)
		}
	} else if err == nil {
		tb.Errorf("GOT: %v; WANT: %v", err, contains)
	} else {
		for _, stub := range contains {
			if stub != "" && !strings.Contains(err.Error(), stub) {
				tb.Errorf("GOT: %v; WANT: %q", err, stub)
			}
		}
	}
}

////////////////////////////////////////

const defaultBufSize = 1024
const defaultMaxKeep = 1024 * 128

const lowConcurrency = 16
const medConcurrency = 128
const highConcurrency = 1024

const lowCap = 100
const medCap = 1000
const largeCap = 10000

////////////////////////////////////////

func makeBuffer() (interface{}, error) {
	return bytes.NewBuffer(make([]byte, defaultBufSize)), nil
}

func resetBuffer(item interface{}) {
	item.(*bytes.Buffer).Reset()
}

func closeBuffer(item interface{}) error {
	item.(*bytes.Buffer).Reset()
	return nil
}

////////////////////////////////////////

func testC(bp gopool.Pool, concurrency, loops int) {
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

func test(t *testing.T, bp gopool.Pool) {
	const concurrency = 128
	const loops = 128
	testC(bp, concurrency, loops)
}

func bench(b *testing.B, bp gopool.Pool, concurrency int) {
	b.ResetTimer() // do not include initialization time in benchmarks
	testC(bp, concurrency, b.N)
}
