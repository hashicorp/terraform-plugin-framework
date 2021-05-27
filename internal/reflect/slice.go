package reflect

import (
	"context"
	"errors"
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectSlice(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) error {
	reflectValue := trueReflectValue(target)

	// this only works with slices, so check that out first
	if reflectValue.Kind() != reflect.Slice {
		return path.NewErrorf("expected a slice type, got %s", target.Type())
	}

	// if we can't set the target, there's no point to any of this
	if !reflectValue.CanSet() {
		return path.NewError(errors.New("value can't be set"))
	}

	// we need our value to become a list of values so we can iterate over
	// them and handle them individually
	var values []tftypes.Value
	err := val.As(&values)
	if err != nil {
		return path.NewError(err)
	}

	// we need to know the type the slice is wrapping
	elemType := reflectValue.Type().Elem()

	// we want an empty version of the slice
	sliced := reflectValue.Slice(0, 0)

	// go over each of the values passed in, create a Go value of the right
	// type for them, and add it to our new slice
	for pos, value := range values {
		// create a new Go value of the type that can go in the slice
		targetValue := reflect.Zero(elemType)

		// add the new target to our slice
		sliced = reflect.Append(sliced, targetValue)

		// update our path so we can have nice errors
		path := path.WithElementKeyInt(int64(pos))

		// reflect the value into our new target
		err := into(ctx, value, sliced.Index(sliced.Len()-1), opts, path)
		if err != nil {
			return err
		}
	}

	// update the target to be our slice
	reflectValue.Set(sliced)
	return nil
}
