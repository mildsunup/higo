package utils

import (
	"time"
)

// 常用时间格式
const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
	ISO8601Format  = "2006-01-02T15:04:05Z07:00"
)

// --- 时间格式化 ---

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// FormatDateTime 格式化日期时间
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

// FormatISO8601 格式化为 ISO8601
func FormatISO8601(t time.Time) string {
	return t.Format(ISO8601Format)
}

// --- 时间解析 ---

// ParseDate 解析日期
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateFormat, s)
}

// ParseDateTime 解析日期时间
func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(DateTimeFormat, s)
}

// ParseISO8601 解析 ISO8601
func ParseISO8601(s string) (time.Time, error) {
	return time.Parse(ISO8601Format, s)
}

// --- 时间计算 ---

// StartOfDay 获取当天开始时间
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取当天结束时间
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek 获取本周开始时间（周一）
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return StartOfDay(t.AddDate(0, 0, -weekday+1))
}

// StartOfMonth 获取本月开始时间
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 获取本月结束时间
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// StartOfYear 获取本年开始时间
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// --- 时间判断 ---

// IsToday 是否是今天
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// IsYesterday 是否是昨天
func IsYesterday(t time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.YearDay() == yesterday.YearDay()
}

// IsSameDay 是否是同一天
func IsSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay()
}

// IsWeekend 是否是周末
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// --- 时间差 ---

// DaysBetween 计算两个时间相差的天数
func DaysBetween(t1, t2 time.Time) int {
	t1 = StartOfDay(t1)
	t2 = StartOfDay(t2)
	return int(t2.Sub(t1).Hours() / 24)
}

// Age 计算年龄
func Age(birthday time.Time) int {
	now := time.Now()
	age := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		age--
	}
	return age
}

// --- 时间戳 ---

// UnixMilli 获取毫秒时间戳
func UnixMilli(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// FromUnixMilli 从毫秒时间戳创建时间
func FromUnixMilli(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}
