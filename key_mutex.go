package lockd

import (
	"fmt"
	"sync"
)

type keyMutex struct {
	mutexes []sync.Mutex
	hasher  Hasher
}

type kmKey struct {
	factory *keyMutex
	keyHash HashResult
}

func (km *kmKey) Lock() {
	idx := int(km.keyHash % HashResult(len(km.factory.mutexes)))
	km.factory.mutexes[idx].Lock()
}

func (km *kmKey) Unlock() {
	idx := int(km.keyHash % HashResult(len(km.factory.mutexes)))
	km.factory.mutexes[idx].Unlock()
}

// GetLock returns lock for given key.
func (km *keyMutex) GetLock(key string) Locker {
	return &kmKey{
		factory: km,
		keyHash: km.hasher(key),
	}
}

// NewKeyMutex returns new keyMutex.
// KeyMutex creates pool of SIZE mutexes.
// It uses hash function to hash key and use it to get index of mutex and lock/unlock it.
func NewKeyMutex(size int) KeyLocker {
	if size <= 0 {
		panic(fmt.Sprintf("lockd: Invalid size provided for NewKeyMutex: %d", size))
	}
	return &keyMutex{
		hasher:  DefaultHasher,
		mutexes: make([]sync.Mutex, size),
	}
}
