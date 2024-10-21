package join

import (
	// Documentation: https://pkg.go.dev/github.com/bits-and-blooms/bitset
	"github.com/bits-and-blooms/bitset"
)

// BloomFilter is a probabilistic data structure used to
// quickly determine if an element is not in a set.
type BloomFilter struct {
	size int64          // The initial size of the bloom filter in bits.
	bits *bitset.BitSet // The bitset underlying our bloom filter implementation.
}

// CreateFilter initializes a BloomFilter with the given size.
func CreateFilter(size int64) (bf *BloomFilter) {
	panic("Not yet implemented")
}

// Insert adds an element into the bloom filter.
func (filter *BloomFilter) Insert(key int64) {
	panic("Not yet implemented")
}

// Contains returns whether the given key can be found in the bloom filter.
func (filter *BloomFilter) Contains(key int64) bool {
	panic("Not yet implemented")
}
