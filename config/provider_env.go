package config

import (
	"context"
	"os"
	"strings"
)

// EnvProvider 环境变量配置提供者
type EnvProvider struct {
	prefix    string
	separator string
}

// NewEnvProvider 创建环境变量配置提供者
// prefix: 环境变量前缀，如 "APP" 则读取 APP_* 变量
// separator: 层级分隔符，如 "_" 则 APP_SERVER_PORT 映射到 server.port
func NewEnvProvider(prefix string) *EnvProvider {
	return &EnvProvider{
		prefix:    strings.ToUpper(prefix),
		separator: "_",
	}
}

func (p *EnvProvider) Name() string {
	return "env:" + p.prefix
}

func (p *EnvProvider) Load(ctx context.Context) (map[string]any, error) {
	result := make(map[string]any)
	prefix := p.prefix + p.separator

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		// 移除前缀并转换为小写
		key = strings.ToLower(strings.TrimPrefix(key, prefix))
		// 将分隔符转换为嵌套结构
		setNestedValue(result, strings.Split(key, strings.ToLower(p.separator)), value)
	}

	return result, nil
}

func (p *EnvProvider) Watch(ctx context.Context, onChange func()) error {
	// 环境变量不支持监听
	return nil
}

// setNestedValue 设置嵌套值
func setNestedValue(data map[string]any, keys []string, value any) {
	for i, key := range keys {
		if i == len(keys)-1 {
			data[key] = value
			return
		}

		if _, ok := data[key]; !ok {
			data[key] = make(map[string]any)
		}

		if next, ok := data[key].(map[string]any); ok {
			data = next
		} else {
			return
		}
	}
}
