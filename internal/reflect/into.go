package reflect

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
func Into(ctx context.Context, val attr.Value, target interface{}, opts Options) diag.Diagnostics {
	var diags diag.Diagnostics

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr {
		err := fmt.Errorf("target must be a pointer, got %T, which is a %s", target, v.Kind())
		diags.AddError(
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert the value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}
	result, diags := BuildValue(ctx, val, v.Elem(), opts, tftypes.NewAttributePath())
	if diags.HasError() {
		return diags
	}
	v.Elem().Set(result)
	return diags
}

// BuildValue constructs a reflect.Value of the same type as `target`,
// populated with the data in `val`. It will defensively instantiate new values
// to set, making it safe for use with pointer types which may be nil. It tries
// to give consumers the ability to override its default behaviors wherever
// possible.
func BuildValue(ctx context.Context, val attr.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	// if this isn't a valid reflect.Value, bail before we accidentally
	// panic
	if !target.IsValid() {
		err := fmt.Errorf("invalid target")
		diags.AddAttributeError(
			path,
			"Value Conversion Error",
			"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return target, diags
	}
	// if this is an attr.Value, build the type from that
	if target.Type().Implements(reflect.TypeOf((*attr.Value)(nil)).Elem()) {
		return NewAttributeValue(ctx, val, target, opts, path)
	}
	// if this tells tftypes how to build an instance of it out of a
	// tftypes.Value, well, that's what we want, so do that instead of our
	// default logic.
	if target.Type().Implements(reflect.TypeOf((*tftypes.ValueConverter)(nil)).Elem()) {
		return NewValueConverter(ctx, val, target, opts, path)
	}

	// grab the tftypes.Value behind the attr.Value for easy null/unknown
	// checking
	tfVal, err := val.ToTerraformValue(ctx)
	if err != nil {
		// TODO: handle error
	}

	// if this can explicitly be set to unknown, do that
	if target.Type().Implements(reflect.TypeOf((*Unknownable)(nil)).Elem()) {
		res, unknownableDiags := NewUnknownable(ctx, val, target, opts, path)
		diags.Append(unknownableDiags...)
		if diags.HasError() {
			return target, diags
		}
		target = res
		// only return if it's unknown; we want to call SetUnknown
		// either way, but if the value is unknown, there's nothing
		// else to do, so bail
		if !tfVal.IsKnown() {
			return target, nil
		}
	}
	// if this can explicitly be set to null, do that
	if target.Type().Implements(reflect.TypeOf((*Nullable)(nil)).Elem()) {
		res, nullableDiags := NewNullable(ctx, val, target, opts, path)
		diags.Append(nullableDiags...)
		if diags.HasError() {
			return target, diags
		}
		target = res
		// only return if it's null; we want to call SetNull either
		// way, but if the value is null, there's nothing else to do,
		// so bail
		if tfVal.IsNull() {
			return target, nil
		}
	}
	if !tfVal.IsKnown() {
		// we already handled unknown the only ways we can
		// we checked that target doesn't have a SetUnknown method we
		// can call
		// we checked that target isn't an AttributeValue
		// all that's left to us now is to set it as an empty value or
		// throw an error, depending on what's in opts
		if !opts.UnhandledUnknownAsEmpty {
			err := fmt.Errorf("unhandled unknown value")
			diags.AddAttributeError(
				path,
				"Value Conversion Error",
				"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return target, diags
		}
		// we want to set unhandled unknowns to the empty value
		return reflect.Zero(target.Type()), diags
	}

	if tfVal.IsNull() {
		// we already handled null the only ways we can
		// we checked that target doesn't have a SetNull method we can
		// call
		// we checked that target isn't an AttributeValue
		// all that's left to us now is to set it as an empty value or
		// throw an error, depending on what's in opts
		if canBeNil(target) || opts.UnhandledNullAsEmpty {
			return reflect.Zero(target.Type()), nil
		}

		err := fmt.Errorf("unhandled null value")
		diags.AddAttributeError(
			path,
			"Value Conversion Error",
			"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return target, diags
	}

	if target.Kind() == reflect.Ptr {
		ptr, ptrDiags := Pointer(ctx, val, target, opts, path)
		diags.Append(ptrDiags...)
		return ptr, diags
	}

	switch {
	case tfVal.Type().Is(tftypes.String), tfVal.Type().Is(tftypes.Bool):
		prim, valDiags := Primitive(ctx, val, target, path)
		diags.Append(valDiags...)
		return prim, diags
	case tfVal.Type().Is(tftypes.Number):
		num, numDiags := Number(ctx, val, target, opts, path)
		diags.Append(numDiags...)
		return num, diags
	case tfVal.Type().Is(tftypes.Object{}):
		obj, objDiags := Object(ctx, val, target, opts, path)
		diags.Append(objDiags...)
		return obj, diags
	case tfVal.Type().Is(tftypes.List{}):
		list, listDiags := List(ctx, val, target, opts, path)
		diags.Append(listDiags...)
		return list, diags
	case tfVal.Type().Is(tftypes.Set{}):
		set, setDiags := Set(ctx, val, target, opts, path)
		diags.Append(setDiags...)
		return set, diags
	case tfVal.Type().Is(tftypes.Tuple{}):
		tup, tupDiags := Tuple(ctx, val, target, opts, path)
		diags.Append(tupDiags...)
		return tup, diags
	case tfVal.Type().Is(tftypes.Map{}):
		m, mDiags := Map(ctx, val, target, opts, path)
		diags.Append(mDiags...)
		return m, diags
	default:
		err := fmt.Errorf("don't know how to reflect %s into %s", val.Type(ctx), target.Type())
		diags.AddAttributeError(
			path,
			"Value Conversion Error",
			"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return target, diags
	}
}
