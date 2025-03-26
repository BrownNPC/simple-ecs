package ecs

// simple-ecs copyright Omer Farooqui all rights reserved..
// this code is MIT licensed.
import (
	"fmt"

	bitset "github.com/BrownNPC/simple-ecs/internal"
)

// when this is false, the library will panic if components
// are not registered before use.
var AutoRegisterComponents = true

// Entity is an id that has components
//
//	please only create entities using the pool
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
	components     []Component
	availableIndex int
	// a bitset is used to store which
	//indexes are occupied by entities
	b        bitset.BitSet
	mut      CustomRWMutex
}
type _Storage interface {
	delete(e Entity)
	getBitset() *bitset.BitSet
}

func (s *Storage[T]) delete(e Entity) {
	s.mut.Lock()
	var zero T
	s.components[e] = zero
	s.b.Unset(uint(e))
	s.mut.Unlock()
}
func (s *Storage[T]) getBitset() *bitset.BitSet {
	return &s.b
}

// takes in other storages and returns
// entities that exist in all of them
//
//	in simple terms:
//	 entities that have all of these components
//
// passing in nil or nothing will return the entities with the component this storage stores
func (s *Storage[T]) And(storages ..._Storage) []Entity {
	mask := s.b.Clone()
	s.mut.RLock()
	defer s.mut.RUnlock()
	if len(storages) > 0 {
		for _, s := range storages {
			if s != nil {
				mask.And(s.getBitset())
			}
		}
	}
	return bitset.ActiveIndices[Entity](&mask)
}

// takes in other storages and returns
// entities that exist in this storage but
// not in the storages passed in
//
//		in simple terms:
//		 entities that have this component
//	  but not the other ones
//
// passing in nil or nothing will return the entities with the component this storage stores
func (s *Storage[T]) ButNot(storages ..._Storage) []Entity {
	mask := s.b.Clone()
	s.mut.RLock()
	defer s.mut.RUnlock()
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
	if !st.EntityHasComponent(e) {
		panic("Tried updating entity's component, but the entity does not have this component, add it first using ecs.Add")
	}
	st.components[e] = component
}

// check if an Entity has a component
func (st *Storage[T]) EntityHasComponent(e Entity) bool {
	//by looking at the bitset of storage
	return st.b.IsSet(uint(e))
}

// get a copy of an entity's component
func (st *Storage[T]) Get(e Entity) T {
	st.mut.RLock()
	c := st.components[e]
	st.mut.RUnlock()
	return c
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
	length           int
	freelistMu       CustomRWMutex
	componentsUsedMu CustomRWMutex
}

// make a new memory pool of components
//
//	size is the number of entities
//	worth of memory to pre-allocate
//
//	the memory usage of the pool depends on
//	how many components your game has and how many
//	entities you allocate
//
// the pool will NOT grow dynamically
func New(size int) *Pool {
	return &Pool{
		stores:         make(map[any]_Storage),
		componentsUsed: make(map[Entity][]_Storage),
		size:           size,
	}
}

// Get an entity
func NewEntity(p *Pool) Entity {
	p.freelistMu.Lock()
	defer p.freelistMu.Unlock()
	if len(p.freeList) == 0 {
		if p.length >= p.size {
			panic("Entity limit exceeded. please initialize more entities by increasing the number you passed to ecs.New()")
		}
		e := Entity(p.length)
		p.aliveEntities.Set(uint(e))
		p.length++
		return e
	}
	// recycle an entity
	var newEntity = p.freeList[0]
	p.freeList = p.freeList[1:]
	return newEntity
}

// give entities back to the pool
func Kill(p *Pool, entities ...Entity) {
	for _, e := range entities {
		p.aliveEntities.Unset(uint(e))
		//mark the entity as available
		p.freelistMu.Lock()
		p.freeList = append(p.freeList, e)
		p.freelistMu.Unlock()
		p.componentsUsedMu.RLock()
		var storagesUsed []_Storage = p.componentsUsed[e]
		p.componentsUsedMu.RUnlock()
		for _, store := range storagesUsed {
			//zero out the component for this entity
			store.delete(e)
		}
		p.componentsUsedMu.Lock()
		// entity no longer has these components
		// set slice length to 0
		p.componentsUsed[e] = p.componentsUsed[e][:0]
		p.componentsUsedMu.Unlock()
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
	var nilptr *T
	_, ok := pool.stores[nilptr]
	if !ok {
		pool.componentsUsedMu.Lock()
		defer pool.componentsUsedMu.Unlock()
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
	st := registerAndGetStorage[T](pool)
	if st.EntityHasComponent(e) {
		return
	}
	st.b.Set(uint(e))
	st.Update(e, component)
	pool.componentsUsedMu.Lock()
	// append reflect.Type of the new component
	pool.componentsUsed[e] =
		append(pool.componentsUsed[e], st)
	pool.componentsUsedMu.Unlock()
}

// remove a component from an entity
func Remove[T any](pool *Pool, e Entity) {
	st := registerAndGetStorage[T](pool)
	if !st.EntityHasComponent(e) {
		return
	}
	st.delete(e)
	pool.componentsUsedMu.RLock()
	var s []_Storage = pool.componentsUsed[e]
	pool.componentsUsedMu.RUnlock()

	pool.componentsUsedMu.Lock()
	defer pool.componentsUsedMu.Unlock()
	store := (_Storage)(st)
	// iterate in reverse
	// incase the component was added recently
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == store {
			// move the _Storage to the end of the slice and
			// shrink the slice by one
			temp := s[len(s)-1]
			s[len(s)-1] = s[i]
			s[i] = temp
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
	st := registerAndGetStorage[T](pool)
	return st.EntityHasComponent(e)
}

// storage contains all components of a type
func GetStorage[A any](pool *Pool) *Storage[A] {
	st := registerAndGetStorage[A](pool)
	return st

}

// same as public register but also gives the storage
// it will not allocate a new storage if it already exists
// this will use the pool's mutexes appropriately
func registerAndGetStorage[T any](pool *Pool) *Storage[T] {
	var nilptr *T
	pool.componentsUsedMu.RLock()
	st, ok := pool.stores[nilptr]
	pool.componentsUsedMu.RUnlock()
	if ok {
		return st.(*Storage[T])
	} else if AutoRegisterComponents {
		// allocate storage
		var st = newStorage[T](pool.size)
		pool.componentsUsedMu.Lock()
		pool.stores[nilptr] = st
		pool.componentsUsedMu.Unlock()
		return st
	}
	var zero T
	panic(fmt.Sprintf("Component of type %T was not registered", zero))
}
