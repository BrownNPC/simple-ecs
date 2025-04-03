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

// add 4 components to an entity
// automatically register component if ecs.AutoRegisterComponents
// is true (default)
// This is just a wrapper arround calling ecs.Add multiple times
func Add4[A any, B any, C any, D any](p *Pool, e Entity,
	c1 A, c2 B, c3 C, c4 D,
) {
	Add3(p, e, c1, c2, c3)
	Add(p, e, c4)
}

// add 5 components to an entity
// automatically register component if ecs.AutoRegisterComponents
// is true (default)
// This is just a wrapper arround calling ecs.Add multiple times
func Add5[A any, B any, C any, D any, E any](p *Pool, e Entity,
	c1 A, c2 B, c3 C, c4 D, c5 E,
) {
	Add4(p, e, c1, c2, c3, c4)
	Add(p, e, c5)
}

// add 6 components to an entity
// automatically register component if ecs.AutoRegisterComponents
// is true (default)
// This is just a wrapper arround calling ecs.Add multiple times
func Add6[A any, B any, C any, D any, E any, F any](p *Pool, e Entity,
	c1 A, c2 B, c3 C, c4 D, c5 E, c6 F,
) {
	Add5(p, e, c1, c2, c3, c4, c5)
	Add(p, e, c6)
}

// add 7 components to an entity
// automatically register component if ecs.AutoRegisterComponents
// is true (default)
// This is just a wrapper arround calling ecs.Add multiple times
func Add7[A any, B any, C any, D any, E any, F any, G any](p *Pool, e Entity,
	c1 A, c2 B, c3 C, c4 D, c5 E, c6 F, c7 G,
) {
	Add6(p, e, c1, c2, c3, c4, c5, c6)
	Add(p, e, c7)
}

// add 8 components to an entity
// automatically register component if ecs.AutoRegisterComponents
// is true (default)
// This is just a wrapper arround calling ecs.Add multiple times
func Add8[A any, B any, C any, D any, E any, F any, G any, H any](p *Pool, e Entity,
	c1 A, c2 B, c3 C, c4 D, c5 E, c6 F, c7 G, c8 H,
) {
	Add7(p, e, c1, c2, c3, c4, c5, c6, c7)
	Add(p, e, c8)
}

// add 9 components to an entity
// automatically register component if ecs.AutoRegisterComponents
// is true (default)
// This is just a wrapper arround calling ecs.Add multiple times
func Add9[A any, B any, C any, D any, E any, F any, G any, H any, I any](p *Pool, e Entity,
	c1 A, c2 B, c3 C, c4 D, c5 E, c6 F, c7 G, c8 H, c9 I,
) {
	Add8(p, e, c1, c2, c3, c4, c5, c6, c7, c8)
	Add(p, e, c9)
}
