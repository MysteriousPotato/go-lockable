package lockable

import (
	"sync"
)

type Map[T comparable, V any] struct {
	internalMap map[T]V
	mu          *sync.RWMutex
	Lockable[T]
}

type MutexMap[T comparable] struct {
	sync.Map
	Lockable[T]
}

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

func NewMutexMap[T comparable]() MutexMap[T] {
	return MutexMap[T]{
		Lockable: Lockable[T]{
			locks:   map[T]*versionedMutex{},
			locksMu: &sync.Mutex{},
		},
	}
}

func (m Map[T, V]) Load(key T) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.internalMap[key]
	return v, ok
}

func (m Map[T, V]) Store(key T, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.internalMap[key] = value
}

func (m Map[T, V]) Delete(key T) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.internalMap, key)
}

func (m Map[T, V]) Range(fn func(key T, value V) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, v := range m.internalMap {
		if keepGoing := fn(k, v); !keepGoing {
			break
		}
	}
}
