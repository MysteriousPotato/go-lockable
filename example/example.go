package example

import (
	"github.com/MysteriousPotato/go-lockable"
)

func main() {
	lockableMap := lockable.NewMap[int, int]()

	//Lock
	for i := 0; i < 10; i++ {
		go func(i int) {
			key := i % 4

			lockableMap.LockKey(key)
			defer lockableMap.UnlockKey(key)

			lockableMap.Store(key, i)
		}(i)
	}

	//Read lock
	for i := 0; i < 10; i++ {
		go func(i int) {
			key := i % 4

			lockableMap.RLockKey(key)
			defer lockableMap.RUnlockKey(key)

			lockableMap.Store(key, i)
		}(i)
	}

	//Lock using callback
	for i := 0; i < 10; i++ {
		go func(i int) {
			key := i % 4

			_, _ = lockableMap.LockKeyDuring(key, func() (interface{}, error) {
				lockableMap.Store(key, i)
				return nil, nil
			})
		}(i)
	}
}
