# Simple-ECS
dead-simple library for writing
game systems in Go

### Simple-ECS Features:
- Easy syntax / api
- Good perfomance
- Easy to understand code (300 lines*)
- Low level (implement what you need)

### Get started
 - I am a beginner
 - I know what ECS is



### What is ECS? (and why you should use it)
I recommend you watch the first couple minutes of [this](https://youtu.be/9LNgSDP1zrw?t=2m40s)
video. feel free to skip around
or watch at 2x speed



#### ECS is an alternative to inheritance.


Instead of creating game objects using Object Oriented Design where
things inherit from each other eg.
Player inherits from Actor, Actor inherits From Entity,
we
think about the **Data**. The goal is to
seperate the logic from the data.
This is known as
Data oriented design.

In this pattern, we have entities
which have components.
**Components are pure-data**
for example a position component
might look like this:
```go
type Position struct{
  X,Y float64
}
```
A health component might
just be an integer.

Using components you can make systems.
Systems are just normal functions that
modify components.

For example, you may have a movement system
that loops over all the entities that
have a Position and a Velocity component
adds the Velocity to the Position of the entity

```
func MovementSystem(entities []entity){
	for _ ent := range entities{
		ent.Position.X = ent.Velocity.X
		ent.Position.Y = ent.Velocity.Y
	}
}
```

### Why use ECS for writing game systems?
  Because Go does not have inheritance.
  The language prefers seperating data from
  logic.

#### How to use Simple ECS for writing systems
Before we jump into the example, understanding how
this library is implemented will help us learn it easily.

	The heart of this ECS is the memory pool
	Think of the pool like a database or a spreadsheet.
	On the Y axis (columns) there are arrays of components

	We use a struct called storage to hold the components arrays
	components can be any data type, but they cannot be interfaces
	These arrays are pre-allocated to a fixed size provided by the user

	an entity is just an index into these arrays
	So on the X axis there are entities which are just indexes
```go
// stores slice of components
type Storage[Component any] struct {
	// slice of components
	components     []Component
	// a bitset is used to store which
	//indexes are occupied by entities
	b   bitset.BitSet
}
```
	The storage struct also has a bitset (like an array of boleans)

	each bit in the bitset corresponds to an entity
	 the bitset is used for maintaining
	a record of which entity has the component the storage is storing

	The pool also has its own bitset that tracks which entities are alive

		there is also a map from entities to a slice of component storages

		we update this map when an entity has a component added to it

		we use this map to go into every storage and zero out the component
		when an entity is killed.
    you dont need to worry about how the pool works

Now here is an example:
```go
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
	// the pool can hold 1000 entities
	var pool = ecs.New(1000)
	// create 1000 entities
	for range 1000 {
		// entities (which are just ids)
		// should only be created using the pool
		var e = ecs.NewEntity(pool)
		// add position and
		// velocity components to the entity
		ecs.Add2(pool, e,
			Position{},
			Velocity{
				X: rand.Float64(),
				Y: rand.Float64(),
			})
	}
	// run movement system 60 times
	for range 60 {
		MovementSystem(pool, 1.0/60)
	}
}

// a system is a regular function that
// operates on the components
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
```

### Motivation + Opinion:
  The other ECS libraries seem
  to focus on having the best
  possible performance,
  sometimes sacrificing a
  simpler syntax. They also provide features
  I dont need.
  Some devs put
  restrictions on the use of
  runtime reflection for negligible
  performance gains. And these libraries had
  many ways to do
  the same thing. (eg. Arche has 2 apis)

  This is just my opinion but most
  games that are made using Go should
  not care about microseconds worth
  of performance gains. As the main reason
  to pick Go over C++, C#, Java or Rust is
  because of Go's simplicity.
  
  Also no hate or anything of that sort
  is intended towards any developer's work.
  Everyone has their own reasons for writing
  their own code. I am not claiming that
  this is the best ECS for Go. I am only claiming
  that it has a simple API,
  but that could be subjective.

### Acknowledgements
  Donburi is another library that
  implements ECS with a simple API.
  But in my opinion this library is
  simpler.
