package gopool_test

import (
	"errors"
	"testing"

	"github.com/karrick/gopool"
)

func TestChanPoolErrorWithoutFactory(t *testing.T) {
	pool, err := gopool.New()
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestChanPoolErrorWithNonPositiveSize(t *testing.T) {
	pool, err := gopool.New(gopool.Size(0))
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}

	pool, err = gopool.New(gopool.Size(-1))
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestChanPoolCreatesSizeItems(t *testing.T) {
	var size = 42
	var factoryInvoked int
	_, err := gopool.New(gopool.Size(size),
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

func TestChanPoolInvokesReset(t *testing.T) {
	var resetInvoked int
	pool, err := gopool.New(
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

func TestChanPoolInvokesClose(t *testing.T) {
	var closeInvoked int
	pool, err := gopool.New(gopool.Size(1),
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

func TestChanPool(t *testing.T) {
	pool, err := gopool.New(gopool.Factory(makeBuffer) /* gopool.Reset(resetBuffer), */, gopool.Close(closeBuffer))
	if err != nil {
		t.Fatalf("Actual: %#v; Expected: %#v", err, nil)
	}
	test(t, pool)
}

func TestChanPoolSize(t *testing.T) {
	pool, err := gopool.New(gopool.Factory(makeBuffer) /* gopool.Reset(resetBuffer), */, gopool.Close(closeBuffer), gopool.Size(lowCap))
	if err != nil {
		t.Fatal(err)
	}
	test(t, pool)
}

func BenchmarkChanLowConcurrency(b *testing.B) {
	pool, _ := gopool.New(gopool.Factory(makeBuffer) /* gopool.Reset(resetBuffer), */, gopool.Close(closeBuffer), gopool.Size(lowCap))
	bench(b, pool, lowConcurrency)
}

func BenchmarkChanMediumConcurrency(b *testing.B) {
	pool, _ := gopool.New(gopool.Factory(makeBuffer) /* gopool.Reset(resetBuffer), */, gopool.Close(closeBuffer), gopool.Size(medCap))
	bench(b, pool, medConcurrency)
}

func BenchmarkChanHighConcurrency(b *testing.B) {
	pool, _ := gopool.New(gopool.Factory(makeBuffer) /* gopool.Reset(resetBuffer), */, gopool.Close(closeBuffer), gopool.Size(largeCap))
	bench(b, pool, highConcurrency)
}
