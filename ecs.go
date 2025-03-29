package ecs

// simple-ecs copyright Omer Farooqui all rights reserved..
// this code is MIT licensed.
import (
	"fmt"
	"sync"

	bitset "github.com/BrownNPC/simple-ecs/internal"
)

// when this is false, the library will panic if components
// are not registered before use.
var AutoRegisterComponents = true

// Entity is an id that has components,
// they can only be created using ecs.NewEntity()
type Entity uint32

// stores slice of components
/*
We use a struct called storage to hold the components array.
this struct also has a bitset where
each bit in the bitset corresponds to an entity.
the bitset is used for maintaining
a record of which entity has the component the storage is storing
*/
type Storage[Component any] struct {
	// slice of components
	components []Component
	// a bitset is used to store which
	//indexes are occupied by entities
	b   bitset.BitSet
	mut sync.Mutex
}
type _Storage interface {
	delete(e Entity)
	getBitset() *bitset.BitSet
	lock()
	unlock()
}

func (s *Storage[T]) delete(e Entity) {
	s.mut.Lock()
	defer s.mut.Unlock()
	var zero T
	s.components[e] = zero
}
func (s *Storage[T]) getBitset() *bitset.BitSet {
	return &s.b
}
func (s *Storage[Component]) lock() {
	s.mut.Lock()
}
func (s *Storage[Component]) unlock() {
	s.mut.Unlock()
}

// takes in other storages and returns
// entities that exist in all of them
//
//	in simple terms:
//	 entities that have all of these components
//
// passing in nil or nothing will return entities that have this storage's component
func (s *Storage[T]) And(storages ..._Storage) []Entity {
	s.mut.Lock()
	defer s.mut.Unlock()
	mask := s.b.Clone()
	s.b.Mu.Lock()
	defer s.b.Mu.Unlock()
	if len(storages) > 0 {
		for _, otherSt := range storages {
			if otherSt != nil {
				otherSt.lock()
				defer otherSt.unlock()
				mask.And(otherSt.getBitset())
			}
		}
	}
	return bitset.ActiveIndices[Entity](&mask)
}

// takes in other storages and returns
// entities that exist in this storage but
// not in the storages passed in
//
//	 in simple terms:
//		entities that have this component
//		but not the other ones
//
// passing in nil or nothing will return the entities with the component this storage stores
func (s *Storage[T]) ButNot(storages ..._Storage) []Entity {
	s.mut.Lock()
	defer s.mut.Unlock()
	mask := s.b.Clone()
	for _, s := range storages {
		if s != nil {
			mask.AndNot(s.getBitset())
		}
	}
	return bitset.ActiveIndices[Entity](&mask)
}

// set an entity's component
// this will panic if the entity does not have this component
func (st *Storage[T]) Update(e Entity, component T) {
	st.mut.Lock()
	defer st.mut.Unlock()
	if !st.b.IsSet(uint(e)) {
		return
	}
	st.components[e] = component
}

// check if an Entity has a component
func (st *Storage[T]) EntityHasComponent(e Entity) bool {
	//by looking at the bitset of storage
	st.mut.Lock()
	defer st.mut.Unlock()
	return st.b.IsSet(uint(e))
}

// get a copy of an entity's component
// You can then update the entity using
// Storage[T].Update()
// if the entity is dead, or does not have this component
// then the returned value will be the zero value of the component
func (s *Storage[T]) Get(e Entity) (component T) {
	s.mut.Lock()
	defer s.mut.Unlock()
	if !s.b.IsSet(uint(e)) {
		return component
	}
	return s.components[e]
}

func newStorage[T any](size int) *Storage[T] {
	return &Storage[T]{
		components: make([]T, size),
	}
}

// A pool holds component storages and does book keeping of
// alive and dead entities
/*
	Think of the pool like a database.
	On the Y axis (columns) there are arrays of components
	components can be any data type
	These arrays are pre-allocated to a fixed size provided by the user

	an entity is just an index into these arrays
	So on the X axis there are entities which are just indexes

*/
type Pool struct {
	// we map pointer to type T to the storage of T
	// *T -> Storage[T]
	stores map[any]_Storage
	// used to track components an entity has
	// we zero out the components when an entity dies
	// and update this map when a component is added to an entity
	// this is only used for internal book keeping of
	// dead and alive entities
	componentsUsed map[Entity][]_Storage
	// which entities are alive
	aliveEntities bitset.BitSet
	// recycle killed entities
	freeList []Entity
	// no. of entities to pre-allocate / max entity count
	size int
	//how many entities are alive
	length int
	// pool will not autoGrow storages if this is false
	autoGrow bool
	// generations track how many times an entity was recycled
	generations []uint64
	mut         sync.Mutex
}

// make a new memory pool of components.
//
// size is the number of entities
// worth of memory to pre-allocate
// and the maximum number of entities if pool.EnableResize is not called
//
// the memory usage of the pool depends on
// how many components your game has and how many
// entities you allocate
func New(size int) *Pool {
	return &Pool{
		stores:         make(map[any]_Storage),
		componentsUsed: make(map[Entity][]_Storage),
		generations:    make([]uint64, size),
		size:           size,
	}
}

// increase the size of the storages when pool runs out of available entities
// resizing has overhead, you should use a fixed size for the pool instead.
func (p *Pool) EnableGrowing() *Pool {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.autoGrow = true
	return p
}

