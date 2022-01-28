package tfsdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Plan represents a Terraform plan.
type Plan struct {
	Raw    tftypes.Value
	Schema Schema
}

// Get populates the struct passed as `target` with the entire plan.
func (p Plan) Get(ctx context.Context, target interface{}) diag.Diagnostics {
	return reflect.Into(ctx, p.Schema.AttributeType(), p.Raw, target, reflect.Options{})
}

// GetAttribute retrieves the attribute found at `path` and populates the
// `target` with the value.
func (p Plan) GetAttribute(ctx context.Context, path *tftypes.AttributePath, target interface{}) diag.Diagnostics {
	attrValue, diags := p.getAttributeValue(ctx, path)

	if diags.HasError() {
		return diags
	}

	if attrValue == nil {
		diags.AddAttributeError(
			path,
			"Plan Read Error",
			"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				"Missing attribute value, however no error was returned. Preventing the panic from this situation.",
		)
		return diags
	}

	valueAsDiags := ValueAs(ctx, attrValue, target)

	// ValueAs does not have path information for its Diagnostics.
	for idx, valueAsDiag := range valueAsDiags {
		valueAsDiags[idx] = diag.WithPath(path, valueAsDiag)
	}

	diags.Append(valueAsDiags...)

	return diags
}

// getAttributeValue retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (p Plan) getAttributeValue(ctx context.Context, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
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

	tfValue, err := p.terraformValueAtPath(path)

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
		diags.Append(attrTypeWithValidate.Validate(ctx, tfValue, path)...)

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

// Set populates the entire plan using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (p *Plan) Set(ctx context.Context, val interface{}) diag.Diagnostics {
	newPlanAttrValue, diags := reflect.FromValue(ctx, p.Schema.AttributeType(), val, tftypes.NewAttributePath())
	if diags.HasError() {
		return diags
	}

	newPlan, err := newPlanAttrValue.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on plan: %w", err)
		diags.AddError(
			"Plan Write Error",
			"An unexpected error was encountered trying to write the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	p.Raw = newPlan
	return diags
}

// SetAttribute sets the attribute at `path` using the supplied Go value.
//
// The attribute path and value must be valid with the current schema. If the
// attribute path already has a value, it will be overwritten. If the attribute
// path does not have a value, it will be added, including any parent attribute
// paths as necessary.
//
// Lists can only have the next element added according to the current length.
func (p *Plan) SetAttribute(ctx context.Context, path *tftypes.AttributePath, val interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	attrType, err := p.Schema.AttributeTypeAtPath(path)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		diags.AddAttributeError(
			path,
			"Plan Write Error",
			"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	newVal, newValDiags := reflect.FromValue(ctx, attrType, val, path)
	diags.Append(newValDiags...)

	if diags.HasError() {
		return diags
	}

	tfVal, err := newVal.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on new plan value: %w", err)
		diags.AddAttributeError(
			path,
			"Plan Write Error",
			"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	if attrTypeWithValidate, ok := attrType.(attr.TypeWithValidate); ok {
		diags.Append(attrTypeWithValidate.Validate(ctx, tfVal, path)...)

		if diags.HasError() {
			return diags
		}
	}

	transformFunc, transformFuncDiags := p.setAttributeTransformFunc(ctx, path, tfVal, nil)
	diags.Append(transformFuncDiags...)

	if diags.HasError() {
		return diags
	}

	p.Raw, err = tftypes.Transform(p.Raw, transformFunc)
	if err != nil {
		err = fmt.Errorf("Cannot transform plan: %w", err)
		diags.AddAttributeError(
			path,
			"Plan Write Error",
			"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	return diags
}

// pathExists walks the current state and returns true if the path can be reached.
// The value at the path may be null or unknown.
func (p Plan) pathExists(_ context.Context, path *tftypes.AttributePath) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	_, remaining, err := tftypes.WalkAttributePath(p.Raw, path)

	if err != nil {
		if errors.Is(err, tftypes.ErrInvalidStep) {
			return false, diags
		}

		diags.AddAttributeError(
			path,
			"Plan Read Error",
			"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Cannot walk attribute path in plan: %s", err),
		)
		return false, diags
	}

	return len(remaining.Steps()) == 0, diags
}

// setAttributeTransformFunc recursively creates a value based on the current
// Plan values along the path. If the value at the path does not yet exist,
// this will perform recursion to add the child value to a parent value,
// creating the parent value if necessary.
func (p Plan) setAttributeTransformFunc(ctx context.Context, path *tftypes.AttributePath, tfVal tftypes.Value, diags diag.Diagnostics) (transformFunc, diag.Diagnostics) {
	exists, pathExistsDiags := p.pathExists(ctx, path)
	diags.Append(pathExistsDiags...)

	if diags.HasError() {
		return nil, diags
	}

	if exists {
		// Overwrite existing value
		return func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
			if p.Equal(path) {
				return tfVal, nil
			}
			return v, nil
		}, diags
	}

	parentPath := path.WithoutLastStep()
	parentAttrType, err := p.Schema.AttributeTypeAtPath(parentPath)

	if err != nil {
		err = fmt.Errorf("error getting parent attribute type in schema: %w", err)
		diags.AddAttributeError(
			parentPath,
			"Plan Write Error",
			"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	parentValue, err := p.terraformValueAtPath(parentPath)

	if err != nil && !errors.Is(err, tftypes.ErrInvalidStep) {
		diags.AddAttributeError(
			parentPath,
			"Plan Read Error",
			"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	if parentValue.IsNull() || !parentValue.IsKnown() {
		// TODO: This will break when DynamicPsuedoType is introduced.
		// tftypes.Type should implement AttributePathStepper, but it currently does not.
		// When it does, we should use: tftypes.WalkAttributePath(p.Raw.Type(), parentPath)
		// Reference: https://github.com/hashicorp/terraform-plugin-go/issues/110
		parentType := parentAttrType.TerraformType(ctx)
		var childValue interface{}

		if !parentValue.IsKnown() {
			childValue = tftypes.UnknownValue
		}

		var parentValueDiags diag.Diagnostics
		parentValue, parentValueDiags = createParentValue(ctx, parentPath, parentType, childValue)
		diags.Append(parentValueDiags...)

		if diags.HasError() {
			return nil, diags
		}
	}

	var childValueDiags diag.Diagnostics
	childStep := path.LastStep()
	parentValue, childValueDiags = upsertChildValue(ctx, parentPath, parentValue, childStep, tfVal)
	diags.Append(childValueDiags...)

	if diags.HasError() {
		return nil, diags
	}

	if attrTypeWithValidate, ok := parentAttrType.(attr.TypeWithValidate); ok {
		diags.Append(attrTypeWithValidate.Validate(ctx, parentValue, parentPath)...)

		if diags.HasError() {
			return nil, diags
		}
	}

	return p.setAttributeTransformFunc(ctx, parentPath, parentValue, diags)
}

func (p Plan) terraformValueAtPath(path *tftypes.AttributePath) (tftypes.Value, error) {
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
