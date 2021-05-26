package reflect

import (
	"context"
	"errors"
	"math/big"
	"reflect"

	tfsdk "github.com/hashicorp/terraform-plugin-framework"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Into uses the data in `val` to populate `target`, using the reflection
// package to recursively reflect into structs and slices. If `target` is an
// AttributeValue, its assignment method will be used instead of reflecting. If
// `target` is a tftypes.ValueConverter, the FromTerraformValue method will be
// used instead of using reflection. Primitives are set using the val.As
// method. Structs use reflection: each exported struct field must have a
// "tfsdk" tag with the name of the field in the tftypes.Value, and all fields
// in the tftypes.Value must have a corresponding property in the struct. Into
// will be called for each struct field. Slices will have Into called for each
// element.
func Into(ctx context.Context, val tftypes.Value, target interface{}, path *tftypes.AttributePath) error {
	if _, ok := target.(tfsdk.AttributeValue); ok {
		// TODO: use builtin assignment through the interface
		return path.NewError(errors.New(("not implemented yet")))
	}
	if vc, ok := target.(tftypes.ValueConverter); ok {
		err := vc.FromTerraform5Value(val)
		if err != nil {
			return path.NewError(err)
		}
		return nil
	}
	kind := trueReflectValue(target).Type().Kind()
	if _, ok := trueReflectValue(target).Interface().(big.Float); ok {
		// cheat, pretend *big.Float is a float64 so it gets reflected
		// as a number
		kind = reflect.Float64
	} else if _, ok := trueReflectValue(target).Interface().(big.Int); ok {
		// cheat, pretend *big.Int is an int64 so it gets reflected as
		// a number
		kind = reflect.Int64
	}
	switch kind {
	case reflect.Struct:
		return reflectObjectIntoStruct(ctx, val, target, path)
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float64,
		reflect.String:
		return reflectPrimitive(ctx, val, target, path)
	case reflect.Slice:
		return reflectSlice(ctx, val, target, path)
	default:
		return path.NewErrorf("don't know how to reflect %s into %T", val.Type(), target)
	}
}
