package ecs

import (
	"reflect"
	"slices"
	"testing"
)

// Test basic pool creation and initial state
func TestNewPool(t *testing.T) {
	p := New(5)
	if p.capacity != 6 { // capacity+1
		t.Errorf("expected capacity 6, got %d", p.capacity)
	}
	if p.TotalEntities != 0 {
		t.Errorf("expected TotalEntities 0, got %d", p.TotalEntities)
	}
}

// Test entity lifecycle: creation, alive status, killing, recycling, and generations
func TestEntityLifecycle(t *testing.T) {
	p := New(3)
	// create entities
	e1 := NewEntity(p)
	e2 := NewEntity(p)
	if e1 == 0 || e2 == 0 || e1 == e2 {
		t.Fatalf("invalid entity IDs: %d, %d", e1, e2)
	}
	if !IsAlive(p, e1) || !IsAlive(p, e2) {
		t.Errorf("entities should be alive after creation")
	}
	// kill e1
	Kill(p, e1)
	if IsAlive(p, e1) {
		t.Errorf("entity %d should be dead after Kill", e1)
	}
	// generation increment
	gen := GetGeneration(p, e1)
	if gen != 1 {
		t.Errorf("expected generation 1 after kill, got %d", gen)
	}
	// reuse should give same ID
	e1r := NewEntity(p)
	if e1r != e1 {
		t.Errorf("expected recycled ID %d, got %d", e1, e1r)
	}
	if !IsAlive(p, e1r) {
		t.Errorf("recycled entity should be alive")
	}
}

// Test IsAliveWithGeneration
func TestIsAliveWithGeneration(t *testing.T) {
	p := New(1)
	e := NewEntity(p)
	gen0 := GetGeneration(p, e)
	if !IsAliveWithGeneration(p, e, gen0) {
		t.Errorf("expected alive with matching generation %d", gen0)
	}
	Kill(p, e)
	// after kill, generation incremented
	gen1 := GetGeneration(p, e)
	if gen1 != gen0+1 {
		t.Fatalf("expected generation %d, got %d", gen0+1, gen1)
	}
	if IsAliveWithGeneration(p, e, gen0) {
		t.Errorf("should not be alive with old generation %d", gen0)
	}
}

// Test component storage operations: Add, Remove, Get, EntityHasComponent
func TestComponentStorage(t *testing.T) {
	type Comp struct{ Value int }
	p := New(5)
	e := NewEntity(p)
	// initially no component
	st := GetStorage[Comp](p)
	if st.EntityHasComponent(e) {
		t.Errorf("entity should have no component initially")
	}
	// add component
	Add(p, e, Comp{Value: 42})
	if !st.EntityHasComponent(e) {
		t.Errorf("entity should have component after Add")
	}
	c := st.Get(e)
	if c.Value != 42 {
		t.Errorf("expected component value 42, got %d", c.Value)
	}
	// remove component
	Remove[Comp](p, e)
	if st.EntityHasComponent(e) {
		t.Errorf("entity should not have component after Remove")
	}
	c2 := st.Get(e)
	if c2.Value != 0 {
		t.Errorf("expected zeroed component after Remove, got %d", c2.Value)
	}
}

// Test query methods: All, And, Or, ButNot
func TestStorageQueries(t *testing.T) {
	type A struct{ X int }
	type B struct{ Y int }
	p := New(10)
	// create entities
	es := make([]Entity, 5)
	for i := range 5 {
		es[i] = NewEntity(p)
	}
	stA := GetStorage[A](p)
	stB := GetStorage[B](p)
	// add A to even entities, B to first three
	for _, e := range es {
		if e%2 == 0 {
			Add(p, e, A{X: int(e)})
		}
		if e <= es[2] {
			Add(p, e, B{Y: int(e)})
		}
	}
	// All A
	expA := []Entity{}
	for _, e := range es {
		if e%2 == 0 {
			expA = append(expA, e)
		}
	}
	gotA := stA.All()
	sortEntities := func(arr []Entity) []Entity {
		sorted := make([]Entity, len(arr))
		copy(sorted, arr)
		slices.Sort(sorted)
		return sorted
	}
	expAs := sortEntities(expA)
	gotAs := sortEntities(gotA)
	if !reflect.DeepEqual(expAs, gotAs) {
		t.Errorf("All A mismatch: expected %v, got %v", expAs, gotAs)
	}
	// And: A and B
	expAnd := []Entity{}
	for _, e := range es {
		if e%2 == 0 && e <= es[2] {
			expAnd = append(expAnd, e)
		}
	}
	gotAnd := stA.And(stB)
	if !reflect.DeepEqual(sortEntities(expAnd), sortEntities(gotAnd)) {
		t.Errorf("And mismatch: expected %v, got %v", expAnd, gotAnd)
	}
	// Or: A or B
	expOr := []Entity{}
	for _, e := range es {
		if e%2 == 0 || e <= es[2] {
			expOr = append(expOr, e)
		}
	}
	gotOr := stA.Or(stB)
	if !reflect.DeepEqual(sortEntities(expOr), sortEntities(gotOr)) {
		t.Errorf("Or mismatch: expected %v, got %v", expOr, gotOr)
	}
	// ButNot: A but not B
	expButNot := []Entity{}
	for _, e := range es {
		if e%2 == 0 && !(e <= es[2]) {
			expButNot = append(expButNot, e)
		}
	}
	gotButNot := stA.ButNot(stB)
	if !reflect.DeepEqual(sortEntities(expButNot), sortEntities(gotButNot)) {
		t.Errorf("ButNot mismatch: expected %v, got %v", expButNot, gotButNot)
	}
}

