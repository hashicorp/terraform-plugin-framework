package proto6server

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// PlanGetAttributeValue is a duplicate of tfsdk.Plan.getAttributeValue,
// except it calls a local duplicate to Plan.terraformValueAtPath as well.
// It is duplicated to prevent any oddities with trying to use
// tfsdk.Plan.GetAttribute, which has some potentially undesirable logic.
// Refer to the tfsdk package for the large amount of testing done there.
//
// TODO: Clean up this abstraction back into an internal Plan type method.
// The extra Plan parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func PlanGetAttributeValue(ctx context.Context, p tfsdk.Plan, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrType, err := p.Schema.AttributeTypeAtPath(path)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		diags.AddAttributeError(
			path,
			"Plan Read Error",
			"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	// if the whole plan is nil, the value of a valid attribute is also nil
	if p.Raw.IsNull() {
		return nil, nil
	}

	tfValue, err := PlanTerraformValueAtPath(p, path)

	// Ignoring ErrInvalidStep will allow this method to return a null value of the type.
	if err != nil && !errors.Is(err, tftypes.ErrInvalidStep) {
		diags.AddAttributeError(
			path,
			"Plan Read Error",
			"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	// TODO: If ErrInvalidStep, check parent paths for unknown value.
	//       If found, convert this value to an unknown value.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/186

	if attrTypeWithValidate, ok := attrType.(attr.TypeWithValidate); ok {
		logging.FrameworkTrace(ctx, "Type implements TypeWithValidate")
		logging.FrameworkDebug(ctx, "Calling provider defined Type Validate")
		diags.Append(attrTypeWithValidate.Validate(ctx, tfValue, path)...)
		logging.FrameworkDebug(ctx, "Called provider defined Type Validate")

		if diags.HasError() {
			return nil, diags
		}
	}

	attrValue, err := attrType.ValueFromTerraform(ctx, tfValue)

	if err != nil {
		diags.AddAttributeError(
			path,
			"Plan Read Error",
			"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	return attrValue, diags
}

// PlanTerraformValueAtPath is a duplicate of
// tfsdk.Plan.terraformValueAtPath.
//
// TODO: Clean up this abstraction back into an internal Plan type method.
// The extra Plan parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func PlanTerraformValueAtPath(p tfsdk.Plan, path *tftypes.AttributePath) (tftypes.Value, error) {
	rawValue, remaining, err := tftypes.WalkAttributePath(p.Raw, path)
	if err != nil {
		return tftypes.Value{}, fmt.Errorf("%v still remains in the path: %w", remaining, err)
	}
	attrValue, ok := rawValue.(tftypes.Value)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("got non-tftypes.Value result %v", rawValue)
	}
	return attrValue, err
}
