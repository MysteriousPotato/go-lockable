package lockable

import (
	"sync"
)

type (
	// Lockable can be used to add lock-by-key support to any struct.
	//
	// The zero value is not ready for use. Refer to [New] for creating ready-to-use instance.
	//
	// Lockable has an interface that's deliberately similar to [sync.RWMutex].
	//
	// For exemple, instead of using:
	//
	//	func main() {
	//		// Read/write lock
	//		myMutex.Lock()
	//		defer myMutex.Unlock()
	//
	//		// Read lock
	//		myMutex.RLock()
	//		defer myMutex.RUnlock()
	//	}
	//
	// You would use it like so:
	//
	//	func main() {
	//		// Read/write lock
	//		myLockable.LockKey(arbitraryKey)
	//		defer myLockable.Unlock(arbitraryKey)
	//
	//		// Read lock
	//		myLockable.RLock(arbitraryKey)
	//		defer myLockable.RUnlock(arbitraryKey)
	//	}
	//
	// You can use it to add lock-by-key support to an already existing struct.
	//
	// Ex.:
	//
	//	func main() {
	//			type ArbitraryType struct {
	//				lockable.Lockable[string]
	//				// properties...
	//			}
	//
	//			arbitrary := ArbitraryType{
	//				Lockable: lockable.New[string](),
	//			}
	//
	//			// You can now use it the way to acquire locks
	//			arbitrary.LockKey("potato")
	//			defer arbitrary.UnlockKey("potato")
	//
	//			// Do stuff...
	//	}
	//
	// [Lockable.LockKeyDuring] and [Lockable.RLockKeyDuring] can be used to automatically manage lock acquisition/release.
	//
	// Ex.:
	// 		func main() {
	//			type ArbitraryType struct {
	//				lockable.Lockable[string]
	//				// properties...
	//			}
	//
	//			arbitrary := ArbitraryType{
	//				Lockable: lockable.New[string](),
	//			}
	//
	//			_, err := arbitrary.LockKeyDuring("potato", func() (any, error) {
	//				// Do stuff...
	//			})
	//		}
	//
	// You can use [Lockable.IsLocked] to check if a lock is currently being held without locking the key.
	// Ex.:
	// 		func main() {
	//			type ArbitraryType struct {
	//				lockable.Lockable[string]
	//				// properties...
	//			}
	//
	//			arbitrary := ArbitraryType{
	//				Lockable: lockable.New[string](),
	//			}
	//
	//			if arbitrary.IsLocked("potato") {
	//			 	//.. do stuff.
	//			}
	//		}
	Lockable[T comparable] struct {
		locks   map[T]*versionedMutex
		locksMu *sync.Mutex
	}
	versionedMutex struct {
		sync.RWMutex
		completedVersion int
		currentVersion   int
	}
)

// New creates a ready-to-use Lockable instance.
//
// Refer to [Lockable] for usage
func New[T comparable]() Lockable[T] {
	return Lockable[T]{
		locks:   map[T]*versionedMutex{},
		locksMu: &sync.Mutex{},
	}
}

// LockKey method is used to acquire read/write locks.
//
// Use [Lockable.RLockKey] for read locks.
func (l Lockable[T]) LockKey(key T) {
	vMu := l.lockKey(key)
	vMu.Lock()
}

// UnlockKey method is used to release read/write locks.
//
// Can safely be called multiple times on the same key.
func (l Lockable[T]) UnlockKey(key T) {
	vMu, ok := l.unlockKey(key)
	if !ok {
		return
	}
	vMu.Unlock()
}

// LockKeyDuring will automatically acquire a read/write lock before executing fn and release it once done.
func (l Lockable[T]) LockKeyDuring(key T, fn func() (any, error)) (any, error) {
	l.LockKey(key)
	defer l.UnlockKey(key)

	return fn()
}

// RLockKey method is used to acquire read locks.
//
// Use [Lockable.RLockKey] for read/write locks.
func (l Lockable[T]) RLockKey(key T) {
	vMu := l.lockKey(key)
	vMu.RLock()
}

// RUnlockKey method is used to release read/write locks.
//
// Can safely be called multiple times on the same key.
func (l Lockable[T]) RUnlockKey(key T) {
	vMu, ok := l.unlockKey(key)
	if !ok {
		return
	}
	vMu.RUnlock()
}

// RLockKeyDuring before executing fn and release it once done.
func (l Lockable[T]) RLockKeyDuring(key T, fn func() (any, error)) (any, error) {
	l.RLockKey(key)
	defer l.RUnlockKey(key)

	return fn()
}

// IsLocked is used to determine whether a key has been locked without locking the key.
func (l Lockable[T]) IsLocked(key T) bool {
	l.locksMu.Lock()
	defer l.locksMu.Unlock()

	keyLock, ok := l.locks[key]
	return ok && keyLock.completedVersion != keyLock.currentVersion
}

func (l Lockable[T]) lockKey(key T) *versionedMutex {
	l.locksMu.Lock()
	defer l.locksMu.Unlock()

	vMu, ok := l.locks[key]
	if !ok {
		vMu = &versionedMutex{
			currentVersion:   0,
			completedVersion: 0,
			RWMutex:          sync.RWMutex{},
		}
		l.locks[key] = vMu
	}

	vMu.currentVersion++
	return vMu
}

func (l Lockable[T]) unlockKey(key T) (*versionedMutex, bool) {
	l.locksMu.Lock()
	defer l.locksMu.Unlock()

	vMu, ok := l.locks[key]
	if !ok {
		return nil, false
	}

	vMu.completedVersion++
	l.tryCleanUp(key, vMu)

	return vMu, true
}

// Clean up the lock if no other locks have been requested for this key
func (l Lockable[T]) tryCleanUp(key T, vMu *versionedMutex) {
	if vMu.currentVersion == vMu.completedVersion {
		delete(l.locks, key)
	}
}
