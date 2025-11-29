package errors

import (
	"encoding/json"
)

// Response API 错误响应
type Response struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ToResponse 转换为 API 响应
func ToResponse(err error) Response {
	if err == nil {
		return Response{Code: 0, Message: "success"}
	}

	var e *Error
	if As(err, &e) {
		return Response{
			Code:    int(e.code),
			Message: e.message,
			Details: e.metadata,
		}
	}

	return Response{
		Code:    int(Unknown),
		Message: err.Error(),
	}
}

// MarshalJSON 序列化错误为 JSON
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(Response{
		Code:    int(e.code),
		Message: e.message,
		Details: e.metadata,
	})
}

// LogFields 返回结构化日志字段
func (e *Error) LogFields() map[string]any {
	fields := map[string]any{
		"error_code":    int(e.code),
		"error_message": e.message,
	}

	if e.cause != nil {
		fields["error_cause"] = e.cause.Error()
	}

	if e.metadata != nil {
		fields["error_metadata"] = e.metadata
	}

	return fields
}
