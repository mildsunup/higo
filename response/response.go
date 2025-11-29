package response

// Response 标准响应结构
type Response[T any] struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      T      `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
}

// PageResponse 分页响应
type PageResponse[T any] struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      T      `json:"data"`
	Total     int64  `json:"total"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	RequestID string `json:"request_id,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
}

// OK 创建成功响应
func OK[T any](data T) Response[T] {
	return Response[T]{
		Code:    0,
		Message: "ok",
		Data:    data,
	}
}

// OKWithMessage 创建带消息的成功响应
func OKWithMessage[T any](data T, message string) Response[T] {
	return Response[T]{
		Code:    0,
		Message: message,
		Data:    data,
	}
}

// Page 创建分页响应
func Page[T any](data T, total int64, page, pageSize int) PageResponse[T] {
	return PageResponse[T]{
		Code:     0,
		Message:  "ok",
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// WithRequestID 设置请求 ID
func (r Response[T]) WithRequestID(id string) Response[T] {
	r.RequestID = id
	return r
}

// WithTraceID 设置追踪 ID
func (r Response[T]) WithTraceID(id string) Response[T] {
	r.TraceID = id
	return r
}

// WithRequestID 设置请求 ID（分页）
func (r PageResponse[T]) WithRequestID(id string) PageResponse[T] {
	r.RequestID = id
	return r
}

// WithTraceID 设置追踪 ID（分页）
func (r PageResponse[T]) WithTraceID(id string) PageResponse[T] {
	r.TraceID = id
	return r
}
