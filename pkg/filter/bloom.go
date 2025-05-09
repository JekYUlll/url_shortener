package filter

import (
	"github.com/bits-and-blooms/bloom/v3"
)

type Filter interface {
	// 添加到过滤器
	Add(shortCode string)
	// 是否存在
	Exists(shortCode string) bool
}

type BloomFilterImpl struct {
	b         *bloom.BloomFilter
	Capacity  uint    // 容量
	ErrorRate float64 // 误差率
}

func NewBloomFilter(capacity uint, errorRate float64) *BloomFilterImpl {
	b := bloom.NewWithEstimates(capacity, errorRate)
	return &BloomFilterImpl{
		b:         b,
		Capacity:  capacity,
		ErrorRate: errorRate,
	}
}

func (f *BloomFilterImpl) Add(shortCode string) {
	f.b.Add([]byte(shortCode))
}

func (f *BloomFilterImpl) Exists(shortCode string) bool {
	return f.b.Test([]byte(shortCode))
}
