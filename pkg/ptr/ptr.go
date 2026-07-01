// Package ptr 提供值和指针之间的通用转换工具。
package ptr

// Of 返回 v 的指针。
//
// 返回的是传入值的副本指针，后续修改原变量不会影响该指针指向的值。
func Of[T any](v T) *T {
	return &v
}

// Clone 返回 p 指向值的副本指针。
//
// 当 p 为 nil 时，返回 nil。返回的指针不会和 p 共享同一个地址。
func Clone[T any](p *T) *T {
	if p == nil {
		return nil
	}
	return Of(*p)
}

// Value 返回 p 指向的值。
//
// 当 p 为 nil 时，返回 T 类型的零值。
func Value[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// ValueOr 返回 p 指向的值。
//
// 当 p 为 nil 时，返回 fallback。
func ValueOr[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}

// From 返回 p 指向的值和 p 是否非空。
//
// 当 p 为 nil 时，第一个返回值是 T 类型的零值，第二个返回值为 false。
func From[T any](p *T) (T, bool) {
	if p == nil {
		var zero T
		return zero, false
	}
	return *p, true
}

// IsNil 返回 p 是否为 nil。
func IsNil[T any](p *T) bool {
	return p == nil
}

// Equal 判断两个指针是否表示相同的可选值。
//
// 两个指针都为 nil 时返回 true；只有一个为 nil 时返回 false；
// 两个指针都非 nil 时比较它们指向的值。
func Equal[T comparable](a *T, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// Slice 将值切片转换为指针切片。
//
// 每个返回指针都指向对应元素值的副本，不会引用原切片的底层数组。
func Slice[T any](values []T) []*T {
	if values == nil {
		return nil
	}

	pointers := make([]*T, 0, len(values))
	for _, value := range values {
		pointers = append(pointers, Of(value))
	}
	return pointers
}

// Values 将指针切片转换为值切片。
//
// nil 指针会转换为 T 类型的零值。
func Values[T any](pointers []*T) []T {
	if pointers == nil {
		return nil
	}

	values := make([]T, 0, len(pointers))
	for _, pointer := range pointers {
		values = append(values, Value(pointer))
	}
	return values
}
