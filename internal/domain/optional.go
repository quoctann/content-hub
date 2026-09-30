package domain

// Optional distinguishes "leave unchanged" (Set == false) from "set to Value",
// where Value may itself be nil to clear a nullable column.
type Optional[T any] struct {
	Set   bool
	Value T
}

// Some returns an Optional that is set to v.
func Some[T any](v T) Optional[T] {
	return Optional[T]{Set: true, Value: v}
}
