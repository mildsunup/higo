package lock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryLocker 内存锁工厂 (仅用于单机测试)
type MemoryLocker struct {
	mu    sync.Mutex
	locks map[string]*memoryLockEntry
}

type memoryLockEntry struct {
	token     string
	expiresAt time.Time
}

// NewMemoryLocker 创建内存锁工厂
func NewMemoryLocker() *MemoryLocker {
	return &MemoryLocker{locks: make(map[string]*memoryLockEntry)}
}

func (l *MemoryLocker) NewLock(key string, opts ...Option) Lock {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if options.Token == "" {
		options.Token = uuid.New().String()
	}
	return &memoryLock{locker: l, key: key, opts: options}
}

type memoryLock struct {
	locker *MemoryLocker
	key    string
	opts   Options
}

func (l *memoryLock) Lock(ctx context.Context) error {
	retries := 0
	for {
		ok, err := l.TryLock(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		retries++
		if l.opts.RetryCount >= 0 && retries > l.opts.RetryCount {
			return ErrLockFailed
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.opts.RetryDelay):
		}
	}
}

func (l *memoryLock) TryLock(ctx context.Context) (bool, error) {
	l.locker.mu.Lock()
	defer l.locker.mu.Unlock()

	now := time.Now()
	if entry, ok := l.locker.locks[l.key]; ok && now.Before(entry.expiresAt) {
		return false, nil
	}

	l.locker.locks[l.key] = &memoryLockEntry{
		token:     l.opts.Token,
		expiresAt: now.Add(l.opts.TTL),
	}
	return true, nil
}

func (l *memoryLock) Unlock(ctx context.Context) error {
	l.locker.mu.Lock()
	defer l.locker.mu.Unlock()

	entry, ok := l.locker.locks[l.key]
	if !ok || entry.token != l.opts.Token {
		return ErrNotHeld
	}
	delete(l.locker.locks, l.key)
	return nil
}

func (l *memoryLock) Refresh(ctx context.Context) error {
	l.locker.mu.Lock()
	defer l.locker.mu.Unlock()

	entry, ok := l.locker.locks[l.key]
	if !ok || entry.token != l.opts.Token {
		return ErrLockExpired
	}
	entry.expiresAt = time.Now().Add(l.opts.TTL)
	return nil
}

var _ Locker = (*MemoryLocker)(nil)
var _ Lock = (*memoryLock)(nil)
