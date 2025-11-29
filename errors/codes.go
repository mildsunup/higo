package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

// Code 错误码
type Code int

// 错误码定义 (按模块划分区间)
// 0: 成功
// 1-999: 通用错误
// 1000-1999: 认证/授权
// 2000-2999: 参数校验
// 3000-3999: 业务错误
// 4000-4999: 外部服务
// 5000-5999: 系统错误
const (
	// 成功
	OK Code = 0

	// 通用错误 1-999
	Unknown         Code = 1
	Canceled        Code = 2
	InvalidArgument Code = 3
	NotFound        Code = 4
	AlreadyExists   Code = 5
	PermissionDenied Code = 6
	Unauthenticated Code = 7
	ResourceExhausted Code = 8
	FailedPrecondition Code = 9
	Aborted         Code = 10
	OutOfRange      Code = 11
	Unimplemented   Code = 12
	Internal        Code = 13
	Unavailable     Code = 14
	DataLoss        Code = 15
	Timeout         Code = 16

	// 认证/授权 1000-1999
	TokenExpired    Code = 1001
	TokenInvalid    Code = 1002
	TokenMissing    Code = 1003
	AccessDenied    Code = 1004
	SessionExpired  Code = 1005

	// 参数校验 2000-2999
	ValidationFailed Code = 2001
	MissingParameter Code = 2002
	InvalidFormat    Code = 2003
	OutOfBounds      Code = 2004

	// 业务错误 3000-3999
	UserNotFound     Code = 3001
	UserExists       Code = 3002
	OrderNotFound    Code = 3003
	InsufficientBalance Code = 3004

	// 外部服务 4000-4999
	DatabaseError    Code = 4001
	CacheError       Code = 4002
	MQError          Code = 4003
	ThirdPartyError  Code = 4004
	NetworkError     Code = 4005

	// 系统错误 5000-5999
	SystemError      Code = 5001
	ConfigError      Code = 5002
	InitError        Code = 5003
)

// codeInfo 错误码信息
type codeInfo struct {
	httpStatus int
	grpcCode   codes.Code
	message    string
}

var codeRegistry = map[Code]codeInfo{
	OK:                 {http.StatusOK, codes.OK, "success"},
	Unknown:            {http.StatusInternalServerError, codes.Unknown, "unknown error"},
	Canceled:           {http.StatusRequestTimeout, codes.Canceled, "request canceled"},
	InvalidArgument:    {http.StatusBadRequest, codes.InvalidArgument, "invalid argument"},
	NotFound:           {http.StatusNotFound, codes.NotFound, "not found"},
	AlreadyExists:      {http.StatusConflict, codes.AlreadyExists, "already exists"},
	PermissionDenied:   {http.StatusForbidden, codes.PermissionDenied, "permission denied"},
	Unauthenticated:    {http.StatusUnauthorized, codes.Unauthenticated, "unauthenticated"},
	ResourceExhausted:  {http.StatusTooManyRequests, codes.ResourceExhausted, "resource exhausted"},
	FailedPrecondition: {http.StatusPreconditionFailed, codes.FailedPrecondition, "failed precondition"},
	Aborted:            {http.StatusConflict, codes.Aborted, "aborted"},
	OutOfRange:         {http.StatusBadRequest, codes.OutOfRange, "out of range"},
	Unimplemented:      {http.StatusNotImplemented, codes.Unimplemented, "unimplemented"},
	Internal:           {http.StatusInternalServerError, codes.Internal, "internal error"},
	Unavailable:        {http.StatusServiceUnavailable, codes.Unavailable, "service unavailable"},
	DataLoss:           {http.StatusInternalServerError, codes.DataLoss, "data loss"},
	Timeout:            {http.StatusGatewayTimeout, codes.DeadlineExceeded, "timeout"},

	TokenExpired:    {http.StatusUnauthorized, codes.Unauthenticated, "token expired"},
	TokenInvalid:    {http.StatusUnauthorized, codes.Unauthenticated, "token invalid"},
	TokenMissing:    {http.StatusUnauthorized, codes.Unauthenticated, "token missing"},
	AccessDenied:    {http.StatusForbidden, codes.PermissionDenied, "access denied"},
	SessionExpired:  {http.StatusUnauthorized, codes.Unauthenticated, "session expired"},

	ValidationFailed: {http.StatusBadRequest, codes.InvalidArgument, "validation failed"},
	MissingParameter: {http.StatusBadRequest, codes.InvalidArgument, "missing parameter"},
	InvalidFormat:    {http.StatusBadRequest, codes.InvalidArgument, "invalid format"},
	OutOfBounds:      {http.StatusBadRequest, codes.OutOfRange, "out of bounds"},

	UserNotFound:        {http.StatusNotFound, codes.NotFound, "user not found"},
	UserExists:          {http.StatusConflict, codes.AlreadyExists, "user already exists"},
	OrderNotFound:       {http.StatusNotFound, codes.NotFound, "order not found"},
	InsufficientBalance: {http.StatusBadRequest, codes.FailedPrecondition, "insufficient balance"},

	DatabaseError:   {http.StatusInternalServerError, codes.Internal, "database error"},
	CacheError:      {http.StatusInternalServerError, codes.Internal, "cache error"},
	MQError:         {http.StatusInternalServerError, codes.Internal, "message queue error"},
	ThirdPartyError: {http.StatusBadGateway, codes.Unavailable, "third party error"},
	NetworkError:    {http.StatusBadGateway, codes.Unavailable, "network error"},

	SystemError: {http.StatusInternalServerError, codes.Internal, "system error"},
	ConfigError: {http.StatusInternalServerError, codes.Internal, "config error"},
	InitError:   {http.StatusInternalServerError, codes.Internal, "initialization error"},
}

// HTTPStatus 返回 HTTP 状态码
func (c Code) HTTPStatus() int {
	if info, ok := codeRegistry[c]; ok {
		return info.httpStatus
	}
	return http.StatusInternalServerError
}

// GRPCCode 返回 gRPC 状态码
func (c Code) GRPCCode() codes.Code {
	if info, ok := codeRegistry[c]; ok {
		return info.grpcCode
	}
	return codes.Unknown
}

// Message 返回默认消息
func (c Code) Message() string {
	if info, ok := codeRegistry[c]; ok {
		return info.message
	}
	return "unknown error"
}

// String 返回错误码字符串
func (c Code) String() string {
	return c.Message()
}

// RegisterCode 注册自定义错误码
func RegisterCode(code Code, httpStatus int, grpcCode codes.Code, message string) {
	codeRegistry[code] = codeInfo{httpStatus, grpcCode, message}
}
