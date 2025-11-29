// Package runtime 提供应用生命周期管理。
//
// 核心功能：
//   - 组件启动/停止顺序管理（按优先级）
//   - 生命周期钩子（BeforeStart/AfterStart/BeforeStop/AfterStop）
//   - 信号处理和优雅关闭
//
// 使用示例：
//
//	app := runtime.New(cfg, logger)
//	app.Register(httpServer, 100)
//	app.Register(grpcServer, 200)
//	if err := app.Run(ctx); err != nil {
//	    log.Fatal(err)
//	}
package runtime
