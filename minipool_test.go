package gopool_test

import (
	"testing"

	"github.com/karrick/gopool"
)

func newMiniPool(tb testing.TB, count int) gopool.Pool {
	pool, err := gopool.NewMiniPool(
		gopool.Close(closeBuffer),
		gopool.Factory(makeBuffer),
		gopool.MinSize(count),
		gopool.Reset(resetBuffer),
		gopool.Size(count))
	if err != nil {
		tb.Fatal(err)
	}
	return pool
}

func TestMiniPool(t *testing.T) {
	test(t, newMiniPool(t, lowCap))
}

func BenchmarkMiniLowConcurrency(b *testing.B) {
	bench(b, newMiniPool(b, lowCap), lowConcurrency)
}

func BenchmarkMiniMediumConcurrency(b *testing.B) {
	bench(b, newMiniPool(b, medCap), medConcurrency)
}

func BenchmarkMiniHighConcurrency(b *testing.B) {
	bench(b, newMiniPool(b, largeCap), highConcurrency)
}
