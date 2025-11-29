package runtime

import "context"

// componentEntry 组件条目
type componentEntry struct {
	component Component
	priority  int
	started   bool
}

// FuncComponent 函数式组件
type FuncComponent struct {
	name    string
	startFn func(ctx context.Context) error
	stopFn  func(ctx context.Context) error
}

// NewFuncComponent 创建函数式组件
func NewFuncComponent(name string, start, stop func(ctx context.Context) error) *FuncComponent {
	return &FuncComponent{name: name, startFn: start, stopFn: stop}
}

func (c *FuncComponent) Name() string                      { return c.name }
func (c *FuncComponent) Start(ctx context.Context) error   { return c.startFn(ctx) }
func (c *FuncComponent) Stop(ctx context.Context) error    { return c.stopFn(ctx) }

// ServerComponent 服务器组件（异步启动）
type ServerComponent struct {
	name    string
	startFn func() error
	stopFn  func(ctx context.Context) error
	errCh   chan error
}

// NewServerComponent 创建服务器组件
func NewServerComponent(name string, start func() error, stop func(ctx context.Context) error) *ServerComponent {
	return &ServerComponent{name: name, startFn: start, stopFn: stop, errCh: make(chan error, 1)}
}

func (c *ServerComponent) Name() string { return c.name }

func (c *ServerComponent) Start(ctx context.Context) error {
	go func() {
		if err := c.startFn(); err != nil {
			c.errCh <- err
		}
	}()
	return nil
}

func (c *ServerComponent) Stop(ctx context.Context) error {
	return c.stopFn(ctx)
}

// Err 返回错误通道
func (c *ServerComponent) Err() <-chan error {
	return c.errCh
}
