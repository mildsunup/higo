package utils

import (
	"golang.org/x/exp/constraints"
)

// --- 数值比较 ---

// Min 返回最小值
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max 返回最大值
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Clamp 限制值在范围内
func Clamp[T constraints.Ordered](value, min, max T) T {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Abs 绝对值
func Abs[T constraints.Signed | constraints.Float](n T) T {
	if n < 0 {
		return -n
	}
	return n
}

// --- 数值判断 ---

// InRange 检查值是否在范围内 [min, max]
func InRange[T constraints.Ordered](value, min, max T) bool {
	return value >= min && value <= max
}

// IsPositive 是否为正数
func IsPositive[T constraints.Signed | constraints.Float](n T) bool {
	return n > 0
}

// IsNegative 是否为负数
func IsNegative[T constraints.Signed | constraints.Float](n T) bool {
	return n < 0
}

// IsZero 是否为零
func IsZero[T constraints.Integer | constraints.Float](n T) bool {
	return n == 0
}

// IsEven 是否为偶数
func IsEven[T constraints.Integer](n T) bool {
	return n%2 == 0
}

// IsOdd 是否为奇数
func IsOdd[T constraints.Integer](n T) bool {
	return n%2 != 0
}

// --- 数值转换 ---

// ToInt 安全转换为 int
func ToInt[T constraints.Integer | constraints.Float](n T) int {
	return int(n)
}

// ToInt64 安全转换为 int64
func ToInt64[T constraints.Integer | constraints.Float](n T) int64 {
	return int64(n)
}

// ToFloat64 安全转换为 float64
func ToFloat64[T constraints.Integer | constraints.Float](n T) float64 {
	return float64(n)
}

// --- 聚合计算 ---

// Sum 求和
func Sum[T constraints.Integer | constraints.Float](nums ...T) T {
	var sum T
	for _, n := range nums {
		sum += n
	}
	return sum
}

// Avg 求平均值
func Avg[T constraints.Integer | constraints.Float](nums ...T) float64 {
	if len(nums) == 0 {
		return 0
	}
	var sum T
	for _, n := range nums {
		sum += n
	}
	return float64(sum) / float64(len(nums))
}

// MinSlice 切片最小值
func MinSlice[T constraints.Ordered](nums []T) (T, bool) {
	if len(nums) == 0 {
		var zero T
		return zero, false
	}
	min := nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
	}
	return min, true
}

// MaxSlice 切片最大值
func MaxSlice[T constraints.Ordered](nums []T) (T, bool) {
	if len(nums) == 0 {
		var zero T
		return zero, false
	}
	max := nums[0]
	for _, n := range nums[1:] {
		if n > max {
			max = n
		}
	}
	return max, true
}
