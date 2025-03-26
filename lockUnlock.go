package ecs

import (
	"sync"

	"github.com/BrownNPC/simple-ecs/internal"
)

// disable the use of mutexes.
// generally you dont need to do this unless in performance
// critical single-threaded scenarios
func DisableMutex() {
	useMutex = false
	internal.UseMutex = false
}

// The library will use Mutexes by default
func EnableMutex() {
	useMutex = true
	internal.UseMutex = true
}

var useMutex = true

// wraps RW mutex to allow for disabling it
type CustomRWMutex struct {
	mut sync.RWMutex
}

func (s *CustomRWMutex) Lock() {
	if useMutex {
		s.mut.Lock()
	}
}

func (s *CustomRWMutex) Unlock() {
	if useMutex {
		s.mut.Unlock()
	}
}

func (s *CustomRWMutex) RLock() {
	if useMutex {
		s.mut.RLock()
	}
}

func (s *CustomRWMutex) RUnlock() {
	if useMutex {
		s.mut.RUnlock()
	}
}
