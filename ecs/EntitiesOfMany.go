package ecs

// return entities which have all these components
//
//	shorthand for
//	a,b := ecs.StorageOf2[A,B](pool)
//	entities := a.Matches(b)
func EntitiesOf2[A any, B any](pool *Pool) []Entity {
	a, b := GetStorage2[A, B](pool)
	return a.Matches(b)
}

// return entities which have all these components
//
//	shorthand for eg.
//	a,b := ecs.StorageOf2[A,B](pool)
//	entities := a.Matches(b)
func EntitiesOf3[A any, B any, C any](pool *Pool) []Entity {
	a, b, c := GetStorage3[A, B, C](pool)
	return a.Matches(b, c)
}

// return entities which have all these components
//
//	shorthand for eg.
//	a,b := ecs.StorageOf2[A,B](pool)
//	entities := a.Matches(b)
func EntitiesOf4[A any, B any, C any, D any](pool *Pool) []Entity {
	a, b, c, d := GetStorage4[A, B, C, D](pool)
	return a.Matches(b, c, d)
}

// return entities which have all these components
//
//	shorthand for eg.
//	a,b := ecs.StorageOf2[A,B](pool)
//	entities := a.Matches(b)
func EntitiesOf5[A any, B any, C any, D any, E any](pool *Pool) []Entity {
	a, b, c, d, e := GetStorage5[A, B, C, D, E](pool)
	return a.Matches(b, c, d, e)
}
