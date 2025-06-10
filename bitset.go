package ecs

import (
	"math/bits"
	"slices"
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

	// Grow the bitset if needed
	if int(word) >= len(b.bits) {
		newSize := word + 1
		newBits := make([]uint64, newSize)
		copy(newBits, b.bits)
		b.bits = newBits
	}

	b.bits[word] |= 1 << bit
}

func (b *bitSet) Clear(i uint32) {
	word := i / 64
	if int(word) >= len(b.bits) {
		return // No-op if beyond current capacity
	}
	bit := i % 64
	b.bits[word] &^= 1 << bit
}

func (b *bitSet) Get(i uint32) bool {
	word := i / 64
	if int(word) >= len(b.bits) {
		return false // Unset if beyond current capacity
	}
	bit := i % 64
	return b.bits[word]&(1<<bit) != 0
}

func (b *bitSet) And(other *bitSet) {
	n := min(len(b.bits), len(other.bits))
	for i := range n {
		b.bits[i] = b.bits[i] & other.bits[i]
	}
	b.bits = b.bits[:n]
}

func (b *bitSet) Or(other *bitSet) {
	n := min(len(b.bits), len(other.bits))
	for i := range n {
		b.bits[i] |= other.bits[i]
	}
	// Extend if other is larger
	if len(other.bits) > len(b.bits) {
		b.bits = append(b.bits, other.bits[len(b.bits):]...)
	}
}

func (b *bitSet) AndNot(other *bitSet) {
	n := min(len(b.bits), len(other.bits))
	for i := range n {
		b.bits[i] &^= other.bits[i]
	}
	b.bits = b.bits[:n]
}

func (b *bitSet) Clone() *bitSet {
	return &bitSet{
		bits: slices.Clone(b.bits),
	}
}
func (b *bitSet) ActiveIDs() []uint32 {
    // Pre-count total bits so we can pre-allocate once
    total := 0
    for _, w := range b.bits {
        total += bits.OnesCount64(w)
    }

    ids := make([]uint32, 0, total)

    // For each word, loop only over set bits
    for wi, w := range b.bits {
        base := uint32(wi) * 64
        for w != 0 {
            // trailing zero gives index of lowest ’1’
            t := bits.TrailingZeros64(w)
            ids = append(ids, base+uint32(t))
            // clear that lowest ’1’
            w &^= 1 << t
        }
    }

    return ids
}
