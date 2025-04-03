package ecs

// simple-ecs copyright Omer Farooqui all rights reserved..
// this code is MIT licensed.
import (
	"fmt"
	"slices"
	"sync"

	bitset "github.com/BrownNPC/simple-ecs/internal"
)

// when this is false, the library will panic if components
// are not registered before use.
var AutoRegisterComponents = true

// Entity is an id that has components,
// they can only be created using ecs.NewEntity()
type Entity uint32

// Entity 0 is never used
const UnusedEntity = Entity(0)

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
	b             bitset.BitSet
	mut           sync.Mutex
	registeredNum int // used for sorting, set when the storage is registered by the pool
}
type _Storage interface {
	delete(e Entity)
	getBitset() *bitset.BitSet
	lock()
	unlock()
	getNum() int
}

func (s *Storage[Component]) getNum() int {
	return s.registeredNum
}

func (s *Storage[Component]) delete(e Entity) {
	s.mut.Lock()
	defer s.mut.Unlock()
	var zero Component
	s.components[e] = zero
}

// takes in other storages and returns
// entities that exist in all of them
//
//	in simple terms:
//	 entities that have all of these components
//
// passing in nil or nothing will return entities that have this storage's component
func (s *Storage[Component]) And(storages ..._Storage) []Entity {
	unlock := s.orderedLock(storages...)
	defer unlock()
	mask := s.b.Clone()
	if len(storages) > 0 {
		for _, otherSt := range storages {
			if otherSt != nil {
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
func (s *Storage[Component]) ButNot(storages ..._Storage) []Entity {
	unlock := s.orderedLock(storages...)
	defer unlock()
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
func (st *Storage[Component]) Update(e Entity, component Component) {
	st.mut.Lock()
	defer st.mut.Unlock()
	if !st.b.IsSet(uint(e)) {
		return
	}
	st.components[e] = component
}

// check if an Entity has a component
func (st *Storage[Component]) EntityHasComponent(e Entity) bool {
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
func (s *Storage[Component]) Get(e Entity) (component Component) {
	s.mut.Lock()
	defer s.mut.Unlock()
	if !s.b.IsSet(uint(e)) {
		return component
	}
	return s.components[e]
}

func newStorage[T any](size int, num int) *Storage[T] {
	return &Storage[T]{
		components:    make([]T, size),
		registeredNum: num,
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
	maxEntities int
	//how many entities are alive
	aliveCount int
	// generations track how many times an entity was recycled
	generations   []uint32
	numComponents int // how many components have been registered so far.
	mut           sync.Mutex
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
	p := &Pool{
		stores:         make(map[any]_Storage),
		componentsUsed: make(map[Entity][]_Storage),
		generations:    make([]uint32, size),
		maxEntities:    size + 1,
	}
	NewEntity(p) // entity 0 is unused
	return p
}

// Get an entity
// this will panic if pool does not have entities available
func NewEntity(p *Pool) Entity {
	p.mut.Lock()
	defer p.mut.Unlock()
	// if no entities are available for recycling
	if len(p.freeList) == 0 {
		if p.aliveCount >= p.maxEntities {

			msg := fmt.Sprintf("Entity limit exceeded. please initialize more entities by increasing the number you passed to ecs.New(). \nGiven size: %d\n Entity: %d", p.maxEntities, p.aliveCount+1)
			panic(msg)
		}
		e := Entity(p.aliveCount)
		p.aliveEntities.Set(uint(e))
		p.aliveCount++
		return e
	}
	// recycle an entity
	var newEntity = p.freeList[0]
	p.freeList = p.freeList[1:]
	var storagesUsed []_Storage = p.componentsUsed[newEntity]
	for _, store := range storagesUsed {
		//zero out the component for this entity
		store.delete(newEntity)
	}
	// entity no longer has these components
	// set slice length to 0
	p.componentsUsed[newEntity] = p.componentsUsed[newEntity][:0]
	p.generations[newEntity] += 1

	return newEntity
}

// give entities back to the pool
func Kill(p *Pool, entities ...Entity) {
	p.mut.Lock()
	defer p.mut.Unlock()
	for _, e := range entities {
		if e == 0 { // cannot kill entity 0 (unused)
			continue
		}
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
func Register[Component any](pool *Pool) {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	var nilptr *Component
	_, ok := pool.stores[nilptr]
	if !ok {
		pool.stores[nilptr] = newStorage[Component](pool.maxEntities, pool.numComponents)
		pool.numComponents++
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
func Add[Component any](pool *Pool, e Entity, component Component) {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	if !pool.aliveEntities.IsSet(uint(e)) {
		return
	}
	st := registerAndGetStorage[Component](pool)
	st.mut.Lock()
	defer st.mut.Unlock()
	if st.b.IsSet(uint(e)) {
		return
	}
	st.b.Set(uint(e))
	st.components[e] = component
	pool.componentsUsed[e] =
		append(pool.componentsUsed[e], st)
}

// remove a component from an entity
func Remove[Component any](pool *Pool, e Entity) {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	st := registerAndGetStorage[Component](pool)
	st.mut.Lock()
	if !st.b.IsSet(uint(e)) {
		return
	}
	st.mut.Unlock()
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
			pool.componentsUsed[e] = s[0 : len(s)-1]
			return
		}
	}
}

// check if an entity has a component
// shorthand for
//
//	POSITION := ecs.GetStorage[Position](pool)
//	POSITION.EntityHasComponent(entity)
func Has[Component any](pool *Pool, e Entity) bool {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	st := registerAndGetStorage[Component](pool)
	st.mut.Lock()
	defer st.mut.Unlock()
	return st.b.IsSet(uint(e))
}

// Check if an entity is alive
func IsAlive(pool *Pool, e Entity) bool {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	return pool.aliveEntities.IsSet(uint(e))
}

// NOTE: Only useful if you are storing entities in components.
//
// Check if an entity is alive, given its generation (reuse count).
//
// NOTE: check if the entity you were storing is alive with this before running
// the system on it
// NOTE: You can get an entity's generation with ecs.GetGeneration.
func IsAliveWithGeneration(pool *Pool, e Entity, generation uint32) bool {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	if pool.generations[e] != generation {
		return false
	}
	return pool.aliveEntities.IsSet(uint(e))
}

// NOTE: Only useful if you are storing entities inside of components.
//
// generation is the number of times the entity has been reused.
//
// NOTE: also see ecs.IsAliveWithGeneration
func GetGeneration(pool *Pool, e Entity) (generation uint32) {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	if e > Entity(pool.maxEntities) {
		return 0
	}
	return pool.generations[e]
}

// storage contains all components of a type
func GetStorage[Component any](pool *Pool) *Storage[Component] {
	pool.mut.Lock()
	defer pool.mut.Unlock()
	st := registerAndGetStorage[Component](pool)
	return st
}

// same as public register but also returns the storage.
// it will not allocate a new storage if it already exists
// this will use the pool's mutexes appropriately
func registerAndGetStorage[Component any](pool *Pool) *Storage[Component] {
	var nilptr *Component
	st, ok := pool.stores[nilptr]
	if ok {
		return st.(*Storage[Component])
	} else if AutoRegisterComponents {
		// allocate storage
		var st = newStorage[Component](pool.maxEntities, pool.numComponents)
		pool.numComponents++
		pool.stores[nilptr] = st
		return st
	}
	var zero Component
	panic(fmt.Sprintf("Component of type %T was not registered", zero))
}

func (s *Storage[Component]) getBitset() *bitset.BitSet {
	return &s.b
}

// If two goroutines call these methods with reversed storage parameters:

// Goroutine 1: storageA.And(storageB)  // Locks A → B
// Goroutine 2: storageB.And(storageA)  // Locks B → A
// This creates a deadlock:

// Goroutine 1 holds lock A, waits for lock B

// Goroutine 2 holds lock B, waits for lock A

// Both goroutines wait forever

// The Solution: Consistent Ordering
// The fix ensures locks are always acquired in the same order regardless of parameter order:
func (s *Storage[Component]) orderedLock(storages ..._Storage) func() {
	all := make([]_Storage, 0, len(storages)+1)
	for _, s := range storages {
		if s != nil {
			all = append(all, s)
		}
	}
	// Sort by underlying storage addresses
	slices.SortFunc(all, func(a, b _Storage) int {
		if a.getNum() < b.getNum() {
			return -1
		}
		return +1
	})
	// Lock in sorted order
	for _, st := range all {
		st.lock()
	}

	// Return unlock function (reverse order)
	return func() {
		for i := len(all) - 1; i >= 0; i-- {
			all[i].unlock()
		}
	}
}

func (s *Storage[Component]) lock() {
	s.mut.Lock()
}
func (s *Storage[Component]) unlock() {
	s.mut.Unlock()
}
