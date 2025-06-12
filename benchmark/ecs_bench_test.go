package ecs_test

import (
	"testing"

	ecs "github.com/BrownNPC/simple-ecs"
)

func BenchmarkQuery(b *testing.B) {
	type Vec2 struct {
		X, Y float64
	}
	type Position Vec2
	type Velocity Vec2

	p := ecs.New(uint32(50_000))
	for range 50_000 {
		e := ecs.NewEntity(p)
		ecs.Add2(p, e, Position{}, Velocity{1, 1})
	}
	// POSITION := ecs.GetStorage[Position](p)
	// VELOCITY := ecs.GetStorage[Velocity](p)
	for b.Loop() {
		// POSITION.All()
	}
}
