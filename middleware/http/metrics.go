package http

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"higo/observability"
)

// Metrics HTTP 指标
type Metrics struct {
	RequestsTotal   observability.Counter
	RequestDuration observability.Histogram
	RequestSize     observability.Histogram
	ResponseSize    observability.Histogram
	ActiveRequests  observability.Gauge
}

// NewMetrics 创建 HTTP 指标
func NewMetrics(p observability.MetricsProvider) *Metrics {
	return &Metrics{
		RequestsTotal:   p.Counter("http_requests_total", "Total HTTP requests", "method", "path", "status"),
		RequestDuration: p.Histogram("http_request_duration_seconds", "HTTP request duration", observability.DurationBuckets, "method", "path"),
		RequestSize:     p.Histogram("http_request_size_bytes", "HTTP request size", observability.SizeBuckets, "method", "path"),
		ResponseSize:    p.Histogram("http_response_size_bytes", "HTTP response size", observability.SizeBuckets, "method", "path"),
		ActiveRequests:  p.Gauge("http_active_requests", "Active HTTP requests"),
	}
}

// Middleware 返回 Gin 指标中间件
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		m.ActiveRequests.Inc()
		defer m.ActiveRequests.Dec()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		m.RequestsTotal.Inc(method, path, status)
		m.RequestDuration.Since(start, method, path)
		m.RequestSize.Observe(float64(c.Request.ContentLength), method, path)
		m.ResponseSize.Observe(float64(c.Writer.Size()), method, path)
	}
}
