package ecs

// add 2 components to an entity
func Add2[A any, B any](p *Pool, e Entity,
	c1 A, c2 B,
) {
	Add(p, e, c1)
	Add(p, e, c2)
}

// add 3 components to an entity
func Add3[A any, B any, C any](p *Pool, e Entity,
	c1 A, c2 B, c3 C,
) {
	Add(p, e, c1)
	Add(p, e, c2)
	Add(p, e, c3)
}
