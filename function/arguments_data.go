// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// ArgumentsData is the zero-based positional argument data sent by Terraform
// for a single function call. Use the Get method or GetArgument method in the
// Function type Run method to fetch the data.
//
// This data is automatically populated by the framework based on the function
// definition. For unit testing, use the NewArgumentsData function to manually
// create the data.
type ArgumentsData struct {
	values []attr.Value
}

// Equal returns true if all the underlying values are equivalent.
func (d ArgumentsData) Equal(o ArgumentsData) bool {
	if len(d.values) != len(o.values) {
		return false
	}

	for index, value := range d.values {
		if !value.Equal(o.values[index]) {
			return false
		}
	}

	return true
}

// Get retrieves all argument data and populates the targets with the values.
// All arguments must be present in the targets, including all parameters and an
// optional variadic parameter, otherwise an error diagnostic will be raised.
// Each target type must be acceptable for the data type in the parameter
// definition.
//
// Variadic parameter argument data must be consumed by a types.List or Go slice
// type with an element type appropriate for the parameter definition ([]T). The
// framework automatically populates this list with elements matching the zero,
// one, or more arguments passed.
func (d ArgumentsData) Get(ctx context.Context, targets ...any) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(d.values) == 0 {
		diags.AddError(
			"Invalid Argument Data Usage",
			"When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. "+
				"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
				"Function does not have argument data.",
		)

		return diags
	}

	if len(targets) != len(d.values) {
		diags.AddError(
			"Invalid Argument Data Usage",
			"When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. "+
				"The Get call requires all parameters and the final variadic parameter, if implemented, to be in the targets. "+
				"This is always an error in the provider code and should be reported to the provider developers.\n\n"+
				fmt.Sprintf("Given targets count: %d, expected targets count: %d", len(targets), len(d.values)),
		)

		return diags
	}

	for position, attrValue := range d.values {
		target := targets[position]

		if fwreflect.IsGenericAttrValue(ctx, target) {
			//nolint:forcetypeassert // Type assertion is guaranteed by the above `reflect.IsGenericAttrValue` function
			*(target.(*attr.Value)) = attrValue

			continue
		}

		tfValue, err := attrValue.ToTerraformValue(ctx)

		if err != nil {
			diags.AddError(
				"Argument Value Conversion Error",
				fmt.Sprintf("An unexpected error was encountered converting a %T to its equivalent Terraform representation. "+
					"This is always an error in the provider code and should be reported to the provider developers.\n\n"+
					"Position: %d\n"+
					"Error: %s",
					attrValue, position, err),
			)

			continue
		}

		reflectDiags := fwreflect.Into(ctx, attrValue.Type(ctx), tfValue, target, fwreflect.Options{}, path.Empty())

		diags.Append(reflectDiags...)
	}

	return diags
}

// GetArgument retrieves the argument data found at the given zero-based
// position and populates the target with the value. The target type must be
// acceptable for the data type in the parameter definition.
//
// Variadic parameter argument data must be consumed by a types.List or Go slice
// type with an element type appropriate for the parameter definition ([]T) at
// the position after all parameters. The framework automatically populates this
// list with elements matching the zero, one, or more arguments passed.
func (d ArgumentsData) GetArgument(ctx context.Context, position int, target any) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(d.values) == 0 {
		diags.AddError(
			"Invalid Argument Data Usage",
			"When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. "+
				"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
				"Function does not have argument data.",
		)

		return diags
	}

	if position >= len(d.values) {
		diags.AddError(
			"Invalid Argument Data Position",
			"When attempting to fetch argument data during the function call, the provider code attempted to read a non-existent argument position. "+
				"Function argument positions are 0-based and any final variadic parameter is represented as one argument position with an ordered list of the parameter data type. "+
				"This is always an error in the provider code and should be reported to the provider developers.\n\n"+
				fmt.Sprintf("Given argument position: %d, last argument position: %d", position, len(d.values)-1),
		)

		return diags
	}

	attrValue := d.values[position]

	if fwreflect.IsGenericAttrValue(ctx, target) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `reflect.IsGenericAttrValue` function
		*(target.(*attr.Value)) = attrValue

		return nil
	}

	tfValue, err := attrValue.ToTerraformValue(ctx)

	if err != nil {
		diags.AddError(
			"Argument Value Conversion Error",
			fmt.Sprintf("An unexpected error was encountered converting a %T to its equivalent Terraform representation. "+
				"This is always an error in the provider code and should be reported to the provider developers.\n\n"+
				"Error: %s", attrValue, err),
		)
		return diags
	}

	reflectDiags := fwreflect.Into(ctx, attrValue.Type(ctx), tfValue, target, fwreflect.Options{}, path.Empty())

	diags.Append(reflectDiags...)

	return diags
}

// NewArgumentsData creates an ArgumentsData. This is only necessary for unit
// testing as the framework automatically creates this data.
func NewArgumentsData(values []attr.Value) ArgumentsData {
	return ArgumentsData{
		values: values,
	}
}
