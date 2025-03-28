package internal

import (
	"slices"
	"sync"
)

const bitsPerByte = 8

type BitSet struct {
	Data []byte
	Mu   sync.Mutex
}

func (b *BitSet) Set(pos uint) {
	b.Mu.Lock()
	defer b.Mu.Unlock()

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
	b.Mu.Lock()
	defer b.Mu.Unlock()

	byteIndex := pos / bitsPerByte
	if byteIndex >= uint(len(b.Data)) {
		return
	}
	bitIndex := pos % bitsPerByte
	b.Data[byteIndex] &^= 1 << bitIndex
}

func (b *BitSet) IsSet(pos uint) bool {
	b.Mu.Lock()
	defer b.Mu.Unlock()

	byteIndex := pos / bitsPerByte
	if byteIndex >= uint(len(b.Data)) {
		return false
	}
	bitIndex := pos % bitsPerByte
	return (b.Data[byteIndex] & (1 << bitIndex)) != 0
}

func (b *BitSet) And(other *BitSet) {
	b.Mu.Lock()
	other.Mu.Lock()
	defer other.Mu.Unlock()
	defer b.Mu.Unlock()
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
	b.Mu.Lock()
	defer b.Mu.Unlock()
	other.Mu.Lock()
	defer other.Mu.Unlock()
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
	b.Mu.Lock()
	defer b.Mu.Unlock()
	minLen := Minint(len(b.Data), len(other.Data))

	for i := 0; i < minLen; i++ {
		b.Data[i] &^= other.Data[i]
	}
}

func ActiveIndices[T ~uint32](b *BitSet) []T {
	b.Mu.Lock()
	defer b.Mu.Unlock()
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
	b.Mu.Lock()
	defer b.Mu.Unlock()
	return BitSet{Data: slices.Clone(b.Data)}
}

func Minint(v1, v2 int) int {
	return v2 + ((v1 - v2) & ((v1 - v2) >> 31))
}
