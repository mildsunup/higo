package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestNew(t *testing.T) {
	err := New(NotFound, "user not found")

	if err.Code() != NotFound {
		t.Errorf("expected code %d, got %d", NotFound, err.Code())
	}

	if err.Message() != "user not found" {
		t.Errorf("expected message 'user not found', got %s", err.Message())
	}

	if err.HTTPStatus() != http.StatusNotFound {
		t.Errorf("expected HTTP status %d, got %d", http.StatusNotFound, err.HTTPStatus())
	}
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("database connection failed")
	err := Wrap(cause, DatabaseError, "failed to get user")

	if err.Unwrap() != cause {
		t.Error("unwrap should return cause")
	}

	expected := "failed to get user: database connection failed"
	if err.Error() != expected {
		t.Errorf("expected error string %q, got %q", expected, err.Error())
	}
}

func TestWrapNil(t *testing.T) {
	err := Wrap(nil, Internal, "should be nil")
	if err != nil {
		t.Error("wrapping nil should return nil")
	}
}

func TestWithMeta(t *testing.T) {
	err := New(ValidationFailed, "validation error").
		WithMeta("field", "email").
		WithMeta("reason", "invalid format")

	if err.GetMeta("field") != "email" {
		t.Error("expected field = email")
	}

	if err.GetMeta("reason") != "invalid format" {
		t.Error("expected reason = invalid format")
	}
}

func TestWithMetadata(t *testing.T) {
	err := New(ValidationFailed, "validation error").
		WithMetadata(map[string]any{
			"field":  "email",
			"reason": "invalid",
		})

	meta := err.Metadata()
	if len(meta) != 2 {
		t.Errorf("expected 2 metadata entries, got %d", len(meta))
	}
}

func TestStackTrace(t *testing.T) {
	err := New(Internal, "test error")
	stack := err.StackTrace()

	if len(stack) == 0 {
		t.Error("expected stack trace")
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		err      error
		expected Code
	}{
		{nil, OK},
		{New(NotFound, "not found"), NotFound},
		{fmt.Errorf("plain error"), Unknown},
	}

	for _, tt := range tests {
		got := GetCode(tt.err)
		if got != tt.expected {
			t.Errorf("GetCode(%v) = %d, want %d", tt.err, got, tt.expected)
		}
	}
}

func TestIsCode(t *testing.T) {
	err := New(NotFound, "user not found")

	if !IsCode(err, NotFound) {
		t.Error("expected IsCode to return true for NotFound")
	}

	if IsCode(err, Internal) {
		t.Error("expected IsCode to return false for Internal")
	}
}

func TestCodeHTTPStatus(t *testing.T) {
	tests := []struct {
		code     Code
		expected int
	}{
		{OK, http.StatusOK},
		{NotFound, http.StatusNotFound},
		{InvalidArgument, http.StatusBadRequest},
		{Unauthenticated, http.StatusUnauthorized},
		{PermissionDenied, http.StatusForbidden},
		{Internal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		got := tt.code.HTTPStatus()
		if got != tt.expected {
			t.Errorf("Code(%d).HTTPStatus() = %d, want %d", tt.code, got, tt.expected)
		}
	}
}

func TestCodeGRPCCode(t *testing.T) {
	tests := []struct {
		code     Code
		expected codes.Code
	}{
		{OK, codes.OK},
		{NotFound, codes.NotFound},
		{InvalidArgument, codes.InvalidArgument},
		{Unauthenticated, codes.Unauthenticated},
		{Internal, codes.Internal},
	}

	for _, tt := range tests {
		got := tt.code.GRPCCode()
		if got != tt.expected {
			t.Errorf("Code(%d).GRPCCode() = %v, want %v", tt.code, got, tt.expected)
		}
	}
}

func TestToResponse(t *testing.T) {
	err := New(NotFound, "user not found").
		WithMeta("user_id", "123")

	resp := ToResponse(err)

	if resp.Code != int(NotFound) {
		t.Errorf("expected code %d, got %d", NotFound, resp.Code)
	}

	if resp.Message != "user not found" {
		t.Errorf("expected message 'user not found', got %s", resp.Message)
	}

	if resp.Details["user_id"] != "123" {
		t.Error("expected details to contain user_id")
	}
}

func TestMarshalJSON(t *testing.T) {
	err := New(NotFound, "not found")

	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("marshal error: %v", jsonErr)
	}

	var resp Response
	if jsonErr := json.Unmarshal(data, &resp); jsonErr != nil {
		t.Fatalf("unmarshal error: %v", jsonErr)
	}

	if resp.Code != int(NotFound) {
		t.Errorf("expected code %d, got %d", NotFound, resp.Code)
	}
}

func TestConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		code     Code
		httpCode int
	}{
		{"BadRequest", BadRequest("bad"), InvalidArgument, http.StatusBadRequest},
		{"Unauthorized", Unauthorized("unauth"), Unauthenticated, http.StatusUnauthorized},
		{"Forbidden", Forbidden("forbidden"), PermissionDenied, http.StatusForbidden},
		{"NotFoundErr", NotFoundErr("not found"), NotFound, http.StatusNotFound},
		{"Conflict", Conflict("exists"), AlreadyExists, http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code() != tt.code {
				t.Errorf("expected code %d, got %d", tt.code, tt.err.Code())
			}
			if tt.err.HTTPStatus() != tt.httpCode {
				t.Errorf("expected HTTP status %d, got %d", tt.httpCode, tt.err.HTTPStatus())
			}
		})
	}
}

func TestErrorChain(t *testing.T) {
	root := fmt.Errorf("root cause")
	wrapped := Wrap(root, DatabaseError, "db error")
	outer := Wrap(wrapped, Internal, "service error")

	// 检查错误链
	if !Is(outer, wrapped) {
		t.Error("outer should contain wrapped")
	}

	// 检查 Unwrap
	var dbErr *Error
	if !As(outer, &dbErr) {
		t.Error("should be able to extract Error from chain")
	}
}

func TestRegisterCode(t *testing.T) {
	const CustomCode Code = 9999

	RegisterCode(CustomCode, http.StatusTeapot, codes.Unknown, "custom error")

	if CustomCode.HTTPStatus() != http.StatusTeapot {
		t.Errorf("expected HTTP status %d, got %d", http.StatusTeapot, CustomCode.HTTPStatus())
	}

	if CustomCode.Message() != "custom error" {
		t.Errorf("expected message 'custom error', got %s", CustomCode.Message())
	}
}
