package fwschemadata_test

func pointer[T any](value T) *T {
	return &value
}
