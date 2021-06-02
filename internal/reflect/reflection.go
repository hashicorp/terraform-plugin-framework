package reflect

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	result, err := buildReflectValue(ctx, val, v.Elem(), opts, tftypes.NewAttributePath())
	if err != nil {
		return err
	}
	v.Elem().Set(result)
	return nil
}

// buildReflectValue constructs a reflect.Value of the same type as `target`,
// populated with the data in `val`. It will defensively instantiate new values
// to set, making it safe for use with pointer types which may be nil. It tries
// to give consumers the ability to override its default behaviors wherever
// possible.
func buildReflectValue(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	// if this isn't a valid reflect.Value, bail before we accidentally
	// panic
	if !target.IsValid() {
		return target, path.NewErrorf("invalid target")
	}
	// if this is an attr.Value, build the type from that
	if target.Type().Implements(reflect.TypeOf((*attr.Value)(nil)).Elem()) {
		return reflectAttributeValue(ctx, val, target, opts, path)
	}
	// if this tells tftypes how to build an instance of it out of a
	// tftypes.Value, well, that's what we want, so do that instead of our
	// default logic.
	if target.Type().Implements(reflect.TypeOf((*tftypes.ValueConverter)(nil)).Elem()) {
		return reflectValueConverter(ctx, val, target, opts, path)
	}
	// if this can explicitly be set to unknown, do that
	if target.Type().Implements(reflect.TypeOf((*setUnknownable)(nil)).Elem()) {
		res, err := reflectUnknownable(ctx, val, target, opts, path)
		if err != nil {
			return target, err
		}
		target = res
		// only return if it's unknown; we want to call SetUnknown
		// either way, but if the value is unknown, there's nothing
		// else to do, so bail
		if !val.IsKnown() {
			return target, nil
		}
	}
	// if this can explicitly be set to null, do that
	if target.Type().Implements(reflect.TypeOf((*setNullable)(nil)).Elem()) {
		res, err := reflectNullable(ctx, val, target, opts, path)
		if err != nil {
			return target, err
		}
		target = res
		// only return if it's null; we want to call SetNull either
		// way, but if the value is null, there's nothing else to do,
		// so bail
		if val.IsNull() {
			return target, nil
		}
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
		// we already handled null the only ways we can
		// we checked that target doesn't have a SetNull method we can
		// call
		// we checked that target isn't an AttributeValue
		// all that's left to us now is to set it as an empty value or
		// throw an error, depending on what's in opts
		if canBeNil(target) || opts.UnhandledNullAsEmpty {
			return reflect.Zero(target.Type()), nil
		}
		return target, path.NewError(errors.New("unhandled null value"))
	}
	// *big.Float and *big.Int are technically pointers, but we want them
	// handled as numbers
	if target.Type() == reflect.TypeOf(big.NewFloat(0)) || target.Type() == reflect.TypeOf(big.NewInt(0)) {
		return reflectNumber(ctx, val, target, opts, path)
	}
	switch target.Kind() {
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
		return reflectPointer(ctx, val, target, opts, path)
	default:
		return target, path.NewErrorf("don't know how to reflect %s into %s", val.Type(), target.Type())
	}
}
