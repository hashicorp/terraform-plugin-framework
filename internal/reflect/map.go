package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectMap(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	underlyingValue := trueReflectValue(target)

	// this only works with maps, so check that out first
	if underlyingValue.Kind() != reflect.Map {
		return target, path.NewErrorf("expected a map type, got %s", target.Type())
	}
	if !val.Type().Is(tftypes.Map{}) {
		return target, path.NewErrorf("can't reflect %s into a map, must be a map", val.Type().String())
	}

	// we need our value to become a map of values so we can iterate over
	// them and handle them individually
	values := map[string]tftypes.Value{}
	err := val.As(&values)
	if err != nil {
		return target, path.NewError(err)
	}

	// we need to know the type the slice is wrapping
	elemType := underlyingValue.Type().Elem()

	// we want an empty version of the map
	m := reflect.MakeMapWithSize(underlyingValue.Type(), len(values))

	// go over each of the values passed in, create a Go value of the right
	// type for them, and add it to our new map
	for key, value := range values {
		// create a new Go value of the type that can go in the map
		targetValue := reflect.Zero(elemType)

		// update our path so we can have nice errors
		path := path.WithElementKeyString(key)

		// reflect the value into our new target
		result, err := buildReflectValue(ctx, value, targetValue, opts, path)
		if err != nil {
			return target, err
		}
		m.SetMapIndex(reflect.ValueOf(key), result)
	}
	return m, nil
}
