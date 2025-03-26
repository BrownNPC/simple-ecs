package ecs

// add 2 components to an entity
// automatically register component if ecs.AutoRegisterComponents
// is true (default)
// This is just a wrapper arround calling ecs.Add multiple times
func Add2[A any, B any](p *Pool, e Entity,
	c1 A, c2 B,
) {
	Add(p, e, c1)
	Add(p, e, c2)
}

// add 3 components to an entity
// automatically register component if ecs.AutoRegisterComponents
// is true (default)
// This is just a wrapper arround calling ecs.Add multiple times
func Add3[A any, B any, C any](p *Pool, e Entity,
	c1 A, c2 B, c3 C,
) {
	Add2(p, e, c1, c2)
	Add(p, e, c3)
}

