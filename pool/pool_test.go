package pool

import (
	"bytes"
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestObject_GetPut(t *testing.T) {
	pool := NewObject(
		func() *bytes.Buffer { return new(bytes.Buffer) },
		func(b *bytes.Buffer) { b.Reset() },
	)

	buf := pool.Get()
	buf.WriteString("hello")

	pool.Put(buf)

	// 再次获取，应该是重置后的
	buf2 := pool.Get()
	if buf2.Len() != 0 {
		t.Error("buffer should be reset")
	}
}

func TestObject_Stats(t *testing.T) {
	pool := NewObject(
		func() int { return 0 },
		nil,
	)

	_ = pool.Get()
	_ = pool.Get()
	pool.Put(1)

	stats := pool.Stats()
	if stats.Gets != 2 {
		t.Errorf("expected 2 gets, got %d", stats.Gets)
	}
	if stats.Puts != 1 {
		t.Errorf("expected 1 put, got %d", stats.Puts)
	}
	if stats.InUse != 1 {
		t.Errorf("expected 1 in use, got %d", stats.InUse)
	}
}

func TestBufferPool(t *testing.T) {
	buf := GetBuffer()
	buf.WriteString("test")

	if buf.String() != "test" {
		t.Error("buffer content mismatch")
	}

	PutBuffer(buf)
}

func TestByteSlicePool(t *testing.T) {
	// 获取小切片
	small := GetBytes(50)
	if cap(small) < 50 {
		t.Errorf("expected cap >= 50, got %d", cap(small))
	}

	// 获取大切片
	large := GetBytes(10000)
	if cap(large) < 10000 {
		t.Errorf("expected cap >= 10000, got %d", cap(large))
	}

	PutBytes(small)
	PutBytes(large)
}

func TestWorker_Submit(t *testing.T) {
	pool := NewWorker(2, 10)
	defer pool.Stop()

	var count atomic.Int32
	for i := 0; i < 5; i++ {
		err := pool.Submit(func() {
			count.Add(1)
		})
		if err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	if count.Load() != 5 {
		t.Errorf("expected 5 tasks completed, got %d", count.Load())
	}
}

func TestWorker_SubmitWait(t *testing.T) {
	pool := NewWorker(1, 1)
	defer pool.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan struct{})
	err := pool.SubmitWait(ctx, func() {
		close(done)
	})

	if err != nil {
		t.Fatalf("submit wait failed: %v", err)
	}

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Error("task not executed")
	}
}

func TestWorker_SubmitTimeout(t *testing.T) {
	pool := NewWorker(1, 0) // 队列大小为 0
	defer pool.Stop()

	// 先占满
	_ = pool.Submit(func() {
		time.Sleep(time.Second)
	})

	// 应该超时
	err := pool.SubmitTimeout(func() {}, 10*time.Millisecond)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestWorker_PanicHandler(t *testing.T) {
	var panicCaught atomic.Bool

	pool := NewWorker(1, 10, WithPanicHandler(func(r any) {
		panicCaught.Store(true)
	}))
	defer pool.Stop()

	_ = pool.Submit(func() {
		panic("test panic")
	})

	time.Sleep(100 * time.Millisecond)

	if !panicCaught.Load() {
		t.Error("panic should be caught")
	}
}

func TestWorker_StopWait(t *testing.T) {
	pool := NewWorker(2, 10)

	var count atomic.Int32
	for i := 0; i < 5; i++ {
		_ = pool.Submit(func() {
			time.Sleep(10 * time.Millisecond)
			count.Add(1)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := pool.StopWait(ctx)
	if err != nil {
		t.Fatalf("stop wait failed: %v", err)
	}

	if count.Load() != 5 {
		t.Errorf("expected 5 tasks completed, got %d", count.Load())
	}
}

func TestWorker_Closed(t *testing.T) {
	pool := NewWorker(1, 10)
	pool.Stop()

	err := pool.Submit(func() {})
	if err != ErrPoolClosed {
		t.Errorf("expected ErrPoolClosed, got %v", err)
	}
}

func TestWorker_Stats(t *testing.T) {
	pool := NewWorker(2, 10)
	defer pool.Stop()

	for i := 0; i < 3; i++ {
		_ = pool.Submit(func() {})
	}

	time.Sleep(50 * time.Millisecond)

	stats := pool.Stats()
	if stats.Gets != 3 {
		t.Errorf("expected 3 submitted, got %d", stats.Gets)
	}
}

func BenchmarkObjectPool(b *testing.B) {
	pool := NewObject(
		func() *bytes.Buffer { return new(bytes.Buffer) },
		func(buf *bytes.Buffer) { buf.Reset() },
	)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			buf.WriteString("benchmark")
			pool.Put(buf)
		}
	})
}

func BenchmarkWorkerPool(b *testing.B) {
	pool := NewWorker(100, 1000)
	defer pool.Stop()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = pool.Submit(func() {})
		}
	})
}
