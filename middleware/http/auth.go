package http

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"higo/response"
)

// TokenValidator Token 验证函数
type TokenValidator func(ctx context.Context, token string) (userID uint64, err error)

// BearerAuth Bearer Token 认证中间件
func BearerAuth(validate TokenValidator) gin.HandlerFunc {
	return BearerAuthWithKey(validate, "user_id")
}

// BearerAuthWithKey Bearer Token 认证中间件（自定义 key）
func BearerAuthWithKey(validate TokenValidator, userIDKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, response.Response[any]{Code: 401, Message: "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(401, response.Response[any]{Code: 401, Message: "invalid authorization format"})
			c.Abort()
			return
		}

		userID, err := validate(c.Request.Context(), parts[1])
		if err != nil {
			c.JSON(401, response.Response[any]{Code: 401, Message: "invalid token"})
			c.Abort()
			return
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

// GetUserID 从 Gin 上下文获取用户 ID
func GetUserID(c *gin.Context) (uint64, bool) {
	return GetUserIDWithKey(c, "user_id")
}

// GetUserIDWithKey 从 Gin 上下文获取用户 ID（自定义 key）
func GetUserIDWithKey(c *gin.Context, key string) (uint64, bool) {
	v, exists := c.Get(key)
	if !exists {
		return 0, false
	}
	userID, ok := v.(uint64)
	return userID, ok
}
