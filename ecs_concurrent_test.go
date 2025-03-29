package ecs_test

import (
	"sync"
	"testing"

	"github.com/BrownNPC/simple-ecs"
)

// Concurrency & Thread Safety
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

// Mutex Integrity  (indirectly tested via race detector)
func TestNoDeadlocks(t *testing.T) {
	pool := ecs.New(1000)
	type Position struct{ X, Y int }
	type Velocity struct{ X, Y int }

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			e := ecs.NewEntity(pool)
			ecs.Add(pool, e, Position{i, 0})
			ecs.Add(pool, e, Velocity{0, i})
			ecs.GetStorage[Position](pool).Update(e, Position{1, 1})
			ecs.Remove[Velocity](pool, e)
			ecs.Kill(pool, e)
		}(i)
	}
	wg.Wait()
}

func TestRecycledEntitiesHaveZeroComponent(t *testing.T) {
for{
	type Health struct{ X, Y float32 }
	var wg sync.WaitGroup
	// first create dead entities with a non-zero component
	var deadEntitiesWithNonZeroComponent int
	var Nentities = 10
	pool := ecs.New(Nentities)
	entities := make([]ecs.Entity, 0, Nentities)
	for i := 0; i < Nentities; i++ {
		e := ecs.NewEntity(pool)
		ecs.Add(pool, e, Health{})
		entities = append(entities, e)
	}
	wg.Add(1)
	go func() {
		ecs.Kill(pool, entities...)
		wg.Done()
	}()
	func(p *ecs.Pool) {
		HEALTH := ecs.GetStorage[Health](pool)
		deadEntities := make([]ecs.Entity, 0, Nentities)
		for _, e := range HEALTH.And(nil) {
			hp := HEALTH.Get(e)
			hp.X -= 1
			hp.Y -= 1
			HEALTH.Update(e, hp)
			if len(deadEntities) > 0 {
				hp := HEALTH.Get(deadEntities[0])
				if hp.X == 0 {
					deadEntitiesWithNonZeroComponent += 1
				}
				deadEntities = deadEntities[1:]
			}

			if !ecs.IsAlive(pool, e) {
				deadEntities = append(deadEntities, e)
			}
		}
	}(pool)
	wg.Wait()
	//make sure dead entities with non zero components were created
	if deadEntitiesWithNonZeroComponent == 0 {
		continue // restart since actual test cant happen
	}
	// Begin the actual test
	POSITION := ecs.GetStorage[Health](pool)
	for i := 0; i < Nentities; i++ {
		e := ecs.NewEntity(pool)
		pos := POSITION.Get(e)
		if pos.X != 0 {
			t.Error("reused entity does not have a zero component")
		}
	}
	break
}}
func TestConcurrentUnsafeApi(t *testing.T) {
	pool := ecs.New(1000)
	type Position struct{ X, Y float32 }
	type Velocity struct{ X, Y float32 }
	var wg sync.WaitGroup
	wg.Add(5)
	go func(pool *ecs.Pool) {
		for i := 0; i < 5; i++ {
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
	}(pool)
	wg.Add(5)
	go func(pool *ecs.Pool) {
		defer wg.Done()
		POSITION, VELOCITY := ecs.GetStorage2[Position, Velocity](pool)
		entities := POSITION.And(VELOCITY)
		POSITION.AcquireLockUnsafe()
		VELOCITY.AcquireLockUnsafe()
		defer POSITION.FreeLockUnsafe()
		defer VELOCITY.FreeLockUnsafe()
		for _, e := range entities {
			pos, vel := POSITION.GetPtrUnsafe(e), VELOCITY.GetPtrUnsafe(e)
			if pos == nil || vel == nil {
				continue
			}
			pos.X += vel.X
			pos.Y += vel.Y
		}
	}(pool)
	wg.Add(5)
	for i := 0; i < 5; i++ {
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
