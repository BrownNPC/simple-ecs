package ecs

func newStorage[Component any](capacity uint32) (s *Storage[Component]) {
	return &Storage[Component]{
		components: make([]Component, capacity),
		b:          newBitset(capacity),
	}
}

func (s *Storage[Component]) EntityHasComponent(e Entity) bool {
	return s.b.Get(e)
}

func (s *Storage[Component]) bits() *bitSet { return s.b }

type storage interface {
	bits() *bitSet
	clear(Entity) // zero out the component for this entity
}

// zero out the components for this entity
// does nothing if the entity is alive
func (s *Storage[Component]) clear(e Entity) {
	s.bits().Clear(e)
	var zero Component
	s.components[e] = zero
}

// update the component of an entity. this does not check if the entity is alive
func (s *Storage[Component]) Update(e Entity, c Component) {
	s.components[e] = c
}

// get a copy of a component
// this does not check if the entity is alive
func (s *Storage[Component]) Get(e Entity) Component {
	return s.components[e]
}

// All entities that have this component
func (s *Storage[Component]) All() []Entity {
	return s.b.ActiveIDs()
}

// All entities that have this component and the other components
func (s *Storage[Component]) And(others ...storage) []Entity {
	bits := s.b.Clone()
	defer bits.Release()
	for _, s2 := range others {
		bits.And(s2.bits())
	}
	return bits.ActiveIDs()
}

// All entities that have this component but not the other components
func (s *Storage[Component]) ButNot(others ...storage) []Entity {
	bits := s.b.Clone()
	defer bits.Release()
	for _, s2 := range others {
		bits.AndNot(s2.bits())
	}
	return bits.ActiveIDs()
}

// All entities that have either components
func (s *Storage[Component]) Or(others ...storage) []Entity {
	bits := s.b.Clone()
	defer bits.Release()
	for _, s2 := range others {
		bits.Or(s2.bits())
	}
	return bits.ActiveIDs()
}
