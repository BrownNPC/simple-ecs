package ecs

// allocate a new component storage
//
//	if you wish to register interfaces,
//	wrap them in a struct
func Register2[A any, B any](p *Pool) {
	Register[A](p)
	Register[B](p)
}

// allocate a new component storage
//
//	if you wish to register interfaces,
//	wrap them in a struct
func Register3[A any, B any, C any](p *Pool) {
	Register2[A, B](p)
	Register[C](p)
}

// allocate a new component storage
//
//	if you wish to register interfaces,
//	wrap them in a struct
func Register4[A any, B any, C any, D any](p *Pool) {
	Register3[A, B, C](p)
	Register[D](p)
}

// allocate a new component storage
//
//	if you wish to register interfaces,
//	wrap them in a struct
func Register5[A any, B any, C any, D any, E any](p *Pool) {
	Register4[A, B, C, D](p)
	Register[E](p)
}
