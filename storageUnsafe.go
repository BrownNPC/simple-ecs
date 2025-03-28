package ecs

//	You probably dont need to use this. The performance
//	gain is negligible.
//
// get a pointer to an entity's component
// so that you dont need to update the entity.
// if the entity is dead or does not have this component, the return value is nil
//
//	this is NOT thread-safe, look at storage.AcquireLockUnsafe.
//	DO NOT store the pointer for later use
//  DO NOT forget to check for a nil return
func (st *Storage[T]) GetPtrUnsafe(e Entity) *T {
	if !st.EntityHasComponent(e) {
		return nil
	}
	return &st.components[e]
}

//	only use this if you are going to use storage.GetPtrUnsafe()
//
// Lock the storage mutex. The idea is that before you start
// modifying the components using their pointers, you lock the storage BEFORE looping
// and unlock it AFTER the loop using storage.FreeLockUnsafe()
// (to prevent memory corruption when you're using goroutines)
//
// Warning: you should only lock the storage BEFORE the loop, and AFTER querying
//
//		func MovementSystem(p *ecs.Pool) {
//		     POSITION, VELOCITY := ecs.GetStorage2[Position, Velocity](p)
//		     // query BEFORE locking
//		     entities := POSITION.And(VELOCITY)
//		     // lock before the loop but after querying
//		     POSITION.AcquireLockUnsafe()
//		     VELOCITY.AcquireLockUnsafe()
//		     // defer before looping
//		     defer POSITION.FreeLockUnsafe()
//		     defer VELOCITY.FreeLockUnsafe()
//	         // loop over entities
//		     for _, e := range entities {
//		     	pos, vel := POSITION.GetPtrUnsafe(e), VELOCITY.GetPtrUnsafe(e)
//  			if pos==nil || vel ==nil{
//  				continue
//  			}
//		     	pos.X += vel.X
//		     	pos.Y += vel.Y
//		     }
//		}
func (st *Storage[Component]) AcquireLockUnsafe() {
	st.mut.Lock()
}

//	only use this if you are going to use storage.GetPtrUnsafe()
//
// UnLock the storage mutex. The idea is that before you start
// modifying the components using their pointers, you lock the storage BEFORE looping
// and unlock it AFTER the loop using storage.FreeLockUnsafe()
// (to prevent memory corruption when you're using goroutines)
//
// Warning: you should only unlock the storage AFTER the loop
//
//		func MovementSystem(p *ecs.Pool) {
//		     POSITION, VELOCITY := ecs.GetStorage2[Position, Velocity](p)
//		     // query BEFORE locking
//		     entities := POSITION.And(VELOCITY)
//		     // lock before the loop but after querying
//		     POSITION.AcquireLockUnsafe()
//		     VELOCITY.AcquireLockUnsafe()
//		     // defer before looping
//		     defer POSITION.FreeLockUnsafe()
//		     defer VELOCITY.FreeLockUnsafe()
//	         // loop over entities
//		     for _, e := range entities {
//		     	pos, vel := POSITION.GetPtrUnsafe(e), VELOCITY.GetPtrUnsafe(e)
//  			if pos==nil || vel ==nil{
//  				continue
//  			}
//		     	pos.X += vel.X
//		     	pos.Y += vel.Y
//		     }
//		}
func (st *Storage[Component]) FreeLockUnsafe() {
	st.mut.Unlock()
}
