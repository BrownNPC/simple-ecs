package ecs

import (
	"math/bits"
	"reflect"
	"testing"
)

func TestSetGetClear(t *testing.T) {
	b := newBitset(0)
	if got := b.Get(5); got {
		t.Errorf("Get before Set: got true, want false")
	}
	b.Set(5)
	if !b.Get(5) {
		t.Errorf("Get after Set: got false, want true")
	}
	b.Clear(5)
	if b.Get(5) {
		t.Errorf("Get after Clear: got true, want false")
	}

	// Clearing out of bounds must not panic and leave bits unchanged
	b2 := newBitset(1)
	b2.Clear(1000)
	if !reflect.DeepEqual(b2.bits, newBitset(64).bits) {
		t.Errorf("Clear out of bounds modified bits: %v", b2.bits)
	}
}

func TestAutoGrow(t *testing.T) {
	b := newBitset(1) // capacity for bits [0..63]
	b.Set(100)
	if !b.Get(100) {
		t.Errorf("Auto-grow failed: bit 100 not set")
	}
	// Ensure underlying slice length grew
	if len(b.bits) < 2 {
		t.Errorf("Expected bits slice length >=2 after Set(100), got %d", len(b.bits))
	}
}

func TestAndOrAndNot(t *testing.T) {
	// Prepare two sets:
	// a: bits {1,3,5}, b: bits {3,4,5}
	a := newBitset(0)
	for _, i := range []uint32{1, 3, 5} {
		a.Set(i)
	}
	b := newBitset(0)
	for _, i := range []uint32{3, 4, 5} {
		b.Set(i)
	}

	// Test Or: a ∪ b = {1,3,4,5}
	or := a.Clone()
	or.Or(b)
	wantOr := []uint32{1, 3, 4, 5}
	if got := or.ActiveIDs(); !reflect.DeepEqual(got, wantOr) {
		t.Errorf("Or: got %v, want %v", got, wantOr)
	}

	// Test And: a ∩ b = {3,5}
	and := a.Clone()
	and.And(b)
	wantAnd := []uint32{3, 5}
	if got := and.ActiveIDs(); !reflect.DeepEqual(got, wantAnd) {
		t.Errorf("And: got %v, want %v", got, wantAnd)
	}

	// Test AndNot: a \ b = {1}
	andNot := a.Clone()
	andNot.AndNot(b)
	wantAndNot := []uint32{1}
	if got := andNot.ActiveIDs(); !reflect.DeepEqual(got, wantAndNot) {
		t.Errorf("AndNot: got %v, want %v", got, wantAndNot)
	}
}

func TestCloneIndependence(t *testing.T) {
	orig := newBitset(0)
	orig.Set(10)
	clone := orig.Clone()
	if !clone.Get(10) {
		t.Fatal("Clone did not copy set bit")
	}
	clone.Clear(10)
	if !orig.Get(10) {
		t.Errorf("Clearing clone should not affect original")
	}
}

func TestActiveIDsDenseSparse(t *testing.T) {
	// Dense: set every bit in [0..127]
	dense := newBitset(128)
	for i := uint32(0); i < 128; i++ {
		dense.Set(i)
	}
	gotDense := dense.ActiveIDs()
	if len(gotDense) != 128 {
		t.Fatalf("Dense ActiveIDs length: got %d, want 128", len(gotDense))
	}
	for i, id := range gotDense {
		if id != uint32(i) {
			t.Errorf("Dense ActiveIDs[%d] = %d, want %d", i, id, i)
		}
	}

	// Sparse: only bits {2, 64, 129}
	sparse := newBitset(0)
	for _, i := range []uint32{2, 64, 129} {
		sparse.Set(i)
	}
	gotSparse := sparse.ActiveIDs()
	wantSparse := []uint32{2, 64, 129}
	if !reflect.DeepEqual(gotSparse, wantSparse) {
		t.Errorf("Sparse ActiveIDs: got %v, want %v", gotSparse, wantSparse)
	}
}

func TestCountMatchesActiveIDs(t *testing.T) {
	// Validate that ones count equals len(ActiveIDs)
	b := newBitset(0)
	positions := []uint32{0, 7, 63, 64, 95, 127}
	for _, i := range positions {
		b.Set(i)
	}
	count := 0
	for _, w := range b.bits {
		count += bits.OnesCount64(w)
	}
	ids := b.ActiveIDs()
	if len(ids) != count {
		t.Errorf("Count mismatch: OnesCount=%d, len(ActiveIDs)=%d", count, len(ids))
	}
}
