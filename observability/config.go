// Package observability 提供统一的可观测性支持
package observability

// Config 可观测性配置
type Config struct {
	ServiceName    string        `yaml:"service_name" mapstructure:"service_name"`
	ServiceVersion string        `yaml:"service_version" mapstructure:"service_version"`
	Environment    string        `yaml:"environment" mapstructure:"environment"`
	Tracing        TracingConfig `yaml:"tracing" mapstructure:"tracing"`
	Metrics        MetricsConfig `yaml:"metrics" mapstructure:"metrics"`
}

// TracingConfig 追踪配置
type TracingConfig struct {
	Enabled  bool    `yaml:"enabled" mapstructure:"enabled"`
	Exporter string  `yaml:"exporter" mapstructure:"exporter"` // otlp, otlp-http, stdout, noop
	Endpoint string  `yaml:"endpoint" mapstructure:"endpoint"`
	Insecure bool    `yaml:"insecure" mapstructure:"insecure"`
	Sampler  string  `yaml:"sampler" mapstructure:"sampler"` // always, never, ratio, parent
	Ratio    float64 `yaml:"ratio" mapstructure:"ratio"`     // 采样率 0.0-1.0
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Path    string `yaml:"path" mapstructure:"path"` // /metrics
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		ServiceName: "unknown",
		Environment: "development",
		Tracing: TracingConfig{
			Enabled:  false,
			Exporter: "otlp",
			Sampler:  "ratio",
			Ratio:    0.1,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
}
