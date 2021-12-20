package reflect

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func Tuple(ctx context.Context, val attr.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, diag.Diagnostics) {
	return List(ctx, val, target, opts, path)
}

// FromSlice returns an attr.Value as produced by `typ` using the data in
// `val`. `val` must be a slice. `typ` must be an attr.TypeWithElementType or
// attr.TypeWithElementTypes. If the slice is nil, the representation of null
// for `typ` will be returned. Otherwise, FromSlice will recurse into FromValue
// for each element in the slice, using the element type or types defined on
// `typ` to construct values for them.
//
// It is meant to be called through FromValue, not directly.
func FromSlice(ctx context.Context, typ attr.Type, val reflect.Value, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	// TODO: support tuples, which are attr.TypeWithElementTypes
	tfType := typ.TerraformType(ctx)

	if val.IsNil() {
		tfVal := tftypes.NewValue(tfType, nil)

		if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
			diags.Append(typeWithValidate.Validate(ctx, tfVal, path)...)

			if diags.HasError() {
				return nil, diags
			}
		}

		attrVal, err := typ.ValueFromTerraform(ctx, tfVal)

		if err != nil {
			diags.AddAttributeError(
				path,
				"Value Conversion Error",
				"An unexpected error was encountered trying to convert from slice value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return nil, diags
		}

		return attrVal, diags
	}

	t, ok := typ.(attr.TypeWithElementType)
	if !ok {
		err := fmt.Errorf("cannot use type %T as schema type %T; %T must be an attr.TypeWithElementType to hold %T", val, typ, typ, val)
		diags.AddAttributeError(
			path,
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert from slice value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	elemType := t.ElementType()
	tfElems := make([]tftypes.Value, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		// The underlying reflect.Slice is fetched by Index(). For set types,
		// the path is value-based instead of index-based. Since there is only
		// the index until the value is retrieved, this will pass the
		// technically incorrect index-based path at first for framework
		// debugging purposes, then correct the path afterwards.
		valPath := path.WithElementKeyInt(i)

		val, valDiags := FromValue(ctx, elemType, val.Index(i).Interface(), valPath)
		diags.Append(valDiags...)

		if diags.HasError() {
			return nil, diags
		}

		tfVal, err := val.ToTerraformValue(ctx)
		if err != nil {
			return nil, append(diags, toTerraformValueErrorDiag(err, path))
		}

		if tfType.Is(tftypes.Set{}) {
			valPath = path.WithElementKeyValue(tfVal)
		}

		if typeWithValidate, ok := elemType.(attr.TypeWithValidate); ok {
			diags.Append(typeWithValidate.Validate(ctx, tfVal, valPath)...)
			if diags.HasError() {
				return nil, diags
			}
		}

		tfElems = append(tfElems, tfVal)
	}

	err := tftypes.ValidateValue(tfType, tfElems)
	if err != nil {
		return nil, append(diags, validateValueErrorDiag(err, path))
	}

	tfVal := tftypes.NewValue(tfType, tfElems)

	if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
		diags.Append(typeWithValidate.Validate(ctx, tfVal, path)...)

		if diags.HasError() {
			return nil, diags
		}
	}

	attrVal, err := typ.ValueFromTerraform(ctx, tfVal)

	if err != nil {
		diags.AddAttributeError(
			path,
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert from slice value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	return attrVal, diags
}
