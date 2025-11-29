package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPoolClosed = errors.New("pool: closed")
	ErrPoolFull   = errors.New("pool: queue full")
)

// Worker 协程池
type Worker struct {
	maxWorkers   int32
	minWorkers   int32
	workers      atomic.Int32
	taskQueue    chan Task
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	closed       atomic.Bool
	panicHandler func(any)

	// 统计
	submitted atomic.Int64
	completed atomic.Int64
	failed    atomic.Int64
}

// WorkerOption 协程池选项
type WorkerOption func(*Worker)

// WithPanicHandler 设置 panic 处理器
func WithPanicHandler(h func(any)) WorkerOption {
	return func(w *Worker) { w.panicHandler = h }
}

// WithMinWorkers 设置最小 worker 数（预热）
func WithMinWorkers(n int) WorkerOption {
	return func(w *Worker) { w.minWorkers = int32(n) }
}

// NewWorker 创建协程池
func NewWorker(maxWorkers, queueSize int, opts ...WorkerOption) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Worker{
		maxWorkers: int32(maxWorkers),
		taskQueue:  make(chan Task, queueSize),
		ctx:        ctx,
		cancel:     cancel,
	}
	for _, opt := range opts {
		opt(w)
	}
	// 预热最小 worker
	w.warmup()
	return w
}

func (w *Worker) warmup() {
	for i := int32(0); i < w.minWorkers; i++ {
		w.spawnWorker()
	}
}

// Submit 提交任务（非阻塞）
func (w *Worker) Submit(task Task) error {
	if w.closed.Load() {
		return ErrPoolClosed
	}

	select {
	case w.taskQueue <- task:
		w.submitted.Add(1)
		w.maybeSpawnWorker()
		return nil
	default:
		return ErrPoolFull
	}
}

// SubmitWait 提交任务（阻塞等待）
func (w *Worker) SubmitWait(ctx context.Context, task Task) error {
	if w.closed.Load() {
		return ErrPoolClosed
	}

	select {
	case w.taskQueue <- task:
		w.submitted.Add(1)
		w.maybeSpawnWorker()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-w.ctx.Done():
		return ErrPoolClosed
	}
}

// SubmitTimeout 提交任务（带超时）
func (w *Worker) SubmitTimeout(task Task, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return w.SubmitWait(ctx, task)
}

// Go 提交任务（类似 go 关键字，忽略错误）
func (w *Worker) Go(task Task) {
	_ = w.Submit(task)
}

func (w *Worker) maybeSpawnWorker() {
	// 只在队列有积压且 worker 未满时 spawn
	if len(w.taskQueue) > 0 && w.workers.Load() < w.maxWorkers {
		w.spawnWorker()
	}
}

func (w *Worker) spawnWorker() {
	if w.workers.Add(1) > w.maxWorkers {
		w.workers.Add(-1)
		return
	}
	w.wg.Add(1)
	go w.worker()
}

func (w *Worker) worker() {
	defer func() {
		if r := recover(); r != nil {
			w.failed.Add(1)
			if w.panicHandler != nil {
				w.panicHandler(r)
			}
		}
		w.workers.Add(-1)
		w.wg.Done()
	}()

	idleTimeout := time.NewTimer(30 * time.Second)
	defer idleTimeout.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case task, ok := <-w.taskQueue:
			if !ok {
				return
			}
			w.runTask(task)
			// 重置空闲计时器
			if !idleTimeout.Stop() {
				select {
				case <-idleTimeout.C:
				default:
				}
			}
			idleTimeout.Reset(30 * time.Second)
		case <-idleTimeout.C:
			// 空闲超时，如果超过最小 worker 数则退出
			if w.workers.Load() > w.minWorkers {
				return
			}
			idleTimeout.Reset(30 * time.Second)
		}
	}
}

func (w *Worker) runTask(task Task) {
	defer func() {
		if r := recover(); r != nil {
			w.failed.Add(1)
			if w.panicHandler != nil {
				w.panicHandler(r)
			}
		}
	}()
	task()
	w.completed.Add(1)
}

// Running 运行中的 worker 数
func (w *Worker) Running() int {
	return int(w.workers.Load())
}

// Pending 等待中的任务数
func (w *Worker) Pending() int {
	return len(w.taskQueue)
}

// Stop 停止池（不等待）
func (w *Worker) Stop() {
	if w.closed.Swap(true) {
		return
	}
	w.cancel()
	close(w.taskQueue)
}

// StopWait 停止并等待所有任务完成
func (w *Worker) StopWait(ctx context.Context) error {
	if w.closed.Swap(true) {
		return nil
	}

	close(w.taskQueue)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.cancel()
		return nil
	case <-ctx.Done():
		w.cancel()
		return ctx.Err()
	}
}

// Stats 返回统计信息
func (w *Worker) Stats() Stats {
	return Stats{
		Gets:    w.submitted.Load(),
		Puts:    w.completed.Load(),
		Misses:  w.failed.Load(),
		InUse:   int64(w.workers.Load()),
		Idle:    int64(len(w.taskQueue)),
		MaxSize: int64(w.maxWorkers),
	}
}

var _ TaskPool = (*Worker)(nil)
