package gopool

import (
	"fmt"
	"math"
	"sync/atomic"
)

type shardpool struct {
	pools        []minipool
	mask, gi, pi uint32
}

const minDepth = 8

func NewShardPoolCount(count int) (*shardpool, error) {
	var width int

	if count/4096 > minDepth {
		width = 4096
	} else if count/256 > minDepth {
		width = 256
	} else {
		width = 16
	}
	depth := int(math.Ceil(float64(count) / float64(width)))
	return NewShardPool(width, depth)
}

func NewShardPool(width, depth int) (*shardpool, error) {
	var mask uint32

	switch width {
	case 16:
		mask = 0xf
	case 256:
		mask = 0xff
	case 4096:
		mask = 0xfff
	default:
		return nil, fmt.Errorf("width must be a nibble boundary: %d", width)
	}

	pool := &shardpool{
		pools: make([]minipool, width),
		mask:  uint32(mask),
	}

	for i := 0; i < width; i++ {
		mini, err := NewMiniPool(Size(depth))
		if err != nil {
			return nil, err
		}
		pool.pools[i] = *mini
	}

	return pool, nil
}

func (pool *shardpool) Close() error {
	var err error
	for _, p := range pool.pools {
		if ep := p.Close(); err == nil {
			err = ep
		}
	}
	return err
}

func (pool *shardpool) Get() (item interface{}) {
	i := atomic.AddUint32(&pool.gi, 1) & pool.mask
	// fmt.Fprintf(os.Stderr, "Shard Get(%d)\n", i)
	return pool.pools[i].Get()
}

func (pool *shardpool) Put(item interface{}) {
	i := atomic.AddUint32(&pool.pi, 1) & pool.mask
	// fmt.Fprintf(os.Stderr, "Shard Put(%d)\n", i)
	pool.pools[i].Put(item)
}
