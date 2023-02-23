package lockable

import (
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestLockableMap(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e", "f", "g"}
	writes := 100

	wg := &sync.WaitGroup{}
	lMap := NewMap[string, int]()
	ch := make(chan string, writes*len(keys))
	for _, key := range keys {
		for i := 0; i < writes; i++ {
			wg.Add(1)
			lMap.LockKey(key)

			go func(key string, i int, ch chan string) {
				defer func() {
					lMap.UnlockKey(key)
					ch <- key + strconv.Itoa(i)
					wg.Done()
				}()

				lMap.Store(key, i)
			}(key, i, ch)
		}
	}
	wg.Wait()
	close(ch)

	insertOrder := sort.StringSlice{}
	for res := range ch {
		insertOrder = append(insertOrder, res)
	}
	if sort.IsSorted(insertOrder) {
		t.Fatal("expected unsorted slice")
	}

	lMap.Range(func(key string, value int) bool {
		if value != writes-1 {
			t.Fatalf("expected %v, got %v", writes-1, value)
		}
		return true
	})
}

func TestLockableUMutexMapLock(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e", "f", "g"}
	writes := 100

	wg := &sync.WaitGroup{}
	lMap := NewMutexMap[string]()
	ch := make(chan string, writes*len(keys))
	for _, key := range keys {
		for i := 0; i < writes; i++ {
			wg.Add(1)
			lMap.LockKey(key)

			go func(key string, i int, ch chan string) {
				defer func() {
					lMap.UnlockKey(key)
					ch <- key + strconv.Itoa(i)
					wg.Done()
				}()
				lMap.Store(key, i)
			}(key, i, ch)
		}
	}
	wg.Wait()
	close(ch)

	insertOrder := sort.StringSlice{}
	for res := range ch {
		insertOrder = append(insertOrder, res)
	}
	if sort.IsSorted(insertOrder) {
		t.Fatal("expected unsorted slice")
	}

	lMap.Range(func(key, value any) bool {
		if value != writes-1 {
			t.Fatalf("expected %v, got %v", writes-1, value)
		}
		return true
	})
}

// This benchmark is mostly meaningless.
//
// It's only goal is to show that using "per key" locks is much more performant when locking during async code. Duh!
func BenchmarkLockableMap(b *testing.B) {
	writes := 5
	reads := 200
	blockingLocks := 1
	blockingDuration := time.Millisecond * 10
	keys := []string{}
	for i := 0; i < 10; i++ {
		keys = append(keys, strconv.Itoa(i))
	}

	lUMuMap := NewMutexMap[string]()
	b.Run("lockableMutexMap", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		for n := 0; n < b.N; n++ {
			for _, key := range keys {
				//writes
				for i := 0; i < writes; i++ {
					wg.Add(1)
					go func(key string, i int) {
						lUMuMap.LockKey(key)
						defer lUMuMap.UnlockKey(key)
						defer wg.Done()

						lUMuMap.Store(key, i)
					}(key, i)
				}
				//reads
				for i := 0; i < reads; i++ {
					wg.Add(1)
					go func(key string) {
						lUMuMap.RLockKey(key)
						defer lUMuMap.RUnlockKey(key)
						defer wg.Done()

						lUMuMap.Load(key)
					}(key)
				}
				//blockingLocks
				for i := 0; i < blockingLocks; i++ {
					wg.Add(1)
					go func(key string) {
						lUMuMap.LockKey(key)
						defer lUMuMap.UnlockKey(key)
						defer wg.Done()

						time.Sleep(blockingDuration)
					}(key)
				}
			}
		}
		wg.Wait()
	})

	lMap := NewMap[string, int]()
	b.Run("lockableMap", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		for n := 0; n < b.N; n++ {
			for _, key := range keys {
				//writes
				for i := 0; i < writes; i++ {
					wg.Add(1)
					go func(key string, i int) {
						defer wg.Done()

						lMap.LockKey(key)
						defer lMap.UnlockKey(key)

						lMap.Store(key, i)
					}(key, i)
				}
				//reads
				for i := 0; i < reads; i++ {
					wg.Add(1)
					go func(key string) {
						defer wg.Done()

						lMap.RLockKey(key)
						defer lMap.RUnlockKey(key)

						lMap.Load(key)
					}(key)
				}
				//blockingLocks
				for i := 0; i < blockingLocks; i++ {
					wg.Add(1)
					go func(key string) {
						lMap.LockKey(key)
						defer lMap.UnlockKey(key)
						defer wg.Done()

						time.Sleep(blockingDuration)
					}(key)
				}
			}
		}
		wg.Wait()
	})

	valueMapWithLock := map[string]interface{}{}
	mapMu := &sync.RWMutex{}
	b.Run("map", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		for n := 0; n < b.N; n++ {
			for _, key := range keys {
				//writes
				for i := 0; i < writes; i++ {
					wg.Add(1)
					go func(key string, i int) {
						mapMu.Lock()
						defer mapMu.Unlock()
						defer wg.Done()

						valueMapWithLock[key] = i
					}(key, i)
				}
				//reads
				for i := 0; i < reads; i++ {
					wg.Add(1)
					go func(key string) {
						mapMu.RLock()
						defer mapMu.RUnlock()
						defer wg.Done()

						_ = valueMapWithLock[key]
					}(key)
				}
				//blockingLocks
				for i := 0; i < blockingLocks; i++ {
					wg.Add(1)
					go func(key string) {
						mapMu.Lock()
						defer mapMu.Unlock()
						defer wg.Done()

						time.Sleep(blockingDuration)
					}(key)
				}
			}
		}
		wg.Wait()
	})
}
