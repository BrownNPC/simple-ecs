package ecs

import (
	"math/bits"
	"sync"
)

type bitSet struct {
	bits []uint64
}

func newBitset(size uint32) *bitSet {
	return &bitSet{
		bits: make([]uint64, (size+63)/64),
	}
}

func (b *bitSet) Set(i uint32) {
	word := i / 64
	bit := i % 64
	b.bits[word] |= 1 << bit
}

func (b *bitSet) Clear(i uint32) {
	word := i / 64
	bit := i % 64
	b.bits[word] &^= 1 << bit
}

func (b *bitSet) Get(i uint32) bool {
	word := i / 64
	bit := i % 64
	return b.bits[word]&(1<<bit) != 0
}

func (b *bitSet) And(other *bitSet) {
	n := min(len(b.bits), len(other.bits))
	for i := range n {
		b.bits[i] &= other.bits[i]
	}
}

func (b *bitSet) Or(other *bitSet) {
	n := min(len(b.bits), len(other.bits))
	for i := range n {
		b.bits[i] |= other.bits[i]
	}
}

func (b *bitSet) AndNot(other *bitSet) {
	n := min(len(b.bits), len(other.bits))
	for i := range n {
		b.bits[i] &^= other.bits[i]
	}
}

var bitSetPool = sync.Pool{
	New: func() any {
		return &bitSet{}
	},
}

func (b *bitSet) Clone() *bitSet {
	cloned := bitSetPool.Get().(*bitSet)

	if cap(cloned.bits) >= len(b.bits) {
		cloned.bits = cloned.bits[:len(b.bits)]
	} else {
		cloned.bits = make([]uint64, len(b.bits))
	}
	copy(cloned.bits, b.bits)

	return cloned
}

func (b *bitSet) Release() {
	clear(b.bits) // Only clear the used part
	bitSetPool.Put(b)
}
func (b *bitSet) ActiveIDs() []uint32 {
	total := 0
	for _, w := range b.bits {
		total += bits.OnesCount64(w)
	}

	ids := make([]uint32, 0, total)
	for wi, w := range b.bits {
		base := uint32(wi) * 64
		for w != 0 {
			t := bits.TrailingZeros64(w)
			ids = append(ids, base+uint32(t))
			w &^= 1 << t
		}
	}
	return ids
}
