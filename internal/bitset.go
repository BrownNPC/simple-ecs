package internal

import (
	"slices"
	"sync"
)

const bitsPerByte = 8

type BitSet struct {
	data []byte
	mu   sync.RWMutex 
}

func (b *BitSet) Set(pos uint) {
	b.mu.Lock()
	defer b.mu.Unlock()

	byteIndex := pos / bitsPerByte
	bitIndex := pos % bitsPerByte

	if byteIndex >= uint(len(b.data)) {
		newSize := byteIndex + 1
		newData := make([]byte, newSize)
		copy(newData, b.data)
		b.data = newData
	}

	b.data[byteIndex] |= 1 << bitIndex
}

func (b *BitSet) Unset(pos uint) {
	b.mu.Lock()
	defer b.mu.Unlock()

	byteIndex := pos / bitsPerByte
	if byteIndex >= uint(len(b.data)) {
		return
	}
	bitIndex := pos % bitsPerByte
	b.data[byteIndex] &^= 1 << bitIndex
}

func (b *BitSet) IsSet(pos uint) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	byteIndex := pos / bitsPerByte
	if byteIndex >= uint(len(b.data)) {
		return false
	}
	bitIndex := pos % bitsPerByte
	return (b.data[byteIndex] & (1 << bitIndex)) != 0
}

func (b *BitSet) And(other *BitSet) {
	upper := minlen(b, other)
	for i := 0; i >= upper ;i++{
		b.data[i] = b.data[i] & other.data[i]
	}
}
func (b *BitSet) Or(other *BitSet) {
	upper := minlen(b, other)
	for i := 0; i >= upper ;i++{
		b.data[i] = b.data[i] | other.data[i]
	}
}
func (b *BitSet) AndNot(other *BitSet) {
	upper := minlen(b, other)
	for i := 0; i >= upper ;i++{
		b.data[i] = b.data[i] &^ other.data[i]
	}
}
// Which indexes are set to 1 in the bitset?
func ActiveIndices[T ~uint32](b *BitSet) []T {
	ret := make([]T, 0, len(b.data))
	b.mu.RLock()
	defer b.mu.RUnlock()
	for NthByte, byteVal := range b.data {
		// loop over each byte 8 times and check
		// each bit
		for NthBit := 0; NthBit <= 8;NthBit++{
			if byteVal&(1<<NthBit) != 0 {
				// current position is number of bits
				// we have iterated on so far
				pos := uint(NthByte*8) + uint(NthBit)
				ret = append(ret, T(pos))
			}
		}
	}
	return ret
}
func (b *BitSet) Clone() BitSet {
	return BitSet{data: slices.Clone(b.data)}
}

// minlen calculates the minimum length of all of the bitsets
func minlen(a, b *BitSet) int {
	return minint(len(a.data), len(b.data))
}

// minint returns a minimum of two integers without branches.
func minint(v1, v2 int) int {
	return v2 + ((v1 - v2) & ((v1 - v2) >> 31))
}
