package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// build a slice of elements, matching the type of `target`, and fill it with
// the data in `val`.
func reflectSlice(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	// this only works with slices, so check that out first
	if target.Kind() != reflect.Slice {
		return target, path.NewErrorf("expected a slice type, got %s", target.Type())
	}
	// TODO: check that the val is a list or set or tuple
	elemTyper, ok := typ.(attr.TypeWithElementType)
	if !ok {
		return target, path.NewErrorf("can't reflect %s using type information provided by %T, %T must be an attr.TypeWithElementType", val.Type(), typ, typ)
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
	elemAttrType := elemTyper.ElementType()

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
		val, err := BuildValue(ctx, elemAttrType, value, targetValue, opts, path)
		if err != nil {
			return target, err
		}

		// add the new target to our slice
		slice = reflect.Append(slice, val)
	}

	return slice, nil
}

func FromSlice(ctx context.Context, typ attr.Type, val reflect.Value, opts OutOfOptions, path *tftypes.AttributePath) (attr.Value, error) {
	// TODO: support tuples, which are attr.TypeWithElementTypes

	if val.Interface() == nil {
		return typ.ValueFromTerraform(ctx, tftypes.NewValue(typ.TerraformType(ctx), nil))
	}

	t := typ.(attr.TypeWithElementType)

	elemType := t.ElementType()
	tfElems := make([]tftypes.Value, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		val, err := FromValue(ctx, elemType, val.Index(i), opts, path.WithElementKeyInt(int64(i)))
		if err != nil {
			return nil, err
		}
		tfVal, err := val.ToTerraformValue(ctx)
		if err != nil {
			return nil, path.NewError(err)
		}
		err = tftypes.ValidateValue(elemType.TerraformType(ctx), tfVal)
		if err != nil {
			return nil, path.NewError(err)
		}
		tfElems = append(tfElems, tftypes.NewValue(elemType.TerraformType(ctx), tfVal))
	}
	err := tftypes.ValidateValue(typ.TerraformType(ctx), tfElems)
	if err != nil {
		return nil, path.NewError(err)
	}
	return typ.ValueFromTerraform(ctx, tftypes.NewValue(typ.TerraformType(ctx), tfElems))
}
