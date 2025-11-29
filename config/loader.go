package config

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
)

// Loader 配置加载器
type Loader struct {
	mu       sync.RWMutex
	opts     Options
	data     map[string]any
	watchers []context.CancelFunc
}

// NewLoader 创建配置加载器
func NewLoader(opts ...Option) *Loader {
	o := Options{}
	for _, opt := range opts {
		opt(&o)
	}
	return &Loader{
		opts: o,
		data: make(map[string]any),
	}
}

// Load 加载配置到目标结构
func (l *Loader) Load(ctx context.Context, target any) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 合并所有 Provider 的配置
	merged := make(map[string]any)
	for _, p := range l.opts.Providers {
		data, err := p.Load(ctx)
		if err != nil {
			return fmt.Errorf("provider %s: %w", p.Name(), err)
		}
		merged = mergeMaps(merged, data)
	}

	// 解析敏感信息
	if l.opts.SecretResolver != nil {
		if err := l.resolveSecrets(ctx, merged); err != nil {
			return fmt.Errorf("resolve secrets: %w", err)
		}
	}

	l.data = merged

	// 解码到目标结构
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           target,
		TagName:          "yaml",
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	})
	if err != nil {
		return fmt.Errorf("create decoder: %w", err)
	}

	if err := decoder.Decode(merged); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	// 校验配置
	if v, ok := target.(Validator); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validate config: %w", err)
		}
	}

	return nil
}

// Watch 启动配置监听
func (l *Loader) Watch(ctx context.Context, target any) error {
	for _, p := range l.opts.Providers {
		watchCtx, cancel := context.WithCancel(ctx)
		l.watchers = append(l.watchers, cancel)

		go func(p Provider) {
			_ = p.Watch(watchCtx, func() {
				if err := l.Load(ctx, target); err == nil && l.opts.OnChange != nil {
					l.opts.OnChange(ChangeEvent{})
				}
			})
		}(p)
	}
	return nil
}

// Stop 停止监听
func (l *Loader) Stop() {
	for _, cancel := range l.watchers {
		cancel()
	}
	l.watchers = nil
}

// Get 获取配置值
func (l *Loader) Get(key string) any {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return getNestedValue(l.data, key)
}

// GetString 获取字符串配置
func (l *Loader) GetString(key string) string {
	if v := l.Get(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt 获取整数配置
func (l *Loader) GetInt(key string) int {
	if v := l.Get(key); v != nil {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		}
	}
	return 0
}

// GetBool 获取布尔配置
func (l *Loader) GetBool(key string) bool {
	if v := l.Get(key); v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// resolveSecrets 解析配置中的敏感信息
func (l *Loader) resolveSecrets(ctx context.Context, data map[string]any) error {
	return walkMap(data, func(key string, value any) (any, error) {
		if s, ok := value.(string); ok && isSecretRef(s) {
			resolved, err := l.opts.SecretResolver.Resolve(ctx, s)
			if err != nil {
				return nil, fmt.Errorf("resolve %s: %w", key, err)
			}
			return resolved, nil
		}
		return value, nil
	})
}

// isSecretRef 检查是否为敏感信息引用
func isSecretRef(s string) bool {
	return strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}")
}

// mergeMaps 深度合并 map
func mergeMaps(base, override map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		if baseMap, ok := result[k].(map[string]any); ok {
			if overrideMap, ok := v.(map[string]any); ok {
				result[k] = mergeMaps(baseMap, overrideMap)
				continue
			}
		}
		result[k] = v
	}
	return result
}

// getNestedValue 获取嵌套值，支持 "a.b.c" 格式
func getNestedValue(data map[string]any, key string) any {
	parts := strings.Split(key, ".")
	current := any(data)

	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return nil
		}
	}
	return current
}

// walkMap 遍历 map 并转换值
func walkMap(data map[string]any, fn func(key string, value any) (any, error)) error {
	for k, v := range data {
		switch val := v.(type) {
		case map[string]any:
			if err := walkMap(val, fn); err != nil {
				return err
			}
		case []any:
			for i, item := range val {
				if m, ok := item.(map[string]any); ok {
					if err := walkMap(m, fn); err != nil {
						return err
					}
				} else {
					newVal, err := fn(fmt.Sprintf("%s[%d]", k, i), item)
					if err != nil {
						return err
					}
					val[i] = newVal
				}
			}
		default:
			newVal, err := fn(k, v)
			if err != nil {
				return err
			}
			data[k] = newVal
		}
	}
	return nil
}

// EnvSecretResolver 环境变量解析器
type EnvSecretResolver struct {
	prefix string
}

// NewEnvSecretResolver 创建环境变量解析器
func NewEnvSecretResolver(prefix string) *EnvSecretResolver {
	return &EnvSecretResolver{prefix: prefix}
}

var envRefPattern = regexp.MustCompile(`\$\{env:([^}]+)\}`)

func (r *EnvSecretResolver) Resolve(ctx context.Context, value string) (string, error) {
	matches := envRefPattern.FindStringSubmatch(value)
	if len(matches) < 2 {
		return value, nil
	}

	envKey := matches[1]
	if r.prefix != "" {
		envKey = r.prefix + "_" + envKey
	}

	envVal := lookupEnv(envKey)
	return envRefPattern.ReplaceAllString(value, envVal), nil
}

// lookupEnv 查找环境变量（支持默认值 VAR:default）
func lookupEnv(key string) string {
	parts := strings.SplitN(key, ":", 2)
	val := os.Getenv(parts[0])
	if val == "" && len(parts) > 1 {
		return parts[1]
	}
	return val
}
