package gopool_test

import (
	"bytes"
	"testing"

	"github.com/karrick/gopool"
)

func newShardPoolCount(tb testing.TB, count int) gopool.Pool {
	pool, err := gopool.NewShardPoolCount(count)
	if err != nil {
		tb.Fatal(err)
	}
	for i := 0; i < count; i++ {
		pool.Put(bytes.NewBuffer(make([]byte, defaultBufSize)))
	}
	return pool
}

func newShardPool(tb testing.TB, width, depth int) gopool.Pool {
	pool, err := gopool.NewShardPool(width, depth)
	if err != nil {
		tb.Fatal(err)
	}
	for i := 0; i < width*depth; i++ {
		pool.Put(bytes.NewBuffer(make([]byte, defaultBufSize)))
	}
	return pool
}

func TestShardPool(t *testing.T) {
	test(t, newShardPoolCount(t, lowCap))
}

func BenchmarkShardLowConcurrency(b *testing.B) {
	bench(b, newShardPoolCount(b, lowCap), lowConcurrency)
}

func BenchmarkShardMediumConcurrency(b *testing.B) {
	bench(b, newShardPoolCount(b, medCap), medConcurrency)
}

func BenchmarkShardHighConcurrency(b *testing.B) {
	bench(b, newShardPoolCount(b, largeCap), highConcurrency)
}
