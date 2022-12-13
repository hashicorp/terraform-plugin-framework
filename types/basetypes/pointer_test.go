package basetypes

func pointer[T any](value T) *T {
	return &value
}
