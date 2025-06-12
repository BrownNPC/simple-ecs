package ecs

import (
	"sync"
)

// An entity is just an integer.
// But if you are storing entities within components, then do not forget to also
// store their generation, and verify them using [IsAliveWithGeneration] before looping
// over said components
type Entity = uint32
type Generation = uint32

// A storage holds a slice of components
type Storage[Component any] struct {
	ID         int
	components []Component
	b          *bitSet
	//[]Entity from Pool
	// used for queries
	parentPoolEntities *sync.Pool
}

// The pool holds Component slices within storages and tracks entity lifetimes
type Pool struct {
	TotalEntities uint32
	capacity      uint32

	mu          sync.RWMutex
	storages    map[any]storage // mapped from nilptr of Component to Storage[Component]
	allStorages []storage       // used for quickly killing entities, faster than iterating a map

	reusableIDs        []uint32
	generations        []Generation // incremented after every entity is killed. Used to prevent errors when we reuse an entity that the user was storing
	entityActiveStatus *bitSet      // track which entities are alive= w

	// passed to storages
	poolEntititySlices sync.Pool // pool of []Entity, used for queries
}

func New(capacity uint32) (p *Pool) {
	capacity++ //index 0 is unused so we should allocated 1 extra entity
	p = &Pool{capacity: capacity}
	p.entityActiveStatus = newBitset(capacity)
	p.storages = make(map[any]storage)
	p.reusableIDs = make([]Entity, 0, capacity)
	p.generations = make([]Generation, capacity)
	p.poolEntititySlices = sync.Pool{
		New: func() any {
			return make([]Entity, p.capacity)
		},
	}
	return p
}

// recycle a dead entity id, or create a new one
func NewEntity(p *Pool) Entity {
	p.mu.Lock()
	defer p.mu.Unlock()
	reusableLen := len(p.reusableIDs)
	if reusableLen > 0 { // reuse
		id := p.reusableIDs[reusableLen-1]
		p.reusableIDs = p.reusableIDs[:reusableLen-1]
		p.entityActiveStatus.Set(id)
		return id
	}
	// new entity
	// entity 0 is unused
	p.TotalEntities++
	id := p.TotalEntities
	p.entityActiveStatus.Set(id)
	return id
}

// You only need this if you are storing Entities within components
func GetGeneration(p *Pool, e Entity) uint32 {
	return p.generations[e]
}

// Give an entity back to the pool, allowing recycling
func Kill(p *Pool, entities ...Entity) {
	p.mu.Lock()
	var toClear []Entity
	for _, e := range entities {
		if !p.entityActiveStatus.Get(e) {
			continue
		}
		p.entityActiveStatus.Clear(e)
		p.generations[e]++
		p.reusableIDs = append(p.reusableIDs, e)
		toClear = append(toClear, e)
	}
	p.mu.Unlock()

	for _, e := range toClear {
		for _, st := range p.allStorages {
			if st.bits().Get(e) { // skip zeroing if no bit
				st.clear(e)
			}
		}
	}
}

// Check if an entity is alive.
//
// This is not enough if you are storing entities within components
func IsAlive(p *Pool, e Entity) bool {
	return p.entityActiveStatus.Get(e)
}

// Check if internal generation for this Entity matches the generation you are storing
//
// You only need this if you are storing entities
func IsAliveWithGeneration(p *Pool, e Entity, generation Generation) bool {
	return generation == p.generations[e]
}

// Get a component storage, allocate it if not already
func GetStorage[Component any](p *Pool) *Storage[Component] {
	nilptr := (*Component)(nil) // Key: nil pointer to Component

	p.mu.RLock()
	st, ok := p.storages[nilptr]
	p.mu.RUnlock()
	if ok {
		return st.(*Storage[Component])
	}

	// Not found, acquire write lock
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double check after acquiring write lock
	if st, ok := p.storages[nilptr]; ok {
		return st.(*Storage[Component])
	}

	// Still not present, safe to create
	newSt := newStorage[Component](p.capacity)
	// pass []Entity, used for queries
	newSt.parentPoolEntities = &p.poolEntititySlices
	p.storages[nilptr] = newSt
	p.allStorages = append(p.allStorages, newSt)
	return newSt
}

// Add a component to an entity.
//
// adding to a dead entity is a no-op. but note that this does not validate generations
func Add[Component any](p *Pool, e Entity, c Component) {
	if !IsAlive(p, e) {
		return
	}
	st := GetStorage[Component](p)
	st.bits().Set(e)
	st.Update(e, c)
}

// Remove a component from an entity
//
// removing from a dead entity is a no-op. but note that this does not validate generations
func Remove[Component any](p *Pool, e Entity) {
	if !IsAlive(p, e) {
		return
	}
	st := GetStorage[Component](p)
	st.clear(e)
}
