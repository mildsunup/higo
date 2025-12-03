package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mildsunup/higo/config"
	"github.com/mildsunup/higo/logger"
	"github.com/mildsunup/higo/response"
	"github.com/mildsunup/higo/runtime"
	"github.com/mildsunup/higo/server"
)

func main() {
	// 初始化日志
	log, _ := logger.New(logger.Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	})

	// 创建 HTTP Handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, response.OK(map[string]string{"message": "Hello, Higo!"}))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, response.OK(map[string]string{"status": "ok"}))
	})

	// 创建 HTTP 服务器
	httpServer := server.NewHTTPServer(mux,
		server.WithName("http"),
		server.WithAddr(":8080"),
	)

	// 创建应用
	app := runtime.New(runtime.Config{
		Name:            "basic-example",
		ShutdownTimeout: 10 * time.Second,
	}, runtime.WithLogger(log))

	// 注册组件
	app.Register(httpServer, 100)

	// 生命周期钩子
	app.OnAfterStart(func(ctx context.Context) error {
		log.Info(ctx, "server listening on :8080")
		return nil
	})

	// 启动应用
	ctx := context.Background()
	if err := app.Run(ctx); err != nil {
		log.Error(ctx, "app failed", logger.Err(err))
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// 使用配置文件的示例
func withConfig() {
	cfg := config.Default()

	log, _ := logger.New(logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
		Output: cfg.Logger.OutputPath,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, response.OK(map[string]any{
			"app":     cfg.App.Name,
			"version": cfg.App.Version,
		}))
	})

	httpServer := server.NewHTTPServer(mux,
		server.WithName("http"),
		server.WithAddr(":"+cfg.Server.HTTP.Port),
		server.WithReadTimeout(cfg.Server.HTTP.ReadTimeout),
		server.WithWriteTimeout(cfg.Server.HTTP.WriteTimeout),
	)

	app := runtime.New(runtime.Config{
		Name:            cfg.App.Name,
		ShutdownTimeout: 30 * time.Second,
	}, runtime.WithLogger(log))

	app.Register(httpServer, 100)

	_ = app.Run(context.Background())
}
