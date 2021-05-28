package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectSlice(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	// this only works with slices, so check that out first
	if target.Kind() != reflect.Slice {
		return target, path.NewErrorf("expected a slice type, got %s", target.Type())
	}

	// we need our value to become a list of values so we can iterate over
	// them and handle them individually
	var values []tftypes.Value
	err := val.As(&values)
	if err != nil {
		return target, path.NewError(err)
	}

	// we need to know the type the slice is wrapping
	elemType := target.Type().Elem()

	// we want an empty version of the slice
	slice := reflect.MakeSlice(target.Type(), 0, len(values))

	// go over each of the values passed in, create a Go value of the right
	// type for them, and add it to our new slice
	for pos, value := range values {
		// create a new Go value of the type that can go in the slice
		targetValue := reflect.Zero(elemType)

		// update our path so we can have nice errors
		path := path.WithElementKeyInt(int64(pos))

		// reflect the value into our new target
		val, err := buildReflectValue(ctx, value, targetValue, opts, path)
		if err != nil {
			return target, err
		}

		// add the new target to our slice
		slice = reflect.Append(slice, val)
	}

	return slice, nil
}
