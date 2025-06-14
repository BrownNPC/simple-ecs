# Simple-ECS
dead-simple library for writing
game systems in Go

### Install
```go get github.com/BrownNPC/simple-ecs```

#### Documentation
[GoDoc can be found here](https://pkg.go.dev/github.com/BrownNPC/simple-ecs#pkg-variables)

[Jump to Example](https://github.com/BrownNPC/simple-ecs?tab=readme-ov-file#now-here-is-an-example)

### Simple-ECS Features:
- Easy syntax / api
- Good perfomance!
- Low level (implement what you need)
- No Dependencies on other libraries


### What is ECS? (and why you should use it)
I recommend you watch [this](https://www.youtube.com/watch?v=JxI3Eu5DPwE)
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
have a Position and a Velocity component,
and then
adds the Velocity to the Position of the entity

```go
func MovementSystem(entities []entity){
	for _ ent := range entities{
		ent.Position.X += ent.Velocity.X
		ent.Position.Y += ent.Velocity.Y
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

We use a struct called storage to hold the components arrays.

components can be any data type

These arrays are pre-allocated to a fixed size provided by the user

An entity is just an index into these arrays

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

Each bit in the bitset corresponds to an entity.
By setting the bit on the bitset, we can keep
a record of whether an entity has the component added to it.

The pool also has its own bitset that tracks which entities are alive
you dont need to worry about how the pool works, just know that the
pool is responsible for creating and deleting entities.

## Now here is an example:
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
type Position Vec2
type Velocity Vec2

func main() {
	// create a memory pool of component arrays
	// the pool can hold a maximum of 1000 alive entities
	var pool = ecs.New(1000)
	// create 1000 entities
	for range 1000 {
		// entities (which are just ids)
		// should only be created using the pool
		var ent = ecs.NewEntity(pool)
		// add position and
		// velocity components to the entity
		ecs.Add2(pool, ent,
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
func MovementSystem(p *ecs.Pool, deltaTime float64) {
	// a storage holds a slice (array) of components
	POSITION, VELOCITY :=
		ecs.GetStorage2[ // helper function so you dont have to call GetStorage twice
			Position,
			Velocity,
		](p)
	// get entities (id/index) that have
	// a position and velocity component
	for _, ent := range POSITION.And(VELOCITY) {
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

### When to not use an ECS
You dont need ECS if your game is going to be very simple
like pong or flappy bird. But if you are making eg. "flappy bird with guns"
then ECS makes sense.
But even if you are using ECS;

**Everything in your game does not *need* to be an entity.**
For example. If you are making "Chess with magic spells",
you might want to represent the board state using a Chess board struct (object)
and the pieces would probably be entities that have components. and you would
probably have systems for animations, the timer, magic spells, and maybe
checking if a piece can move to a square etc.

Your user interface (UI) would probably also not benefit from being entities.

### Motivation:
  The other ECS libraries seem
  to focus on having the best
  possible performance,
  sometimes sacrificing a
  simpler syntax. They also provide features
  I dont need.
  And these libraries had
  many ways to do
  the same thing. (eg. Arche has 2 apis)

	I made this library to have less features,
	and sacrifice a little performance
	for more simplicity.
	Note: if you care about every nanosecond of performance, dont use my library.

### Acknowledgements
  Donburi is another library that
  implements ECS with a simple API.


## Running tests
`go test -count 10 -race ./...`
