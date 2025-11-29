package lock

import (
	"context"
	"testing"
	"time"
)

func TestMemoryLock_LockUnlock(t *testing.T) {
	locker := NewMemoryLocker()
	lock := locker.NewLock("key", WithTTL(time.Second))
	ctx := context.Background()

	if err := lock.Lock(ctx); err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	if err := lock.Unlock(ctx); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
}

func TestMemoryLock_TryLock(t *testing.T) {
	locker := NewMemoryLocker()
	lock1 := locker.NewLock("key", WithTTL(time.Second))
	lock2 := locker.NewLock("key", WithTTL(time.Second))
	ctx := context.Background()

	ok, _ := lock1.TryLock(ctx)
	if !ok {
		t.Error("First TryLock should succeed")
	}

	ok, _ = lock2.TryLock(ctx)
	if ok {
		t.Error("Second TryLock should fail")
	}

	lock1.Unlock(ctx)

	ok, _ = lock2.TryLock(ctx)
	if !ok {
		t.Error("TryLock should succeed after unlock")
	}
}
