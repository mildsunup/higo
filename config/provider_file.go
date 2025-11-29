package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// FileProvider 文件配置提供者
type FileProvider struct {
	path     string
	optional bool
}

// FileOption 文件提供者选项
type FileOption func(*FileProvider)

// WithOptional 设置为可选文件
func WithOptional() FileOption {
	return func(p *FileProvider) {
		p.optional = true
	}
}

// NewFileProvider 创建文件配置提供者
func NewFileProvider(path string, opts ...FileOption) *FileProvider {
	p := &FileProvider{path: path}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *FileProvider) Name() string {
	return "file:" + p.path
}

func (p *FileProvider) Load(ctx context.Context) (map[string]any, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		if os.IsNotExist(err) && p.optional {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("read file %s: %w", p.path, err)
	}

	var result map[string]any
	ext := filepath.Ext(p.path)

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
	case ".json":
		if err := yaml.Unmarshal(data, &result); err != nil { // yaml 兼容 json
			return nil, fmt.Errorf("parse json: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	return result, nil
}

func (p *FileProvider) Watch(ctx context.Context, onChange func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	go func() {
		defer watcher.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					// 防抖
					time.Sleep(100 * time.Millisecond)
					onChange()
				}
			case <-watcher.Errors:
				// 忽略错误，继续监听
			}
		}
	}()

	return watcher.Add(p.path)
}