// Get an entity
// this will panic if pool does not have entities available
func NewEntity(p *Pool) Entity {
	p.mut.Lock()
	defer p.mut.Unlock()
	// if no entities are available for recycling
	if len(p.freeList) == 0 {
		if !p.autoGrow && p.length >= p.size {

			msg := fmt.Sprintf("Entity limit exceeded. please initialize more entities by increasing the number you passed to ecs.New(). OR enable growing of the pool by calling ecs.New(x).EnableGrowing()\nGiven size: %d\n Entity: %d", p.size, p.length+1)
			panic(msg)
		}
		e := Entity(p.length)
		p.aliveEntities.Set(uint(e))
		p.length++
		return e
	}
	// recycle an entity
	var newEntity = p.freeList[0]
	p.freeList = p.freeList[1:]
	//update generation of this entity
	if int(newEntity) > int(len(p.generations)) {
		//resize generations slice
		//potentially allowing for +1000 entities
		p.size = max(p.size, p.size+1000) // we check this because ecs.Add modifies Pool.size
		newGens := make([]uint64, p.size)
		copy(newGens, p.generations)
		p.generations = newGens
	}
	// finally update generation
	p.generations[newEntity] += 1
	var storagesUsed []_Storage = p.componentsUsed[newEntity]
	for _, store := range storagesUsed {
		//zero out the component for this entity
		store.delete(newEntity)
	}
	// entity no longer has these components
	// set slice length to 0
	p.componentsUsed[newEntity] = p.componentsUsed[newEntity][:0]

	return newEntity
}

// give entities back to the pool
func Kill(p *Pool, entities ...Entity) {
	p.mut.Lock()
	defer p.mut.Unlock()
	for _, e := range entities {
		p.aliveEntities.Unset(uint(e))
		//mark the entity as available
		p.freeList = append(p.freeList, e)
		var storagesUsed []_Storage = p.componentsUsed[e]
		for _, store := range storagesUsed {
			//mark as dead but dont zero out the component for this entity
			store.lock()
			store.getBitset().Unset(uint(e))
			store.unlock()
		}
	}
}

// allocate a new component storage
//
//	will panic if you register components twice
//
// Components cannot be aliases eg.
//
//	type Position Vec2 // correct
//	type Position = Vec2 // incorrect
func Register[T any](pool *Pool) {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	var nilptr *T
	_, ok := pool.stores[nilptr]
	if !ok {
		pool.stores[nilptr] = newStorage[T](pool.size)
		return
	}
	panic(fmt.Sprintln("Component", nilptr, `is already registered 
If you are using type aliases
use concrete types instead
Example:
type Position Vec2 // correct
type Position = Vec2 // incorrect `))
}

// add a component to an entity
// automatically register component if ecs.AutoRegisterComponents
// is true (default)
func Add[T any](pool *Pool, e Entity, component T) {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	if !IsAlive(pool, e) {
		return
	}
	st := registerAndGetStorage[T](pool)
	st.mut.Lock()
	defer st.mut.Unlock()
	if st.b.IsSet(uint(e)) {
		return
	}
	capacity := cap(st.components)
	if capacity <= int(e) {
		// resize, potentially allowing for 1000 more entities
		pool.size = max(pool.size, pool.size+1000)
		newSt := newStorage[T](pool.size)
		// the larger array is now the new array for the storage
		copy(newSt.components, st.components)
		st.components = newSt.components

	}
	st.b.Set(uint(e))
	st.components[e] = component
	pool.componentsUsed[e] =
		append(pool.componentsUsed[e], st)
}

// remove a component from an entity
func Remove[T any](pool *Pool, e Entity) {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	st := registerAndGetStorage[T](pool)
	if !st.b.IsSet(uint(e)) {
		return
	}
	st.delete(e)
	var s []_Storage = pool.componentsUsed[e]

	store := (_Storage)(st)
	// iterate in reverse
	// incase the component was added recently
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == store {
			if len(s) == 1 {
				pool.componentsUsed[e] = s[:0]
				return
			}
			// move the _Storage to the end of the slice and
			// shrink the slice by one
			last := s[len(s)-1]
			s[len(s)-1] = s[i]
			s[i] = last
			// "delete" last element
			pool.componentsUsed[e] = s[0 : len(s)-2]
			return
		}
	}
}

// check if an entity has a component
// shorthand for
//
//	POSITION := ecs.GetStorage[Position](pool)
//	POSITION.EntityHasComponent(e)
func Has[T any](pool *Pool, e Entity) bool {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	st := registerAndGetStorage[T](pool)
	return st.b.IsSet(uint(e))
}

// Check if an entity is alive
func IsAlive(pool *Pool, e Entity) bool {
	return pool.aliveEntities.IsSet(uint(e))
}

// storage contains all components of a type
func GetStorage[A any](pool *Pool) *Storage[A] {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	st := registerAndGetStorage[A](pool)
	return st
}

// the generation is the number of times the entity has been recycled
func GetGeneration(pool *Pool, e Entity) uint64 {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	if int(e) > len(pool.generations) {
		return 0
	}
	return pool.generations[e]
}

// same as public register but also gives the storage
// it will not allocate a new storage if it already exists
// this will use the pool's mutexes appropriately
func registerAndGetStorage[T any](pool *Pool) *Storage[T] {
	var nilptr *T
	st, ok := pool.stores[nilptr]
	if ok {
		return st.(*Storage[T])
	} else if AutoRegisterComponents {
		// allocate storage
		var st = newStorage[T](pool.size)
		pool.stores[nilptr] = st
		return st
	}
	var zero T
	panic(fmt.Sprintf("Component of type %T was not registered", zero))
}
