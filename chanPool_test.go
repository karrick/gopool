package gopool

import (
	"bytes"
	"errors"
	"testing"
)

func makeBuffer() (interface{}, error) {
	return bytes.NewBuffer(make([]byte, 1024)), nil
}

func resetBuffer(item interface{}) {
	item.(*bytes.Buffer).Reset()
}

func closeBuffer(item interface{}) error {
	item.(*bytes.Buffer).Reset()
	return nil
}

func TestChanPoolErrorWithoutFactory(t *testing.T) {
	pool, err := New()
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestChanPoolErrorWithNonPositiveSize(t *testing.T) {
	pool, err := New(Size(0))
	if pool != nil {
		t.Errorf("Actual: %#v; Expected: %#v", pool, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}

	pool, err = New(Size(-1))
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
	_, err := New(Size(size),
		Factory(func() (interface{}, error) {
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
	pool, err := New(
		Factory(func() (interface{}, error) {
			return nil, nil
		}),
		Reset(func(item interface{}) {
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
	pool, err := New(Size(1),
		Factory(func() (interface{}, error) {
			return nil, nil
		}),
		Close(func(_ interface{}) error {
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
	pool, err := New(Factory(makeBuffer), Reset(resetBuffer), Close(closeBuffer))
	if err != nil {
		t.Fatalf("Actual: %#v; Expected: %#v", err, nil)
	}
	test(t, pool)
}

func TestChanPoolSize(t *testing.T) {
	pool, err := New(Factory(makeBuffer), Reset(resetBuffer), Close(closeBuffer), Size(20))
	if err != nil {
		t.Fatalf("Actual: %#v; Expected: %#v", err, nil)
	}
	test(t, pool)
}

func BenchmarkLowConcurrency(b *testing.B) {
	pool, _ := New(Factory(makeBuffer), Reset(resetBuffer), Close(closeBuffer), Size(100))
	bench(b, pool, lowConcurrency)
}

func BenchmarkMediumConcurrency(b *testing.B) {
	pool, _ := New(Factory(makeBuffer), Reset(resetBuffer), Close(closeBuffer), Size(1000))
	bench(b, pool, medConcurrency)
}

func BenchmarkHighConcurrency(b *testing.B) {
	pool, _ := New(Factory(makeBuffer), Reset(resetBuffer), Close(closeBuffer), Size(10000))
	bench(b, pool, highConcurrency)
}
