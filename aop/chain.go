package aop

import "context"

// Chain 拦截器链
type Chain struct {
	interceptors []Interceptor
}

// NewChain 创建拦截器链
func NewChain(interceptors ...Interceptor) *Chain {
	return &Chain{interceptors: interceptors}
}

// Use 添加拦截器
func (c *Chain) Use(interceptors ...Interceptor) *Chain {
	c.interceptors = append(c.interceptors, interceptors...)
	return c
}

// Clone 克隆拦截器链
func (c *Chain) Clone() *Chain {
	cloned := make([]Interceptor, len(c.interceptors))
	copy(cloned, c.interceptors)
	return &Chain{interceptors: cloned}
}

// Execute 执行拦截器链
// target: 目标方法，作为链的最后一环执行
func (c *Chain) Execute(ctx context.Context, method string, target func(*Invocation) error, args ...any) (*Invocation, error) {
	inv := &Invocation{
		ctx:    ctx,
		method: method,
		args:   args,
	}
	// 将 target 作为最后一个拦截器
	inv.chain = append(c.interceptors, InterceptorFunc(target))
	err := inv.Proceed()
	return inv, err
}

// ExecuteWithTarget 执行拦截器链（带目标对象）
func (c *Chain) ExecuteWithTarget(ctx context.Context, target any, method string, fn func(*Invocation) error, args ...any) (*Invocation, error) {
	inv := &Invocation{
		ctx:    ctx,
		method: method,
		target: target,
		args:   args,
	}
	inv.chain = append(c.interceptors, InterceptorFunc(fn))
	err := inv.Proceed()
	return inv, err
}

// Handler 泛型处理器类型
type Handler[Req, Resp any] func(ctx context.Context, req Req) (Resp, error)

// Wrap 包装泛型处理器，应用拦截器链
func Wrap[Req, Resp any](chain *Chain, method string, h Handler[Req, Resp]) Handler[Req, Resp] {
	return func(ctx context.Context, req Req) (Resp, error) {
		var resp Resp
		inv, err := chain.Execute(ctx, method, func(inv *Invocation) error {
			r, e := h(inv.Context(), req)
			resp = r
			inv.SetResult(r)
			inv.SetError(e)
			return e
		}, req)
		if err != nil {
			return resp, err
		}
		if inv.Error() != nil {
			return resp, inv.Error()
		}
		if inv.Result() != nil {
			return inv.Result().(Resp), nil
		}
		return resp, nil
	}
}
