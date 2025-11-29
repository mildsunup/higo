package http

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"higo/logger"
)

// Recovery panic 恢复中间件
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				ctx := c.Request.Context()

				// trace_id/span_id 由 logger 自动从 ctx 提取
				log.Error(ctx, "HTTP panic recovered",
					logger.Any("panic", r),
					logger.String("stack", string(debug.Stack())),
					logger.String("method", c.Request.Method),
					logger.String("path", c.Request.URL.Path),
					logger.String("client_ip", c.ClientIP()),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}
