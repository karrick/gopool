package bufpool

import (
	"bytes"
	"testing"
)

func TestChanPoolErrorWithoutFactory(t *testing.T) {
	bp, err := NewChanPool()
	if bp != nil {
		t.Errorf("Actual: %#v; Expected: %#v", bp, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestChanPoolErrorWithNonPositivePoolSize(t *testing.T) {
	bp, err := NewChanPool(PoolSize(0))
	if bp != nil {
		t.Errorf("Actual: %#v; Expected: %#v", bp, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}

	bp, err = NewChanPool(PoolSize(-1))
	if bp != nil {
		t.Errorf("Actual: %#v; Expected: %#v", bp, nil)
	}
	if err == nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, "not nil")
	}
}

func TestChanPoolInvokesReset(t *testing.T) {
	var resetInvoked bool
	bp, err := NewChanPool(
		Reset(func(item interface{}) {
			resetInvoked = true
		}),
		Factory(func() (interface{}, error) {
			return nil, nil
		}))
	if err != nil {
		t.Fatal(err)
	}
	bp.Put(bp.Get())
	if actual, expected := resetInvoked, true; actual != expected {
		t.Errorf("Actual: %#v; Expected: %#v", actual, expected)
	}
}

func bufferFactory() (interface{}, error) {
	return bytes.NewBuffer(make([]byte, 1024)), nil
}

func reset(item interface{}) {
	item.(*bytes.Buffer).Reset()
}

func TestChanPoolPoolSize(t *testing.T) {
	bp, err := NewChanPool(PoolSize(10), Factory(bufferFactory))
	if err != nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, nil)
	}
	test(t, bp)
}

func TestChanPool(t *testing.T) {
	bp, err := NewChanPool(Factory(bufferFactory))
	if err != nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, nil)
	}
	test(t, bp)
}

func BenchmarkLowChanPool(b *testing.B) {
	bp, _ := NewChanPool(Factory(bufferFactory), PoolSize(3))
	bench(b, bp, lowConcurrency)
}

func BenchmarkMedChanPool(b *testing.B) {
	bp, _ := NewChanPool(PoolSize(5), Factory(bufferFactory))
	bench(b, bp, medConcurrency)
}

func BenchmarkHighChanPool(b *testing.B) {
	bp, _ := NewChanPool(Factory(bufferFactory))
	bench(b, bp, highConcurrency)
}
