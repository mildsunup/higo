package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileProvider_Load(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
app:
  name: test-app
  env: testing
server:
  http:
    enabled: true
    port: "8080"
    mode: debug
logger:
  level: info
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	provider := NewFileProvider(configPath)
	data, err := provider.Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	app, ok := data["app"].(map[string]any)
	if !ok {
		t.Fatal("app config not found")
	}

	if app["name"] != "test-app" {
		t.Errorf("expected app.name = test-app, got %v", app["name"])
	}
}

func TestEnvProvider_Load(t *testing.T) {
	// 设置环境变量
	os.Setenv("TEST_APP_NAME", "env-app")
	os.Setenv("TEST_SERVER_HTTP_PORT", "9090")
	defer func() {
		os.Unsetenv("TEST_APP_NAME")
		os.Unsetenv("TEST_SERVER_HTTP_PORT")
	}()

	provider := NewEnvProvider("TEST")
	data, err := provider.Load(context.Background())
	if err != nil {
		t.Fatalf("load env config: %v", err)
	}

	app, ok := data["app"].(map[string]any)
	if !ok {
		t.Fatal("app config not found")
	}

	if app["name"] != "env-app" {
		t.Errorf("expected app.name = env-app, got %v", app["name"])
	}
}

func TestLoader_MergeProviders(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
app:
  name: file-app
  env: development
server:
  http:
    enabled: true
    port: "8080"
    mode: debug
logger:
  level: info
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	// 设置环境变量覆盖
	os.Setenv("MERGE_APP_NAME", "env-override")
	defer os.Unsetenv("MERGE_APP_NAME")

	loader := NewLoader(
		WithProvider(NewFileProvider(configPath)),
		WithProvider(NewEnvProvider("MERGE")),
	)

	var cfg Config
	if err := loader.Load(context.Background(), &cfg); err != nil {
		t.Fatalf("load config: %v", err)
	}

	// 环境变量应该覆盖文件配置
	if cfg.App.Name != "env-override" {
		t.Errorf("expected app.name = env-override, got %s", cfg.App.Name)
	}

	// 文件配置应该保留
	if cfg.App.Env != "development" {
		t.Errorf("expected app.env = development, got %s", cfg.App.Env)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				App:    AppConfig{Name: "test"},
				Server: ServerConfig{HTTP: HTTPConfig{Enabled: true, Port: "8080"}},
				Logger: LoggerConfig{Level: "info"},
			},
			wantErr: false,
		},
		{
			name: "missing app name",
			cfg: Config{
				Server: ServerConfig{HTTP: HTTPConfig{Enabled: true, Port: "8080"}},
			},
			wantErr: true,
		},
		{
			name: "no server enabled",
			cfg: Config{
				App: AppConfig{Name: "test"},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: Config{
				App:    AppConfig{Name: "test"},
				Server: ServerConfig{HTTP: HTTPConfig{Enabled: true, Port: "8080"}},
				Logger: LoggerConfig{Level: "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoader_Get(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
app:
  name: test
  env: dev
server:
  http:
    enabled: true
    port: "8080"
logger:
  level: info
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	loader := NewLoader(WithProvider(NewFileProvider(configPath)))

	var cfg Config
	if err := loader.Load(context.Background(), &cfg); err != nil {
		t.Fatalf("load config: %v", err)
	}

	// 测试 Get 方法
	if loader.GetString("app.name") != "test" {
		t.Errorf("expected app.name = test")
	}

	if loader.GetString("server.http.port") != "8080" {
		t.Errorf("expected server.http.port = 8080")
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.App.Name != "higo" {
		t.Errorf("expected default app.name = higo")
	}

	if !cfg.Server.HTTP.Enabled {
		t.Error("expected HTTP server enabled by default")
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}
}
