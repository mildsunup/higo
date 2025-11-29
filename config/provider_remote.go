package config

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// RemoteType 远程配置类型
type RemoteType string

const (
	RemoteConsul RemoteType = "consul"
	RemoteEtcd   RemoteType = "etcd"
	RemoteEtcd3  RemoteType = "etcd3"
)

// RemoteProvider 远程配置提供者
type RemoteProvider struct {
	typ           RemoteType
	endpoint      string
	path          string
	configType    string
	watchInterval time.Duration
}

// RemoteOption 远程提供者选项
type RemoteOption func(*RemoteProvider)

// WithConfigType 设置配置类型
func WithConfigType(t string) RemoteOption {
	return func(p *RemoteProvider) {
		p.configType = t
	}
}

// WithRemoteWatchInterval 设置监听间隔
func WithRemoteWatchInterval(d time.Duration) RemoteOption {
	return func(p *RemoteProvider) {
		p.watchInterval = d
	}
}

// NewRemoteProvider 创建远程配置提供者
func NewRemoteProvider(typ RemoteType, endpoint, path string, opts ...RemoteOption) *RemoteProvider {
	p := &RemoteProvider{
		typ:           typ,
		endpoint:      endpoint,
		path:          path,
		configType:    "yaml",
		watchInterval: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *RemoteProvider) Name() string {
	return fmt.Sprintf("%s:%s%s", p.typ, p.endpoint, p.path)
}

func (p *RemoteProvider) Load(ctx context.Context) (map[string]any, error) {
	v := viper.New()
	v.SetConfigType(p.configType)

	if err := v.AddRemoteProvider(string(p.typ), p.endpoint, p.path); err != nil {
		return nil, fmt.Errorf("add remote provider: %w", err)
	}

	if err := v.ReadRemoteConfig(); err != nil {
		return nil, fmt.Errorf("read remote config: %w", err)
	}

	return v.AllSettings(), nil
}

func (p *RemoteProvider) Watch(ctx context.Context, onChange func()) error {
	ticker := time.NewTicker(p.watchInterval)
	defer ticker.Stop()

	var lastData map[string]any

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			data, err := p.Load(ctx)
			if err != nil {
				continue
			}

			if !mapsEqual(lastData, data) {
				lastData = data
				onChange()
			}
		}
	}
}

// mapsEqual 简单比较两个 map 是否相等
func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || fmt.Sprintf("%v", v) != fmt.Sprintf("%v", bv) {
			return false
		}
	}
	return true
}
