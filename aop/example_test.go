package aop_test

import (
	"context"
	"fmt"
	"time"

	"github.com/mildsunup/higo/aop"
)

// UserService 示例服务
type UserService struct{}

func (s *UserService) GetUser(ctx context.Context, id int64) (string, error) {
	return fmt.Sprintf("user-%d", id), nil
}

func Example_basicUsage() {
	service := &UserService{}

	// 创建拦截器链
	chain := aop.NewChain(
		aop.InterceptorFunc(func(inv *aop.Invocation) error {
			fmt.Println("before:", inv.Method())
			err := inv.Proceed()
			fmt.Println("after:", inv.Method())
			return err
		}),
	)

	// 执行
	inv, _ := chain.ExecuteWithTarget(
		context.Background(),
		service,
		"GetUser",
		func(inv *aop.Invocation) error {
			result, err := service.GetUser(inv.Context(), inv.Args()[0].(int64))
			inv.SetResult(result)
			return err
		},
		int64(123),
	)

	fmt.Println("result:", inv.Result())
	// Output:
	// before: GetUser
	// after: GetUser
	// result: user-123
}

func Example_genericHandler() {
	// 定义处理器
	handler := func(ctx context.Context, id int64) (string, error) {
		return fmt.Sprintf("user-%d", id), nil
	}

	// 创建拦截器链
	chain := aop.NewChain(
		aop.Metrics(func(method string, d time.Duration, success bool) {
			fmt.Printf("metrics: %s success=%v\n", method, success)
		}),
	)

	// 包装处理器
	wrapped := aop.Wrap(chain, "GetUser", handler)

	// 调用
	result, _ := wrapped(context.Background(), 456)
	fmt.Println("result:", result)
	// Output:
	// metrics: GetUser success=true
	// result: user-456
}
