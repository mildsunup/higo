package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCStatus 转换为 gRPC Status
func ToGRPCStatus(err error) *status.Status {
	if err == nil {
		return status.New(codes.OK, "")
	}

	var e *Error
	if As(err, &e) {
		return status.New(e.code.GRPCCode(), e.message)
	}

	return status.New(codes.Unknown, err.Error())
}

// FromGRPCStatus 从 gRPC Status 创建错误
func FromGRPCStatus(s *status.Status) *Error {
	if s == nil || s.Code() == codes.OK {
		return nil
	}

	code := grpcCodeToCode(s.Code())
	return New(code, s.Message())
}

// FromGRPCError 从 gRPC 错误创建错误
func FromGRPCError(err error) *Error {
	if err == nil {
		return nil
	}

	s, ok := status.FromError(err)
	if !ok {
		return Wrap(err, Unknown, err.Error())
	}

	return FromGRPCStatus(s)
}

// grpcCodeToCode 将 gRPC 状态码映射到内部错误码
func grpcCodeToCode(c codes.Code) Code {
	switch c {
	case codes.OK:
		return OK
	case codes.Canceled:
		return Canceled
	case codes.Unknown:
		return Unknown
	case codes.InvalidArgument:
		return InvalidArgument
	case codes.DeadlineExceeded:
		return Timeout
	case codes.NotFound:
		return NotFound
	case codes.AlreadyExists:
		return AlreadyExists
	case codes.PermissionDenied:
		return PermissionDenied
	case codes.ResourceExhausted:
		return ResourceExhausted
	case codes.FailedPrecondition:
		return FailedPrecondition
	case codes.Aborted:
		return Aborted
	case codes.OutOfRange:
		return OutOfRange
	case codes.Unimplemented:
		return Unimplemented
	case codes.Internal:
		return Internal
	case codes.Unavailable:
		return Unavailable
	case codes.DataLoss:
		return DataLoss
	case codes.Unauthenticated:
		return Unauthenticated
	default:
		return Unknown
	}
}

// GRPCCode 返回错误的 gRPC 状态码
func GRPCCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	var e *Error
	if As(err, &e) {
		return e.code.GRPCCode()
	}

	return codes.Unknown
}
