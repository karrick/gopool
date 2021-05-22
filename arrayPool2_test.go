package gopool_test

import (
	"errors"
	"testing"

	"github.com/karrick/gopool"
)

func TestArrayPool2SizeChecks(t *testing.T) {
	t.Run("no size", func(t *testing.T) {
		_, err := gopool.NewArrayPool2()
		ensureError(t, err, "without specifying")
	})
	t.Run("minsize greater than max size", func(t *testing.T) {
		_, err := gopool.NewArrayPool2(gopool.MinSize(10), gopool.Size(9))
		ensureError(t, err, "greater than")
	})
	t.Run("maxsize less than one", func(t *testing.T) {
		_, err := gopool.NewArrayPool2(gopool.Size(0))
		ensureError(t, err, "at least one item")

		_, err = gopool.NewArrayPool2()
		ensureError(t, err, "maximum or minimum size")
	})
	t.Run("minsize less than zero", func(t *testing.T) {
		_, err := gopool.NewArrayPool2(
			gopool.MinSize(-1))
		ensureError(t, err, "negative minimum size")
	})
	t.Run("minsize without factory", func(t *testing.T) {
		_, err := gopool.NewArrayPool2(
			gopool.MinSize(10))
		ensureError(t, err, "factory")
	})
}

func TestArrayPool2ErrorWithoutFactory(t *testing.T) {
	_, err := gopool.NewArrayPool2(gopool.MinSize(1))
	ensureError(t, err, "factory")
}

func TestArrayPool2ErrorWithNonPositiveSize(t *testing.T) {
	pool, err := gopool.NewArrayPool2(gopool.Size(0))
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}

	pool, err = gopool.NewArrayPool2(gopool.Size(-1))
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestArrayPool2CreatesSizeItems(t *testing.T) {
	var size = 42
	var factoryInvoked int
	_, err := gopool.NewArrayPool2(
		gopool.MinSize(size),
		gopool.Size(size+13),
		gopool.Factory(func() (interface{}, error) {
			factoryInvoked++
			return nil, nil
		}))
	ensureError(t, err)

	if actual, expected := factoryInvoked, size; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func TestArrayPool2InvokesReset(t *testing.T) {
	var resetInvoked int
	pool, err := gopool.NewArrayPool2(
		gopool.Factory(func() (interface{}, error) {
			return nil, nil
		}),
		gopool.MinSize(13),
		// gopool.Size(42),
		gopool.Reset(func(item interface{}) {
			resetInvoked++
		}))
	ensureError(t, err)
	pool.Put(pool.Get())
	if actual, expected := resetInvoked, 1; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func TestArrayPool2InvokesClose(t *testing.T) {
	var closeInvoked int

	pool, err := gopool.NewArrayPool2(gopool.Size(1),
		gopool.Factory(func() (interface{}, error) {
			return nil, nil
		}),
		gopool.Size(7),
		gopool.MinSize(3),
		gopool.Close(func(_ interface{}) error {
			closeInvoked++
			return errors.New("foo")
		}),
	)

	if err != nil {
		t.Fatal(err)
	}
	ensureError(t, pool.Close(), "foo, foo, foo")
	if actual, expected := closeInvoked, 3; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func newArrayPool2(tb testing.TB, count int) gopool.Pool {
	pool, err := gopool.NewArrayPool2(
		gopool.Close(closeBuffer),
		gopool.Factory(makeBuffer),
		gopool.MinSize(count),
		gopool.Reset(resetBuffer),
		gopool.Size(count))
	ensureError(tb, err)
	return pool
}

func TestArrayPool2(t *testing.T) {
	test(t, newArrayPool2(t, lowCap))
}

func BenchmarkArrayPool2LowConcurrency(b *testing.B) {
	bench(b, newArrayPool2(b, lowCap), lowConcurrency)
}

func BenchmarkArrayPool2MediumConcurrency(b *testing.B) {
	bench(b, newArrayPool2(b, medCap), medConcurrency)
}

func BenchmarkArrayPool2HighConcurrency(b *testing.B) {
	bench(b, newArrayPool2(b, largeCap), highConcurrency)
}
