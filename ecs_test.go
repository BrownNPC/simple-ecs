// run the tests with:  go:build -race

package ecs_test

import (
	"testing"

	ecs "github.com/BrownNPC/simple-ecs" // Replace with actual import path
)

func TestNewEntityCreation(t *testing.T) {

	t.Run("Create entities up to limit", func(t *testing.T) {
		poolSize := 5
		p := ecs.New(poolSize)

		// Create entities up to pool size
		var entities []ecs.Entity
		for i := 0; i < poolSize; i++ {
			e := ecs.NewEntity(p)
			entities = append(entities, e)
		}

		// Verify all entities are unique and within bounds
		seen := make(map[ecs.Entity]bool)
		for _, e := range entities {
			if e < 0 || e >= ecs.Entity(poolSize) {
				t.Errorf("Entity %d out of bounds", e)
			}
			if seen[e] {
				t.Errorf("Duplicate entity %d created", e)
			}
			seen[e] = true
		}
	})

	t.Run("Panic when exceeding limit", func(t *testing.T) {
		poolSize := 1
		p := ecs.New(poolSize)

		// Fill pool
		for i := 0; i < poolSize; i++ {
			ecs.NewEntity(p)
		}

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when exceeding entity limit")
			}
		}()

		// This should panic
		ecs.NewEntity(p)
	})
}

func TestEntityReuse(t *testing.T) {

	poolSize := 3
	p := ecs.New(poolSize)

	// Create and kill an entity
	e := ecs.NewEntity(p)
	ecs.Kill(p, e)

	t.Run("Reuse from free list", func(t *testing.T) {
		newEntity := ecs.NewEntity(p)
		if newEntity != e {
			t.Errorf("Expected reused entity %d, got %d", e, newEntity)
		}
	})

	t.Run("Components properly reset", func(t *testing.T) {
		type Position struct{ x, y float32 }

		// Add component to original entity
		ecs.Add(p, e, Position{10, 20})
		ecs.Kill(p, e)

		// Reuse entity
		reused := ecs.NewEntity(p)
		// make sure entity is reused
		if reused != e {
			t.Errorf("Entity was not reused.")
		}

		// Check components
		if ecs.Has[Position](p, reused) {
			t.Error("Reused entity has residual component")
		}

		// Verify storage is clean
		storage := ecs.GetStorage[Position](p)
		if storage.EntityHasComponent(reused) {
			t.Error("Storage shows component for reused entity")
		}
	})
}

func TestKillEntities(t *testing.T) {

	poolSize := 5
	p := ecs.New(poolSize)
	e := ecs.NewEntity(p)

	t.Run("Mark as dead", func(t *testing.T) {
		ecs.Kill(p, e)
		if ecs.IsAlive(p, e) {
			t.Error("Entity still marked alive after kill")
		}
	})

	t.Run("Remove components from storage", func(t *testing.T) {
		type Health struct{ value int }

		// Add component and kill
		ecs.Add(p, e, Health{100})
		ecs.Kill(p, e)

		// Check storage
		storage := ecs.GetStorage[Health](p)
		if storage.EntityHasComponent(e) {
			t.Error("Component still present in storage after kill")
		}
	})

}

func TestIsAliveCheck(t *testing.T) {

	poolSize := 3
	p := ecs.New(poolSize)

	t.Run("True for alive entities", func(t *testing.T) {
		e := ecs.NewEntity(p)
		if !ecs.IsAlive(p, e) {
			t.Error("New entity not marked alive")
		}
	})

	t.Run("False for killed entities", func(t *testing.T) {
		e := ecs.NewEntity(p)
		ecs.Kill(p, e)
		if ecs.IsAlive(p, e) {
			t.Error("Killed entity still marked alive")
		}
	})
}

// Edge Cases & Error Handling
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

	// Remove on dead entity (should be no-op)
	ecs.Remove[Position](pool, e)
}

func TestUpdateDeadEntityPanic(t *testing.T) {
	t.Run("updating dead entity should not panic.", func(t *testing.T) {
		pool := ecs.New(10)
		e := ecs.NewEntity(pool)
		type Position struct{ X, Y float32 }

		ecs.Add(pool, e, Position{1, 2})
		ecs.Kill(pool, e)
		ecs.GetStorage[Position](pool).Update(e, Position{3, 4})
	})
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

