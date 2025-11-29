package security

import (
	"errors"
	"unicode"
)

var (
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong    = errors.New("password must be at most 128 characters")
	ErrPasswordNoUpper    = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower    = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit    = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial  = errors.New("password must contain at least one special character")
)

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// DefaultPasswordPolicy 默认密码策略
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      8,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: false, // 可选
	}
}

// StrictPasswordPolicy 严格密码策略
func StrictPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      12,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
	}
}

// Validate 验证密码强度
func (p PasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return ErrPasswordTooShort
	}
	if len(password) > p.MaxLength {
		return ErrPasswordTooLong
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	if p.RequireUpper && !hasUpper {
		return ErrPasswordNoUpper
	}
	if p.RequireLower && !hasLower {
		return ErrPasswordNoLower
	}
	if p.RequireDigit && !hasDigit {
		return ErrPasswordNoDigit
	}
	if p.RequireSpecial && !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// ValidatePassword 使用默认策略验证密码
func ValidatePassword(password string) error {
	return DefaultPasswordPolicy().Validate(password)
}
