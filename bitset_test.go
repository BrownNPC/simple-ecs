package ecs

import (
	"reflect"
	"testing"
)

func TestSetGetClear(t *testing.T) {
	b := newBitset(128)

	b.Set(42)
	if !b.Get(42) {
		t.Errorf("Expected bit 42 to be set")
	}

	b.Clear(42)
	if b.Get(42) {
		t.Errorf("Expected bit 42 to be cleared")
	}
}


func TestAnd(t *testing.T) {
	a := newBitset(128)
	b := newBitset(128)

	a.Set(10)
	a.Set(20)
	b.Set(20)
	b.Set(30)

	a.And(b)

	if a.Get(10) {
		t.Errorf("Expected bit 10 to be cleared")
	}
	if !a.Get(20) {
		t.Errorf("Expected bit 20 to remain set")
	}
	if a.Get(30) {
		t.Errorf("Expected bit 30 to be cleared")
	}
}

func TestOr(t *testing.T) {
	a := newBitset(128)
	b := newBitset(128)

	a.Set(5)
	b.Set(6)

	a.Or(b)

	if !a.Get(5) || !a.Get(6) {
		t.Errorf("Expected both bits 5 and 6 to be set after OR")
	}
}

func TestAndNot(t *testing.T) {
	a := newBitset(128)
	b := newBitset(128)

	a.Set(5)
	a.Set(6)
	b.Set(6)

	a.AndNot(b)

	if !a.Get(5) {
		t.Errorf("Expected bit 5 to remain set")
	}
	if a.Get(6) {
		t.Errorf("Expected bit 6 to be cleared after AndNot")
	}
}

func TestActiveIDs(t *testing.T) {
	b := newBitset(128)
	expected := []uint32{3, 5, 64, 127}

	for _, i := range expected {
		b.Set(i)
	}

	ids := b.ActiveIDs()

	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("ActiveIDs mismatch.\nExpected: %v\nGot: %v", expected, ids)
	}
}


