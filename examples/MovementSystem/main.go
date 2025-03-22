package main

import (
	ecs "github.com/BrownNPC/simple-ecs"
	"math/rand"
)

// Define component types
type Vec2 struct {
	X, Y float64
}

// components need to be concrete types
// type Position = Vec2 // is incorrect
type Position Vec2
type Velocity Vec2

func main() {
	// create a memory pool of component arrays
	var p = ecs.New(1000)
	// create 1000 entities
	for range 1000 {
		var e = ecs.NewEntity(p)
		// add position and
		// velocity components
		ecs.Add2(p, e,
			Position{},
			Velocity{
				X: rand.Float64(),
				Y: rand.Float64(),
			})
	}
	// run movement system 60 times
	for range 60 {
		MovementSystem(p, 1.0/60)
	}
}

// a system is a regular function
func MovementSystem(p *ecs.Pool,
	deltaTime float64,
) {
	// a storage holds a slice (array) of components
	POSITION, VELOCITY :=
		ecs.GetStorage2[
			Position,
			Velocity,
		](p)
	// get entities (id/index) that have
	// a position and velocity component
	for _, ent := range POSITION.Matches(VELOCITY) {
		// use the entity to index the
		// position and velocity slices
		pos, vel :=
			POSITION.Get(ent),
			VELOCITY.Get(ent)
		pos.X += vel.X * deltaTime
		pos.Y += vel.Y * deltaTime
		// update position of entity
		POSITION.Update(ent, pos)
	}
}
