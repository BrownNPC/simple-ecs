package ecs

// storage contains all components of a type
func GetStorage2[A any, B any](p *Pool) (*Storage[A], *Storage[B]) {
	return GetStorage[A](p),
		GetStorage[B](p)
}

// storage contains all components of a type
func GetStorage3[A any, B any, C any](p *Pool) (
	*Storage[A], *Storage[B],
	*Storage[C],
) {
	a, b := GetStorage2[A, B](p)
	c := GetStorage[C](p)
	return a, b, c
}

// storage contains all components of a type
func GetStorage4[A any, B any,
	C any, D any](p *Pool) (
	*Storage[A], *Storage[B],
	*Storage[C], *Storage[D],
) {
	a, b, c := GetStorage3[A, B, C](p)
	d := GetStorage[D](p)
	return a, b, c, d
}

// storage contains all components of a type
func GetStorage5[A any, B any, C any,
	D any, E any](p *Pool) (
	*Storage[A], *Storage[B],
	*Storage[C], *Storage[D],
	*Storage[E],
) {
	a, b, c, d := GetStorage4[A, B, C, D](p)
	e := GetStorage[E](p)
	return a, b, c, d, e
}

// storage contains all components of a type
func GetStorage6[A any, B any, C any,
	D any, E any, F any](p *Pool) (
	*Storage[A], *Storage[B],
	*Storage[C], *Storage[D],
	*Storage[E], *Storage[F]) {
	a, b, c, d, e := GetStorage5[A, B, C, D, E](p)
	f := GetStorage[F](p)
	return a, b, c, d, e, f
}

// storage contains all components of a type
func GetStorage7[A any, B any, C any,
	D any, E any, F any, G any](p *Pool) (
	*Storage[A], *Storage[B],
	*Storage[C], *Storage[D],
	*Storage[E], *Storage[F],
	*Storage[G],
) {
	a, b, c, d, e, f := GetStorage6[A, B, C, D, E, F](p)
	g := GetStorage[G](p)
	return a, b, c, d, e, f, g
}

// storage contains all components of a type
func GetStorage8[A any, B any, C any,
	D any, E any, F any, G any, H any](p *Pool) (
	*Storage[A], *Storage[B],
	*Storage[C], *Storage[D],
	*Storage[E], *Storage[F],
	*Storage[G], *Storage[H],
) {
	a, b, c, d, e, f, g := GetStorage7[A, B, C, D, E, F, G](p)
	h := GetStorage[H](p)
	return a, b, c, d, e, f, g, h
}

// storage contains all components of a type
func GetStorage9[A any, B any, C any,
	D any, E any, F any, G any, H any, I any](p *Pool) (
	*Storage[A], *Storage[B],
	*Storage[C], *Storage[D],
	*Storage[E], *Storage[F],
	*Storage[G], *Storage[H],
	*Storage[I],
) {
	a, b, c, d, e, f, g, h := GetStorage8[A, B, C, D, E, F, G, H](p)
	i := GetStorage[I](p)
	return a, b, c, d, e, f, g, h, i
}
