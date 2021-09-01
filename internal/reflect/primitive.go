package reflect

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Primitive builds a string or boolean, depending on the type of `target`, and
// populates it with the data in `val`.
//
// It is meant to be called through `Into`, not directly.
func Primitive(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, path *tftypes.AttributePath) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch target.Kind() {
	case reflect.Bool:
		var b bool
		err := val.As(&b)
		if err != nil {
			diags.AddAttributeError(
				path,
				"Value Conversion Error",
				"An unexpected error was encountered trying to convert to boolean. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return target, diags
		}
		return reflect.ValueOf(b).Convert(target.Type()), nil
	case reflect.String:
		var s string
		err := val.As(&s)
		if err != nil {
			diags.AddAttributeError(
				path,
				"Value Conversion Error",
				"An unexpected error was encountered trying to convert to string. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return target, diags
		}
		return reflect.ValueOf(s).Convert(target.Type()), nil
	default:
		err := fmt.Errorf("unrecognized type: %s", target.Kind())
		diags.AddAttributeError(
			path,
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to primitive. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return target, diags
	}
}

// FromString returns an attr.Value as produced by `typ` from a string.
//
// It is meant to be called through OutOf, not directly.
func FromString(ctx context.Context, typ attr.Type, val string, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	err := tftypes.ValidateValue(tftypes.String, val)
	if err != nil {
		return nil, append(diags, validateValueErrorDiag(err, path))
	}
	tfStr := tftypes.NewValue(tftypes.String, val)

	if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
		diags.Append(typeWithValidate.Validate(ctx, tfStr)...)

		if diags.HasError() {
			return nil, diags
		}
	}

	str, err := typ.ValueFromTerraform(ctx, tfStr)
	if err != nil {
		return nil, append(diags, valueFromTerraformErrorDiag(err, path))
	}

	return str, diags
}

// FromBool returns an attr.Value as produced by `typ` from a bool.
//
// It is meant to be called through OutOf, not directly.
func FromBool(ctx context.Context, typ attr.Type, val bool, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	err := tftypes.ValidateValue(tftypes.Bool, val)
	if err != nil {
		return nil, append(diags, validateValueErrorDiag(err, path))
	}
	tfBool := tftypes.NewValue(tftypes.Bool, val)

	if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
		diags.Append(typeWithValidate.Validate(ctx, tfBool)...)

		if diags.HasError() {
			return nil, diags
		}
	}

	b, err := typ.ValueFromTerraform(ctx, tfBool)
	if err != nil {
		return nil, append(diags, valueFromTerraformErrorDiag(err, path))
	}

	return b, diags
}
