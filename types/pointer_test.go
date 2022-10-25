package types

func pointer[T any](value T) *T {
	return &value
}
