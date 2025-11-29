package runtime

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WaitSignal 等待中断信号
func WaitSignal(ctx context.Context) os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		return sig
	case <-ctx.Done():
		return nil
	}
}
