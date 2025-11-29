package ddd

// ValueObject 值对象接口
// 值对象通过属性值来标识，而非唯一标识符
type ValueObject interface {
	// Equals 判断两个值对象是否相等
	Equals(other ValueObject) bool
}

// Validatable 可验证接口
type Validatable interface {
	Validate() error
}

// ValidatableValueObject 可验证的值对象
type ValidatableValueObject interface {
	ValueObject
	Validatable
}
