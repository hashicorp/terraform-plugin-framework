package tfsdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/internal/totftypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Plan represents a Terraform plan.
type Plan struct {
	Raw    tftypes.Value
	Schema Schema
}

// Get populates the struct passed as `target` with the entire plan.
func (p Plan) Get(ctx context.Context, target interface{}) diag.Diagnostics {
	return p.data().Get(ctx, target)
}

// GetAttribute retrieves the attribute found at `path` and populates the
// `target` with the value.
func (p Plan) GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics {
	return p.data().GetAtPath(ctx, path, target)
}

// PathMatches returns all matching path.Paths from the given path.Expression.
//
// If a parent path is null or unknown, which would prevent a full expression
// from matching, the parent path is returned rather than no match to prevent
// false positives.
func (p Plan) PathMatches(ctx context.Context, pathExpr path.Expression) (path.Paths, diag.Diagnostics) {
	return p.data().PathMatches(ctx, pathExpr)
}

// Set populates the entire plan using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (p *Plan) Set(ctx context.Context, val interface{}) diag.Diagnostics {
	newPlanAttrValue, diags := reflect.FromValue(ctx, p.Schema.Type(), val, path.Empty())
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
func (p *Plan) SetAttribute(ctx context.Context, path path.Path, val interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = logging.FrameworkWithAttributePath(ctx, path.String())

	tftypesPath, tftypesPathDiags := totftypes.AttributePath(ctx, path)

	diags.Append(tftypesPathDiags...)

	if diags.HasError() {
		return diags
	}

	attrType, err := p.Schema.AttributeTypeAtPath(tftypesPath)
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

	if attrTypeWithValidate, ok := attrType.(xattr.TypeWithValidate); ok {
		logging.FrameworkTrace(ctx, "Type implements TypeWithValidate")
		logging.FrameworkDebug(ctx, "Calling provider defined Type Validate")
		diags.Append(attrTypeWithValidate.Validate(ctx, tfVal, path)...)
		logging.FrameworkDebug(ctx, "Called provider defined Type Validate")

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

func (p Plan) data() fwschemadata.Data {
	return fwschemadata.Data{
		Schema:         p.Schema,
		TerraformValue: p.Raw,
	}
}

// setAttributeTransformFunc recursively creates a value based on the current
// Plan values along the path. If the value at the path does not yet exist,
// this will perform recursion to add the child value to a parent value,
// creating the parent value if necessary.
func (p Plan) setAttributeTransformFunc(ctx context.Context, path path.Path, tfVal tftypes.Value, diags diag.Diagnostics) (transformFunc, diag.Diagnostics) {
	exists, pathExistsDiags := p.data().PathExists(ctx, path)
	diags.Append(pathExistsDiags...)

	if diags.HasError() {
		return nil, diags
	}

	tftypesPath, tftypesPathDiags := totftypes.AttributePath(ctx, path)

	diags.Append(tftypesPathDiags...)

	if diags.HasError() {
		return nil, diags
	}

	if exists {
		// Overwrite existing value
		return func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
			if p.Equal(tftypesPath) {
				return tfVal, nil
			}
			return v, nil
		}, diags
	}

	parentPath := path.ParentPath()
	parentTftypesPath := tftypesPath.WithoutLastStep()
	parentAttrType, err := p.Schema.AttributeTypeAtPath(parentTftypesPath)

	if err != nil {
		err = fmt.Errorf("error getting parent attribute type in schema: %w", err)
		diags.AddAttributeError(
			parentPath,
			"Plan Write Error",
			"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	parentValue, err := p.data().TerraformValueAtTerraformPath(ctx, parentTftypesPath)

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
	childStep, _ := path.Steps().LastStep()
	parentValue, childValueDiags = upsertChildValue(ctx, parentPath, parentValue, childStep, tfVal)
	diags.Append(childValueDiags...)

	if diags.HasError() {
		return nil, diags
	}

	if attrTypeWithValidate, ok := parentAttrType.(xattr.TypeWithValidate); ok {
		logging.FrameworkTrace(ctx, "Type implements TypeWithValidate")
		logging.FrameworkDebug(ctx, "Calling provider defined Type Validate")
		diags.Append(attrTypeWithValidate.Validate(ctx, parentValue, parentPath)...)
		logging.FrameworkDebug(ctx, "Called provider defined Type Validate")

		if diags.HasError() {
			return nil, diags
		}
	}

	return p.setAttributeTransformFunc(ctx, parentPath, parentValue, diags)
}
