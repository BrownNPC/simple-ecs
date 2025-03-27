package internal

import (
	"slices"
	"sync"
)

const bitsPerByte = 8

type BitSet struct {
	Data []byte
	mu   sync.Mutex
}

func (b *BitSet) Set(pos uint) {
	b.mu.Lock()
	defer b.mu.Unlock()

	byteIndex := pos / bitsPerByte
	bitIndex := pos % bitsPerByte

	if byteIndex >= uint(len(b.Data)) {
		newSize := byteIndex + 1
		newData := make([]byte, newSize)
		copy(newData, b.Data)
		b.Data = newData
	}

	b.Data[byteIndex] |= 1 << bitIndex
}

func (b *BitSet) Unset(pos uint) {
	b.mu.Lock()
	defer b.mu.Unlock()

	byteIndex := pos / bitsPerByte
	if byteIndex >= uint(len(b.Data)) {
		return
	}
	bitIndex := pos % bitsPerByte
	b.Data[byteIndex] &^= 1 << bitIndex
}

func (b *BitSet) IsSet(pos uint) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	byteIndex := pos / bitsPerByte
	if byteIndex >= uint(len(b.Data)) {
		return false
	}
	bitIndex := pos % bitsPerByte
	return (b.Data[byteIndex] & (1 << bitIndex)) != 0
}

func (b *BitSet) And(other *BitSet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	otherLen := len(other.Data)
	bLen := len(b.Data)
	minLen := Minint(bLen, otherLen)

	for i := 0; i < minLen; i++ {
		b.Data[i] &= other.Data[i]
	}

	if bLen > minLen {
		b.Data = b.Data[:minLen]
	}
}

func (b *BitSet) Or(other *BitSet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	otherLen := len(other.Data)
	bLen := len(b.Data)

	if otherLen > bLen {
		newData := make([]byte, otherLen)
		copy(newData, b.Data)
		b.Data = newData
	}

	for i := 0; i < otherLen; i++ {
		b.Data[i] |= other.Data[i]
	}
}

func (b *BitSet) AndNot(other *BitSet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	minLen := Minint(len(b.Data), len(other.Data))

	for i := 0; i < minLen; i++ {
		b.Data[i] &^= other.Data[i]
	}
}

func ActiveIndices[T ~uint32](b *BitSet) []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	ret := make([]T, 0, len(b.Data))
	for NthByte, byteVal := range b.Data {
		for NthBit := 0; NthBit < 8; NthBit++ {
			if byteVal&(1<<NthBit) != 0 {
				pos := uint(NthByte*8) + uint(NthBit)
				ret = append(ret, T(pos))
			}
		}
	}
	return ret
}

func (b *BitSet) Clone() BitSet {
	b.mu.Lock()
	defer b.mu.Unlock()
	return BitSet{Data: slices.Clone(b.Data)}
}

func Minint(v1, v2 int) int {
	return v2 + ((v1 - v2) & ((v1 - v2) >> 31))
}
