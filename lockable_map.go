package lockable

import (
	"sync"
)

// Map implements a map that can acquire key-specific locks.
//
// The zero value is not ready for use. Refer to [NewMap] to create a ready-to-use instance.
//
// Map has an interface that's deliberately similar to [sync.Map], but uses a combination of RWMutex/map to simulate it.
//
// [MutexMap] can be used instead if you want to use sync.Map internally.
// Refer to [sync.Map]'s document for use cases.
//
// Ex. usage:
//
//	func main() {
//		lockableMap := lockable.NewMap[string, int]()
//
//
//		// This will only lock access the "potato" key
//		// Keys do not need to exist prior to acquiring the key lock
//		lockableMap.LockKey("potato")
//		defer lockableMap.UnlockKey("potato")
//
//		// Do async stuff....
//
//		lockableMap.Store("potato", 10)
//	}
//
// Refer to [Lockable] for more detailed exemples of locking.
type Map[T comparable, V any] struct {
	internalMap map[T]V
	mu          *sync.RWMutex
	Lockable[T]
}

// MutexMap acts the same as [Map] except it uses a sync.Map.
//
// The zero value is not ready for use. Refer to [NewMutexMap] to create a ready-to-use instance.
//
//	Ex. usage:
//	func main() {
//		lockableMap := lockable.NewMutexMap[string]()
//
//
//		// This will only lock access the "potato" key
//		// Keys do not need to exist prior to acquiring the key lock
//		lockableMap.LockKey("potato")
//		defer lockableMap.UnlockKey("potato")
//
//		// Do async stuff....
//
//		lockableMap.Store("potato", 10)
//	}
//
// Refer to [Lockable] for more detailed exemples of locking
type MutexMap[T comparable] struct {
	sync.Map
	Lockable[T]
}

// NewMap creates a ready-to-use Map instance.
//
// Refer to [Map] for usage.
func NewMap[T comparable, V any]() Map[T, V] {
	return Map[T, V]{
		internalMap: map[T]V{},
		mu:          &sync.RWMutex{},
		Lockable: Lockable[T]{
			locks:   map[T]*versionedMutex{},
			locksMu: &sync.Mutex{},
		},
	}
}

// NewMutexMap creates a ready-to-use MutexMap instance.
//
// Refer to [MutexMap] for usage.
func NewMutexMap[T comparable]() MutexMap[T] {
	return MutexMap[T]{
		Lockable: Lockable[T]{
			locks:   map[T]*versionedMutex{},
			locksMu: &sync.Mutex{},
		},
	}
}

// Load effectively serves the same purpose as [sync.Map.Load]
func (m Map[T, V]) Load(key T) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.internalMap[key]
	return v, ok
}

// Store effectively serves the same purpose as [sync.Map.Store]
func (m Map[T, V]) Store(key T, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.internalMap[key] = value
}

// Delete effectively serves the same purpose as [sync.Map.Delete]
func (m Map[T, V]) Delete(key T) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.internalMap, key)
}

// Range effectively serves the same purpose as [sync.Map.Range]
func (m Map[T, V]) Range(fn func(key T, value V) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, v := range m.internalMap {
		if keepGoing := fn(k, v); !keepGoing {
			break
		}
	}
}
