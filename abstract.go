package lockd

import (
	"context"
	"sync"
)

// Locker is something, which can be locked and unlocked.
type Locker = sync.Locker

// KeyLocker is lock, which locks on per-key basis.
type KeyLocker interface {
	GetLock(key string) Locker
}

// ContextKeyLocker is lock which locks on per-key basis and uses conext for locking methods.
type ContextKeyLocker interface {
	GetLock(key string) ContextLocker
}

// ContextLocker locks resource but uses context and may return error/timeout.
type ContextLocker interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}
