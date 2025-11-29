package utils

import (
	"strings"
	"unicode"
	"unsafe"
)

// --- 字符串转换 ---

// ToSnakeCase 转换为蛇形命名 (camelCase -> camel_case)
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ToCamelCase 转换为驼峰命名 (snake_case -> snakeCase)
func ToCamelCase(s string) string {
	var result strings.Builder
	upper := false
	for _, r := range s {
		if r == '_' || r == '-' {
			upper = true
			continue
		}
		if upper {
			result.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ToPascalCase 转换为帕斯卡命名 (snake_case -> SnakeCase)
func ToPascalCase(s string) string {
	camel := ToCamelCase(s)
	if len(camel) == 0 {
		return camel
	}
	return strings.ToUpper(camel[:1]) + camel[1:]
}

// --- 字符串检查 ---

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return len(s) == 0
}

// IsBlank 检查字符串是否为空或只包含空白字符
func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty 检查字符串是否非空
func IsNotEmpty(s string) bool {
	return len(s) > 0
}

// IsNotBlank 检查字符串是否非空且不只包含空白字符
func IsNotBlank(s string) bool {
	return len(strings.TrimSpace(s)) > 0
}

// --- 字符串操作 ---

// DefaultIfEmpty 如果为空则返回默认值
func DefaultIfEmpty(s, defaultVal string) string {
	if IsEmpty(s) {
		return defaultVal
	}
	return s
}

// DefaultIfBlank 如果为空白则返回默认值
func DefaultIfBlank(s, defaultVal string) string {
	if IsBlank(s) {
		return defaultVal
	}
	return s
}

// Truncate 截断字符串
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// TruncateWithSuffix 截断字符串并添加后缀
func TruncateWithSuffix(s string, maxLen int, suffix string) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= len(suffix) {
		return suffix[:maxLen]
	}
	return s[:maxLen-len(suffix)] + suffix
}

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// --- 高性能转换 ---

// StringToBytes 零拷贝字符串转字节切片（只读）
func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString 零拷贝字节切片转字符串
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// --- 掩码 ---

// MaskPhone 手机号脱敏 (138****1234)
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskEmail 邮箱脱敏 (t***@example.com)
func MaskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 1 {
		return email
	}
	return email[:1] + "***" + email[at:]
}

// MaskIDCard 身份证脱敏 (110***********1234)
func MaskIDCard(idCard string) string {
	if len(idCard) < 8 {
		return idCard
	}
	return idCard[:3] + strings.Repeat("*", len(idCard)-7) + idCard[len(idCard)-4:]
}
