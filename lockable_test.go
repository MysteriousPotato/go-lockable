package lockable_test

import (
	"github.com/MysteriousPotato/go-lockable"
	"sync"
	"testing"
)

func TestIsLocked(t *testing.T) {
	lock := lockable.New[string]()

	lock.RLockKey("potato")
	if !lock.IsLocked("potato") {
		t.Fatal("expected isLocked true, got false")
	}

	lock.RUnlockKey("potato")
	if lock.IsLocked("potato") {
		t.Fatal("expected isLocked false, got true")
	}
}

func BenchmarkLockableLock(b *testing.B) {
	l := lockable.New[string]()
	for i := 0; i < b.N; i++ {
		l.LockKey("key")
		l.UnlockKey("key")
	}
}

func BenchmarkLockableRLock(b *testing.B) {
	l := lockable.New[string]()
	for i := 0; i < b.N; i++ {
		l.RLockKey("key")
		l.RUnlockKey("key")
	}
}

func BenchmarkStdMutexLock(b *testing.B) {
	l := &sync.RWMutex{}
	for i := 0; i < b.N; i++ {
		l.Lock()
		l.Unlock()
	}
}

func BenchmarkStdMutexRLock(b *testing.B) {
	mu := &sync.RWMutex{}
	for i := 0; i < b.N; i++ {
		mu.Lock()
		mu.Unlock()
	}
}
