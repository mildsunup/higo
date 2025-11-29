package utils

// --- 指针工具 ---

// Ptr 返回值的指针
func Ptr[T any](v T) *T {
	return &v
}

// Val 返回指针的值，nil 返回零值
func Val[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// ValOr 返回指针的值，nil 返回默认值
func ValOr[T any](p *T, defaultVal T) T {
	if p == nil {
		return defaultVal
	}
	return *p
}

// IsNil 检查指针是否为 nil
func IsNil[T any](p *T) bool {
	return p == nil
}

// IsNotNil 检查指针是否非 nil
func IsNotNil[T any](p *T) bool {
	return p != nil
}

// --- Optional 类型 ---

// Optional 可选值
type Optional[T any] struct {
	value   T
	present bool
}

// Some 创建有值的 Optional
func Some[T any](v T) Optional[T] {
	return Optional[T]{value: v, present: true}
}

// None 创建空的 Optional
func None[T any]() Optional[T] {
	return Optional[T]{}
}

// OfNullable 从指针创建 Optional
func OfNullable[T any](p *T) Optional[T] {
	if p == nil {
		return None[T]()
	}
	return Some(*p)
}

// IsPresent 是否有值
func (o Optional[T]) IsPresent() bool {
	return o.present
}

// IsEmpty 是否为空
func (o Optional[T]) IsEmpty() bool {
	return !o.present
}

// Get 获取值（无值时 panic）
func (o Optional[T]) Get() T {
	if !o.present {
		panic("optional: no value present")
	}
	return o.value
}

// OrElse 获取值或默认值
func (o Optional[T]) OrElse(defaultVal T) T {
	if o.present {
		return o.value
	}
	return defaultVal
}

// OrElseGet 获取值或通过函数获取默认值
func (o Optional[T]) OrElseGet(supplier func() T) T {
	if o.present {
		return o.value
	}
	return supplier()
}

// IfPresent 如果有值则执行
func (o Optional[T]) IfPresent(consumer func(T)) {
	if o.present {
		consumer(o.value)
	}
}

// MapOpt 转换值
func MapOpt[T, R any](o Optional[T], mapper func(T) R) Optional[R] {
	if !o.present {
		return None[R]()
	}
	return Some(mapper(o.value))
}

// FlatMapOpt 扁平化转换
func FlatMapOpt[T, R any](o Optional[T], mapper func(T) Optional[R]) Optional[R] {
	if !o.present {
		return None[R]()
	}
	return mapper(o.value)
}
