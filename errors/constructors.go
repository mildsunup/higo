package errors

// --- 通用错误构造器 ---

// ErrInvalidArgument 无效参数
func ErrInvalidArgument(message string) *Error {
	return New(InvalidArgument, message)
}

// ErrNotFound 未找到
func ErrNotFound(message string) *Error {
	return New(NotFound, message)
}

// ErrAlreadyExists 已存在
func ErrAlreadyExists(message string) *Error {
	return New(AlreadyExists, message)
}

// ErrPermissionDenied 权限拒绝
func ErrPermissionDenied(message string) *Error {
	return New(PermissionDenied, message)
}

// ErrUnauthenticated 未认证
func ErrUnauthenticated(message string) *Error {
	return New(Unauthenticated, message)
}

// ErrInternal 内部错误
func ErrInternal(message string) *Error {
	return New(Internal, message)
}

// ErrUnavailable 服务不可用
func ErrUnavailable(message string) *Error {
	return New(Unavailable, message)
}

// ErrTimeout 超时
func ErrTimeout(message string) *Error {
	return New(Timeout, message)
}

// --- HTTP 兼容构造器 ---

// BadRequest 400 错误
func BadRequest(message string) *Error {
	return New(InvalidArgument, message)
}

// Unauthorized 401 错误
func Unauthorized(message string) *Error {
	return New(Unauthenticated, message)
}

// Forbidden 403 错误
func Forbidden(message string) *Error {
	return New(PermissionDenied, message)
}

// NotFoundErr 404 错误
func NotFoundErr(message string) *Error {
	return New(NotFound, message)
}

// Conflict 409 错误
func Conflict(message string) *Error {
	return New(AlreadyExists, message)
}

// InternalErr 500 错误
func InternalErr(err error) *Error {
	return Wrap(err, Internal, "internal server error")
}

// ServiceUnavailable 503 错误
func ServiceUnavailable(message string) *Error {
	return New(Unavailable, message)
}

// --- 业务错误构造器 ---

// ErrValidation 校验错误
func ErrValidation(message string) *Error {
	return New(ValidationFailed, message)
}

// ErrMissingParam 缺少参数
func ErrMissingParam(param string) *Error {
	return Newf(MissingParameter, "missing parameter: %s", param)
}

// ErrDatabase 数据库错误
func ErrDatabase(err error) *Error {
	return Wrap(err, DatabaseError, "database error")
}

// ErrCache 缓存错误
func ErrCache(err error) *Error {
	return Wrap(err, CacheError, "cache error")
}

// ErrMQ 消息队列错误
func ErrMQ(err error) *Error {
	return Wrap(err, MQError, "message queue error")
}

// ErrThirdParty 第三方服务错误
func ErrThirdParty(err error, service string) *Error {
	return Wrap(err, ThirdPartyError, "third party error: "+service)
}
