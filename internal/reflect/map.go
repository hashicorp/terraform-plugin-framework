package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Map creates a map value that matches the type of `target`, and populates it
// with the contents of `val`.
func Map(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	underlyingValue := trueReflectValue(target)

	// this only works with maps, so check that out first
	if underlyingValue.Kind() != reflect.Map {
		return target, path.NewErrorf("expected a map type, got %s", target.Type())
	}
	if !val.Type().Is(tftypes.Map{}) {
		return target, path.NewErrorf("can't reflect %s into a map, must be a map", val.Type().String())
	}
	elemTyper, ok := typ.(attr.TypeWithElementType)
	if !ok {
		return target, path.NewErrorf("can't reflect map using type information provided by %T, %T must be an attr.TypeWithElementType", typ, typ)
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
	elemAttrType := elemTyper.ElementType()

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
		result, err := BuildValue(ctx, elemAttrType, value, targetValue, opts, path)
		if err != nil {
			return target, err
		}
		m.SetMapIndex(reflect.ValueOf(key), result)
	}
	return m, nil
}

// FromMap returns an attr.Value representing the data contained in `val`.
// `val` must be a map type with keys that are a string type. The attr.Value
// will be of the type produced by `typ`.
//
// It is meant to be called through OutOf, not directly.
func FromMap(ctx context.Context, typ attr.TypeWithElementType, val reflect.Value, path *tftypes.AttributePath) (attr.Value, error) {
	tfType := typ.TerraformType(ctx)

	if val.IsNil() {
		tfVal := tftypes.NewValue(tfType, nil)

		if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
			// TODO: Diagnostics to error handling, e.g. go-multierror? Warning handling?
			_ = typeWithValidate.Validate(ctx, tfVal)
		}

		return typ.ValueFromTerraform(ctx, tfVal)
	}

	elemType := typ.ElementType()
	tfElems := map[string]tftypes.Value{}
	for _, key := range val.MapKeys() {
		if key.Kind() != reflect.String {
			return nil, path.NewErrorf("map keys must be strings, got %s", key.Type())
		}
		val, err := FromValue(ctx, elemType, val.MapIndex(key).Interface(), path.WithElementKeyString(key.String()))
		if err != nil {
			return nil, err
		}
		tfVal, err := val.ToTerraformValue(ctx)
		if err != nil {
			return nil, path.NewError(err)
		}

		tfElemType := elemType.TerraformType(ctx)
		err = tftypes.ValidateValue(tfElemType, tfVal)

		if err != nil {
			return nil, path.NewError(err)
		}

		tfElemVal := tftypes.NewValue(tfElemType, tfVal)

		if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
			// TODO: Diagnostics to error handling, e.g. go-multierror? Warning handling?
			_ = typeWithValidate.Validate(ctx, tfElemVal)
		}

		tfElems[key.String()] = tfElemVal
	}

	err := tftypes.ValidateValue(tfType, tfElems)
	if err != nil {
		return nil, path.NewError(err)
	}

	tfVal := tftypes.NewValue(tfType, tfElems)

	if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
		// TODO: Diagnostics to error handling, e.g. go-multierror? Warning handling?
		_ = typeWithValidate.Validate(ctx, tfVal)
	}

	return typ.ValueFromTerraform(ctx, tfVal)
}
