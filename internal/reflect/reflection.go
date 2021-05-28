package reflect

import (
	"context"
	"errors"
	"fmt"
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
func Into(ctx context.Context, val tftypes.Value, target interface{}, opts Options) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer, got %T, which is a %s", target, v.Kind())
	}
	result, err := buildReflectValue(ctx, val, v, opts, tftypes.NewAttributePath())
	if err != nil {
		return err
	}
	v.Set(result)
	return nil
}

func buildReflectValue(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	if _, ok := target.Interface().(tfsdk.AttributeValue); ok {
		// TODO: use builtin assignment through the interface
		return target, path.NewError(errors.New(("not implemented yet")))
	}
	if v, ok := target.Interface().(setUnknownable); ok {
		err := v.SetUnknown(!val.IsKnown())
		if err != nil {
			return target, path.NewError(err)
		}
		if !val.IsKnown() {
			return target, nil
		}
	}
	if v, ok := target.Interface().(setNullable); ok {
		err := v.SetNull(val.IsNull())
		if err != nil {
			return target, path.NewError(err)
		}
		if val.IsNull() {
			return target, nil
		}
	}
	if vc, ok := target.Interface().(tftypes.ValueConverter); ok {
		err := vc.FromTerraform5Value(val)
		if err != nil {
			return target, path.NewError(err)
		}
		return target, nil
	}
	if !val.IsKnown() {
		// we already handled unknown the only ways we can
		// we checked that target doesn't have a SetUnknown method we
		// can call
		// we checked that target isn't an AttributeValue
		// all that's left to us now is to set it as an empty value or
		// throw an error, depending on what's in opts
		if !opts.UnhandledUnknownAsEmpty {
			return target, path.NewError(errors.New("unhandled unknown value"))
		}
		// we want to set unhandled unknowns to the empty value
		return reflect.Zero(target.Type()), nil
	}

	if val.IsNull() {
		if canBeNil(target) || opts.UnhandledNullAsEmpty {
			return reflect.Zero(target.Type()), nil
		}
		return target, path.NewError(errors.New("unhandled null value"))
	}
	kind := target.Kind()
	if _, ok := target.Interface().(*big.Float); ok {
		// cheat, pretend *big.Float is a float64 so it gets reflected
		// as a number
		kind = reflect.Float64
	} else if _, ok := target.Interface().(*big.Int); ok {
		// cheat, pretend *big.Int is an int64 so it gets reflected as
		// a number
		kind = reflect.Int64
	}
	switch kind {
	case reflect.Struct:
		return reflectStructFromObject(ctx, val, target, opts, path)
	case reflect.Bool, reflect.String:
		return reflectPrimitive(ctx, val, target, path)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		// numbers are the wooooorst and need their own special handling
		// because we can't just hand them off to tftypes and also
		// because we can't just make people use *big.Floats, because a
		// nil *big.Float will crash everything if we don't handle it
		// as a special case, so let's just special case numbers and
		// let people use the types they want
		return reflectNumber(ctx, val, target, opts, path)
	case reflect.Slice:
		return reflectSlice(ctx, val, target, opts, path)
	case reflect.Map:
		return reflectMap(ctx, val, target, opts, path)
	case reflect.Ptr:
		// TODO: handle pointers
		return target, path.NewErrorf("not implemented yet")
	case reflect.Interface:
		// TODO: handle interfaces
		return target, path.NewErrorf("not implemented yet")
	default:
		return target, path.NewErrorf("don't know how to reflect %s into %s", val.Type(), target.Type())
	}
}
