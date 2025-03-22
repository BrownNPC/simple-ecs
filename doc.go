/*
entity component system that is performant an easy to use.

	The heart of this ECS is the memory pool
	Think of the pool like a database.
	On the Y axis (columns) there are arrays of components

	We use a struct called storage to hold the components arrays
	 components can be any data type, but they cannot be interfaces
	These arrays are pre-allocated to a fixed size provided by the user

	an entity is just an index into these arrays
	So on the X axis there are entities which are just indexes

	The storage struct also has a bitset.

	each bit in the bitset corresponds to an entity
	 the bitset is used for maintaining
	a record of which entity has the component the storage is storing

	The pool also has its own bitset that tracks which entities are alive

		there is also a map from entities to a slice of component storages

		we update this map when an entity has a component added to it

		we use this map to go into every storage and zero out the component
		when an entity is killed
*/
package ecs
