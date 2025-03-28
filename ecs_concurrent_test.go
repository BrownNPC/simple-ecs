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

func TestConcurrentUnsafeApi(t *testing.T) {
	pool := ecs.New(1000)
	type Position struct{ X, Y float32 }
	type Velocity struct{ X, Y float32 }
	var wg sync.WaitGroup
	wg.Add(50)
	go func(pool *ecs.Pool) {
		for i := 0; i < 50; i++ {
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
	wg.Add(1)
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
			if pos==nil || vel ==nil{
				continue
			}
			pos.X += vel.X
			pos.Y += vel.Y
		}
	}(pool)
	wg.Add(50)
	for i := 0; i < 50; i++ {
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
