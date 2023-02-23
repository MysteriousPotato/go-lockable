package lockable

import (
	"sync"
)

type Lockable[T comparable] struct {
	locks   map[T]*versionedMutex
	locksMu *sync.Mutex
}

type versionedMutex struct {
	sync.RWMutex
	completedVersion int
	currentVersion   int
}

func (m Lockable[T]) LockKey(key T) {
	vMu := m.lockKey(key, false)
	vMu.Lock()
}

func (m Lockable[T]) UnlockKey(key T) {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	vMu, ok := m.locks[key]
	if !ok {
		return
	}

	vMu.completedVersion++
	m.tryCleanUp(key, vMu)
	vMu.Unlock()
}

func (m Lockable[T]) LockKeyDuring(key T, fn func() (interface{}, error)) (interface{}, error) {
	m.LockKey(key)
	defer m.UnlockKey(key)

	return fn()
}

func (m Lockable[T]) RLockKey(key T) {
	vMu := m.lockKey(key, true)
	vMu.RLock()
}

// This is solely to mimic RWMutex. It is equivalent of calling [Lockable.UnlockKey]
func (m Lockable[T]) RUnlockKey(key T) {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	vMu, ok := m.locks[key]
	if !ok {
		return
	}

	vMu.completedVersion++
	m.tryCleanUp(key, vMu)
	vMu.RUnlock()
}

func (m Lockable[T]) RLockKeyDuring(key T, fn func() (interface{}, error)) (interface{}, error) {
	m.RLockKey(key)
	defer m.RUnlockKey(key)

	return fn()
}

func (m Lockable[T]) lockKey(key T, isRead bool) *versionedMutex {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	vMu, ok := m.locks[key]
	if !ok {
		vMu = &versionedMutex{
			currentVersion:   0,
			completedVersion: 0,
			RWMutex:          sync.RWMutex{},
		}
		m.locks[key] = vMu
	}

	vMu.currentVersion++
	return vMu
}

// Clean up the lock if no other locks have been requested for this key
func (m Lockable[T]) tryCleanUp(key T, vMu *versionedMutex) {
	if vMu.currentVersion == vMu.completedVersion {
		delete(m.locks, key)
	}
}
