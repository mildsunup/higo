package http

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mildsunup/higo/logger"
)

// Logging 日志中间件
func Logging(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		ctx := c.Request.Context()

		fields := []logger.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.String("query", query),
			logger.Int("status", status),
			logger.Duration("latency", latency),
			logger.String("client_ip", c.ClientIP()),
			logger.Int("body_size", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, logger.String("error", c.Errors.String()))
		}

		// trace_id/span_id 由 logger 自动从 ctx 提取
		switch {
		case status >= 500:
			log.Error(ctx, "HTTP request", fields...)
		case status >= 400:
			log.Warn(ctx, "HTTP request", fields...)
		default:
			log.Info(ctx, "HTTP request", fields...)
		}
	}
}
