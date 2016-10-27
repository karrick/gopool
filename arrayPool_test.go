package gopool_test

import (
	"errors"
	"testing"

	"github.com/karrick/gopool"
)

func TestArrayPoolErrorWithoutFactory(t *testing.T) {
	pool, err := gopool.NewArrayPool()
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestArrayPoolErrorWithNonPositiveSize(t *testing.T) {
	pool, err := gopool.NewArrayPool(gopool.Size(0))
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}

	pool, err = gopool.NewArrayPool(gopool.Size(-1))
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestArrayPoolCreatesSizeItems(t *testing.T) {
	var size = 42
	var factoryInvoked int
	_, err := gopool.NewArrayPool(gopool.Size(size),
		gopool.Factory(func() (interface{}, error) {
			factoryInvoked++
			return nil, nil
		}))
	if err != nil {
		t.Fatal(err)
	}

	if actual, expected := factoryInvoked, size; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func TestArrayPoolInvokesReset(t *testing.T) {
	var resetInvoked int
	pool, err := gopool.NewArrayPool(
		gopool.Factory(func() (interface{}, error) {
			return nil, nil
		}),
		gopool.Reset(func(item interface{}) {
			resetInvoked++
		}))
	if err != nil {
		t.Fatal(err)
	}
	pool.Put(pool.Get())
	if actual, expected := resetInvoked, 1; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func TestArrayPoolInvokesClose(t *testing.T) {
	var closeInvoked int
	pool, err := gopool.NewArrayPool(gopool.Size(1),
		gopool.Factory(func() (interface{}, error) {
			return nil, nil
		}),
		gopool.Close(func(_ interface{}) error {
			closeInvoked++
			return errors.New("foo")
		}))
	if err != nil {
		t.Fatal(err)
	}
	if err := pool.Close(); err == nil || err.Error() != "foo" {
		t.Errorf("Actual: %#v; Expected: %#v", err, "foo")
	}
	if actual, expected := closeInvoked, 1; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func TestArrayPool(t *testing.T) {
	pool, err := gopool.NewArrayPool(gopool.Factory(makeBuffer), gopool.Reset(resetBuffer), gopool.Close(closeBuffer))
	if err != nil {
		t.Fatalf("Actual: %#v; Expected: %#v", err, nil)
	}
	test(t, pool)
}

func TestArrayPoolSize(t *testing.T) {
	pool, err := gopool.NewArrayPool(gopool.Factory(makeBuffer), gopool.Reset(resetBuffer), gopool.Close(closeBuffer), gopool.Size(20))
	if err != nil {
		t.Fatalf("Actual: %#v; Expected: %#v", err, nil)
	}
	test(t, pool)
}

func BenchmarkArrayLowConcurrency(b *testing.B) {
	pool, _ := gopool.NewArrayPool(gopool.Factory(makeBuffer), gopool.Reset(resetBuffer), gopool.Close(closeBuffer), gopool.Size(100))
	bench(b, pool, lowConcurrency)
}

func BenchmarkArrayMediumConcurrency(b *testing.B) {
	pool, _ := gopool.NewArrayPool(gopool.Factory(makeBuffer), gopool.Reset(resetBuffer), gopool.Close(closeBuffer), gopool.Size(1000))
	bench(b, pool, medConcurrency)
}

func BenchmarkArrayHighConcurrency(b *testing.B) {
	pool, _ := gopool.NewArrayPool(gopool.Factory(makeBuffer), gopool.Reset(resetBuffer), gopool.Close(closeBuffer), gopool.Size(10000))
	bench(b, pool, highConcurrency)
}
