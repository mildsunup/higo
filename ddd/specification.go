package ddd

// Specification 规约接口
// 用于封装业务规则，支持组合
type Specification[T any] interface {
	IsSatisfiedBy(entity T) bool
}

// SpecFunc 函数式规约
type SpecFunc[T any] func(T) bool

func (f SpecFunc[T]) IsSatisfiedBy(entity T) bool { return f(entity) }

// Spec 创建函数式规约
func Spec[T any](fn func(T) bool) Specification[T] {
	return SpecFunc[T](fn)
}

// And 与组合
func And[T any](left, right Specification[T]) Specification[T] {
	return SpecFunc[T](func(entity T) bool {
		return left.IsSatisfiedBy(entity) && right.IsSatisfiedBy(entity)
	})
}

// Or 或组合
func Or[T any](left, right Specification[T]) Specification[T] {
	return SpecFunc[T](func(entity T) bool {
		return left.IsSatisfiedBy(entity) || right.IsSatisfiedBy(entity)
	})
}

// Not 取反
func Not[T any](spec Specification[T]) Specification[T] {
	return SpecFunc[T](func(entity T) bool {
		return !spec.IsSatisfiedBy(entity)
	})
}

// All 全部满足
func All[T any](specs ...Specification[T]) Specification[T] {
	return SpecFunc[T](func(entity T) bool {
		for _, spec := range specs {
			if !spec.IsSatisfiedBy(entity) {
				return false
			}
		}
		return true
	})
}

// Any 任一满足
func Any[T any](specs ...Specification[T]) Specification[T] {
	return SpecFunc[T](func(entity T) bool {
		for _, spec := range specs {
			if spec.IsSatisfiedBy(entity) {
				return true
			}
		}
		return false
	})
}
