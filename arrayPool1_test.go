package gopool_test

import (
	"errors"
	"testing"

	"github.com/karrick/gopool"
)

func TestArrayPool1ErrorWithoutFactory(t *testing.T) {
	t.Skip()
	pool, err := gopool.NewArrayPool1()
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestArrayPool1ErrorWithNonPositiveSize(t *testing.T) {
	pool, err := gopool.NewArrayPool1(gopool.Size(0))
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}

	pool, err = gopool.NewArrayPool1(gopool.Size(-1))
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestArrayPool1CreatesSizeItems(t *testing.T) {
	t.Skip()
	var size = 42
	var factoryInvoked int
	_, err := gopool.NewArrayPool1(gopool.Size(size),
		gopool.Factory(func() (interface{}, error) {
			factoryInvoked++
			return nil, nil
		}))
	ensureError(t, err)

	if actual, expected := factoryInvoked, size; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func TestArrayPool1InvokesReset(t *testing.T) {
	var resetInvoked int
	pool, err := gopool.NewArrayPool1(
		gopool.Factory(func() (interface{}, error) {
			return nil, nil
		}),
		gopool.Reset(func(item interface{}) {
			resetInvoked++
		}))
	ensureError(t, err)
	pool.Put(pool.Get())
	if actual, expected := resetInvoked, 1; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func TestArrayPool1InvokesClose(t *testing.T) {
	var closeInvoked int
	pool, err := gopool.NewArrayPool1(gopool.Size(1),
		gopool.Factory(func() (interface{}, error) {
			return nil, nil
		}),
		gopool.Close(func(_ interface{}) error {
			closeInvoked++
			return errors.New("foo")
		}))
	ensureError(t, err)
	if err := pool.Close(); err == nil || err.Error() != "foo" {
		t.Errorf("Actual: %#v; Expected: %#v", err, "foo")
	}
	if actual, expected := closeInvoked, 1; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func newArrayPool1(tb testing.TB, count int) gopool.Pool {
	pool, err := gopool.NewArrayPool1(
		gopool.Close(closeBuffer),
		gopool.Factory(makeBuffer),
		gopool.MinSize(count),
		gopool.Reset(resetBuffer),
		gopool.Size(count))
	ensureError(tb, err)
	return pool
}

func TestArrayPool1(t *testing.T) {
	test(t, newArrayPool1(t, lowCap))
}

func BenchmarkArrayPool1LowConcurrency(b *testing.B) {
	bench(b, newArrayPool1(b, lowCap), lowConcurrency)
}

func BenchmarkArrayPool1MediumConcurrency(b *testing.B) {
	bench(b, newArrayPool1(b, medCap), medConcurrency)
}

func BenchmarkArrayPool1HighConcurrency(b *testing.B) {
	bench(b, newArrayPool1(b, largeCap), highConcurrency)
}
