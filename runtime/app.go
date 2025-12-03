package runtime

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/mildsunup/higo/logger"
)

// App 应用实现
type App struct {
	cfg        Config
	log        Logger
	components []componentEntry
	hooks      struct {
		beforeStart []Hook
		afterStart  []Hook
		beforeStop  []Hook
		afterStop   []Hook
	}
	state atomic.Int32
	mu    sync.Mutex
}

// Option 应用选项
type Option func(*App)

// WithLogger 设置日志
func WithLogger(log Logger) Option {
	return func(a *App) {
		a.log = log
	}
}

// New 创建应用
func New(cfg Config, opts ...Option) *App {
	app := &App{
		cfg: cfg,
		log: logger.Nop(), // 默认空日志，避免 nil 判断
	}
	for _, opt := range opts {
		opt(app)
	}
	app.state.Store(int32(StateCreated))
	return app
}

// Register 注册组件（priority 越小越先启动，越后停止）
func (a *App) Register(c Component, priority int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.components = append(a.components, componentEntry{component: c, priority: priority})
}

// OnBeforeStart 注册启动前钩子
func (a *App) OnBeforeStart(h Hook) { a.hooks.beforeStart = append(a.hooks.beforeStart, h) }

// OnAfterStart 注册启动后钩子
func (a *App) OnAfterStart(h Hook) { a.hooks.afterStart = append(a.hooks.afterStart, h) }

// OnBeforeStop 注册停止前钩子
func (a *App) OnBeforeStop(h Hook) { a.hooks.beforeStop = append(a.hooks.beforeStop, h) }

// OnAfterStop 注册停止后钩子
func (a *App) OnAfterStop(h Hook) { a.hooks.afterStop = append(a.hooks.afterStop, h) }

// State 获取状态
func (a *App) State() State { return State(a.state.Load()) }

func (a *App) setState(s State) {
	old := State(a.state.Swap(int32(s)))
	a.log.Info(nil, "app state changed", logger.String("from", old.String()), logger.String("to", s.String()))
}

// Start 启动应用
func (a *App) Start(ctx context.Context) error {
	if a.State() != StateCreated && a.State() != StateStopped {
		return nil
	}

	a.setState(StateStarting)

	// 执行启动前钩子
	for _, h := range a.hooks.beforeStart {
		if err := h(ctx); err != nil {
			a.setState(StateFailed)
			return err
		}
	}

	// 按优先级排序
	a.mu.Lock()
	sort.Slice(a.components, func(i, j int) bool {
		return a.components[i].priority < a.components[j].priority
	})
	a.mu.Unlock()

	// 启动组件
	for i := range a.components {
		c := &a.components[i]
		a.log.Info(ctx, "starting component", logger.String("name", c.component.Name()))

		if err := c.component.Start(ctx); err != nil {
			a.setState(StateFailed)
			a.log.Error(ctx, "component start failed", logger.String("name", c.component.Name()), logger.Err(err))
			// 回滚已启动的组件
			a.stopStarted(ctx)
			return err
		}
		c.started = true
	}

	// 执行启动后钩子
	for _, h := range a.hooks.afterStart {
		if err := h(ctx); err != nil {
			a.setState(StateFailed)
			a.stopStarted(ctx)
			return err
		}
	}

	a.setState(StateRunning)
	a.log.Info(ctx, "app started", logger.String("name", a.cfg.Name))
	return nil
}

// Stop 停止应用
func (a *App) Stop(ctx context.Context) error {
	if a.State() != StateRunning && a.State() != StateFailed {
		return nil
	}

	a.setState(StateStopping)

	// 执行停止前钩子
	for _, h := range a.hooks.beforeStop {
		_ = h(ctx) // 忽略错误，继续停止
	}

	// 停止组件
	a.stopStarted(ctx)

	// 执行停止后钩子
	for _, h := range a.hooks.afterStop {
		_ = h(ctx)
	}

	a.setState(StateStopped)
	a.log.Info(ctx, "app stopped", logger.String("name", a.cfg.Name))
	return nil
}

func (a *App) stopStarted(ctx context.Context) {
	// 逆序停止
	for i := len(a.components) - 1; i >= 0; i-- {
		c := &a.components[i]
		if !c.started {
			continue
		}

		a.log.Info(ctx, "stopping component", logger.String("name", c.component.Name()))

		if err := c.component.Stop(ctx); err != nil {
			a.log.Error(ctx, "component stop failed", logger.String("name", c.component.Name()), logger.Err(err))
		}
		c.started = false
	}
}

// Run 运行应用（阻塞直到收到信号）
func (a *App) Run(ctx context.Context) error {
	if err := a.Start(ctx); err != nil {
		return err
	}

	// 等待信号
	sig := WaitSignal(ctx)
	a.log.Info(ctx, "received signal", logger.String("signal", sig.String()))

	// 带超时停止
	stopCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer cancel()

	return a.Stop(stopCtx)
}

// Health 健康检查
func (a *App) Health(ctx context.Context) error {
	if a.State() != StateRunning {
		return ErrNotRunning
	}

	for _, c := range a.components {
		if !c.started {
			continue
		}
		if hc, ok := c.component.(HealthChecker); ok {
			if err := hc.Health(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

var _ Application = (*App)(nil)
