package ecs_test

import (
	"sync"
	"testing"

	"github.com/BrownNPC/simple-ecs"
)

// 5. Concurrency & Thread Safety
func TestConcurrentEntityCreation(t *testing.T) {
	pool := ecs.New(1000)
	var wg sync.WaitGroup
	entities := make(chan ecs.Entity, 1000)

	// Create 1000 entities concurrently
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			entities <- ecs.NewEntity(pool)
		}()
	}

	wg.Wait()
	close(entities)

	// Verify all entities are unique
	unique := make(map[ecs.Entity]bool)
	for e := range entities {
		if unique[e] {
			t.Fatalf("Duplicate entity created: %v", e)
		}
		unique[e] = true
	}
}

func TestConcurrentComponentOperations(t *testing.T) {
	for i := 0; i <= 1000; i++ {
		pool := ecs.New(100)
		e := ecs.NewEntity(pool)

		type Position struct{ X, Y float32 }

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(3)
			// Concurrent Add
			go func() {
				defer wg.Done()
				ecs.Add(pool, e, Position{1, 2})
			}()
			// Concurrent Remove
			go func() {
				defer wg.Done()
				ecs.Remove[Position](pool, e)
			}()
			// Concurrent Get
			go func() {
				defer wg.Done()
				_ = ecs.GetStorage[Position](pool).Get(e)
			}()
		}
		wg.Wait()
	}
}

func TestParallelReadWrite(t *testing.T) {
	pool := ecs.New(100)
	e := ecs.NewEntity(pool)
	type Position struct{ X, Y float32 }

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		// Writers
		go func(i int) {
			defer wg.Done()
			ecs.Add(pool, e, Position{X: float32(i)})
			ecs.GetStorage[Position](pool).Update(e, Position{X: float32(i)})
		}(i)
		// Readers
		go func() {
			defer wg.Done()
			_ = ecs.Has[Position](pool, e)
			_ = ecs.GetStorage[Position](pool).Get(e)
		}()
	}
	wg.Wait()
}

// 6. Edge Cases & Error Handling
func TestInvalidEntityHandling(t *testing.T) {
	pool := ecs.New(10)
	e := ecs.NewEntity(pool)
	type Position struct{ X, Y float32 }

	// Kill entity first
	ecs.Kill(pool, e)

	// Try operations on dead entity
	ecs.Add(pool, e, Position{1, 2})
	if ecs.Has[Position](pool, e) {
		t.Error("Dead entity should not receive components")
	}

	// Test Remove on dead entity (should be no-op)
	ecs.Remove[Position](pool, e)
}

func TestUpdateDeadEntityPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Updating dead entity should panic")
		}
	}()

	pool := ecs.New(10)
	e := ecs.NewEntity(pool)
	type Position struct{ X, Y float32 }

	ecs.Add(pool, e, Position{1, 2})
	ecs.Kill(pool, e)
	ecs.GetStorage[Position](pool).Update(e, Position{3, 4})
}

func TestComponentZeroValue(t *testing.T) {
	pool := ecs.New(10)
	e := ecs.NewEntity(pool)
	type Position struct{ X, Y float32 }

	// Add component and kill entity
	ecs.Add(pool, e, Position{1, 2})
	ecs.Kill(pool, e)

	// Verify component was zeroed
	st := ecs.GetStorage[Position](pool)
	if st.EntityHasComponent(e) {
		t.Error("Killed entity's component should be zeroed")
	}
}

// Mutex Integrity Test (indirectly tested via race detector)
func TestNoDeadlocks(t *testing.T) {
	pool := ecs.New(1000)
	type Position struct{ X, Y float32 }
	type Velocity struct{ X, Y float32 }

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			e := ecs.NewEntity(pool)
			ecs.Add(pool, e, Position{float32(i), 0})
			ecs.Add(pool, e, Velocity{0, float32(i)})
			ecs.GetStorage[Position](pool).Update(e, Position{1, 1})
			ecs.Remove[Velocity](pool, e)
			ecs.Kill(pool, e)
		}(i)
	}
	wg.Wait()
}
