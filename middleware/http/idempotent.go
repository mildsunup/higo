package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mildsunup/higo/idempotent"
)

const (
	HeaderIdempotencyKey = "X-Idempotency-Key"
	HeaderIdempotentHit  = "X-Idempotent-Hit"
)

// IdempotentConfig 幂等中间件配置
type IdempotentConfig struct {
	Handler          *idempotent.Handler
	Methods          []string // 需要幂等的方法，默认 POST, PUT, PATCH
	KeyFunc          func(*gin.Context) string
	CheckFingerprint bool
}

// DefaultIdempotentConfig 默认配置
func DefaultIdempotentConfig(h *idempotent.Handler) IdempotentConfig {
	return IdempotentConfig{
		Handler:          h,
		Methods:          []string{http.MethodPost, http.MethodPut, http.MethodPatch},
		CheckFingerprint: true,
		KeyFunc: func(c *gin.Context) string {
			return c.GetHeader(HeaderIdempotencyKey)
		},
	}
}

// Idempotent 幂等中间件
func Idempotent(cfg IdempotentConfig) gin.HandlerFunc {
	methods := make(map[string]bool)
	for _, m := range cfg.Methods {
		methods[m] = true
	}

	return func(c *gin.Context) {
		// 只处理指定方法
		if !methods[c.Request.Method] {
			c.Next()
			return
		}

		key := cfg.KeyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		// 计算指纹
		var fingerprint string
		if cfg.CheckFingerprint {
			body, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			fingerprint = idempotent.Fingerprint(c.Request.Method, c.Request.URL.Path, body)
		}

		// 使用响应捕获器
		w := &responseCapture{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = w

		result, err := cfg.Handler.Execute(c.Request.Context(), key, fingerprint, func() ([]byte, error) {
			c.Next()

			if len(c.Errors) > 0 {
				return nil, c.Errors.Last()
			}
			return w.body.Bytes(), nil
		})

		if err != nil {
			if err == idempotent.ErrDuplicateRequest {
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{
					"error": "duplicate request",
					"code":  "DUPLICATE_REQUEST",
				})
				return
			}
			// 其他错误已在 fn 中处理
			return
		}

		// 缓存命中，返回缓存响应
		if result.Cached {
			c.Header(HeaderIdempotentHit, "true")
			c.Writer = w.ResponseWriter
			c.Data(w.status, w.ResponseWriter.Header().Get("Content-Type"), result.Response)
			c.Abort()
		}
	}
}

type responseCapture struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (r *responseCapture) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseCapture) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
