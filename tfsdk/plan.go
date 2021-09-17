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

// GetAttribute retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (p Plan) GetAttribute(ctx context.Context, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
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
	if err != nil {
		diags.AddAttributeError(
			path,
			"Plan Read Error",
			"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

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

	newPlanVal, err := newPlanAttrValue.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on plan: %w", err)
		diags.AddError(
			"Plan Write Error",
			"An unexpected error was encountered trying to write the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	newPlan := tftypes.NewValue(p.Schema.AttributeType().TerraformType(ctx), newPlanVal)

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
// Lists can only have the first element added if empty and can only add the
// next element according to the current length, otherwise this will return an
// error.
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

	newTfVal, err := newVal.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on new plan value: %w", err)
		diags.AddAttributeError(
			path,
			"Plan Write Error",
			"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	tfVal := tftypes.NewValue(attrType.TerraformType(ctx), newTfVal)

	if attrTypeWithValidate, ok := attrType.(attr.TypeWithValidate); ok {
		diags.Append(attrTypeWithValidate.Validate(ctx, tfVal, path)...)

		if diags.HasError() {
			return diags
		}
	}

	transformFunc, transformFuncDiags := p.setAttributeTransformFunc(ctx, path, tfVal)
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

func (p Plan) setAttributeTransformFunc(ctx context.Context, path *tftypes.AttributePath, tfVal tftypes.Value) (func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error), diag.Diagnostics) {
	var diags diag.Diagnostics

	_, remaining, err := tftypes.WalkAttributePath(p.Raw, path)

	if err != nil && !errors.Is(err, tftypes.ErrInvalidStep) {
		err = fmt.Errorf("Cannot walk attribute path in plan: %w", err)
		diags.AddAttributeError(
			path,
			"Plan Write Error",
			"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	if len(remaining.Steps()) == 0 {
		// Overwrite existing value
		return func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
			if p.Equal(path) {
				return tfVal, nil
			}
			return v, nil
		}, diags
	}

	var parentTfVal tftypes.Value
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

	parentTfType := parentAttrType.TerraformType(ctx)
	parentValue, err := p.terraformValueAtPath(parentPath)

	if err != nil && !errors.Is(err, tftypes.ErrInvalidStep) {
		diags.AddAttributeError(
			parentPath,
			"Plan Read Error",
			"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	// Check if parent needs to get created
	if parentValue.Equal(tftypes.NewValue(tftypes.Object{}, nil)) {
		// NewValue will panic if required attributes are missing in the
		// tftypes.Object.
		vals := map[string]tftypes.Value{}
		for name, t := range parentTfType.(tftypes.Object).AttributeTypes {
			vals[name] = tftypes.NewValue(t, nil)
		}
		parentValue = tftypes.NewValue(parentTfType, vals)
	} else if parentValue.Equal(tftypes.Value{}) {
		parentValue = tftypes.NewValue(parentTfType, nil)
	}

	switch step := remaining.Steps()[len(remaining.Steps())-1].(type) {
	case tftypes.AttributeName:
		// Add to Object
		if !parentValue.Type().Is(tftypes.Object{}) {
			diags.AddAttributeError(
				parentPath,
				"Plan Write Error",
				"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Cannot add attribute into parent type: %s", parentValue.Type()),
			)
			return nil, diags
		}

		var parentAttrs map[string]tftypes.Value
		err = parentValue.Copy().As(&parentAttrs)

		if err != nil {
			diags.AddAttributeError(
				parentPath,
				"Plan Write Error",
				"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Unable to extract object elements from parent value: %s", err),
			)
			return nil, diags
		}

		parentAttrs[string(step)] = tfVal
		parentTfVal = tftypes.NewValue(parentTfType, parentAttrs)
	case tftypes.ElementKeyInt:
		// Add new List element
		if !parentValue.Type().Is(tftypes.List{}) {
			diags.AddAttributeError(
				parentPath,
				"Plan Write Error",
				"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Cannot add list element into parent type: %s", parentValue.Type()),
			)
			return nil, diags
		}

		var parentElems []tftypes.Value
		err = parentValue.Copy().As(&parentElems)

		if err != nil {
			diags.AddAttributeError(
				parentPath,
				"Plan Write Error",
				"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Unable to extract list elements from parent value: %s", err),
			)
			return nil, diags
		}

		if int(step) > len(parentElems) {
			diags.AddAttributeError(
				parentPath,
				"Plan Write Error",
				"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Cannot add list element %d as list currently has %d length. To prevent ambiguity, SetAttribute can only add the next element to a list. Add empty elements into the list prior to this call, if appropriate.", int(step)+1, len(parentElems)),
			)
			return nil, diags
		}

		parentElems = append(parentElems, tfVal)
		parentTfVal = tftypes.NewValue(parentTfType, parentElems)
	case tftypes.ElementKeyString:
		// Add new Map element
		if !parentValue.Type().Is(tftypes.Map{}) {
			diags.AddAttributeError(
				parentPath,
				"Plan Write Error",
				"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Cannot add map value into parent type: %s", parentValue.Type()),
			)
			return nil, diags
		}

		var parentElems map[string]tftypes.Value
		err = parentValue.Copy().As(&parentElems)

		if err != nil {
			diags.AddAttributeError(
				parentPath,
				"Plan Write Error",
				"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Unable to extract map elements from parent value: %s", err),
			)
			return nil, diags
		}

		parentElems[string(step)] = tfVal
		parentTfVal = tftypes.NewValue(parentTfType, parentElems)
	case tftypes.ElementKeyValue:
		// Add new Set element
		if !parentValue.Type().Is(tftypes.Set{}) {
			diags.AddAttributeError(
				parentPath,
				"Plan Write Error",
				"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Cannot add set element into parent type: %s", parentValue.Type()),
			)
			return nil, diags
		}

		var parentElems []tftypes.Value
		err = parentValue.Copy().As(&parentElems)

		if err != nil {
			diags.AddAttributeError(
				parentPath,
				"Plan Write Error",
				"An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Unable to extract set elements from parent value: %s", err),
			)
			return nil, diags
		}

		parentElems = append(parentElems, tfVal)
		parentTfVal = tftypes.NewValue(parentTfType, parentElems)
	}

	if attrTypeWithValidate, ok := parentAttrType.(attr.TypeWithValidate); ok {
		diags.Append(attrTypeWithValidate.Validate(ctx, parentTfVal, parentPath)...)

		if diags.HasError() {
			return nil, diags
		}
	}

	if len(remaining.Steps()) == 1 {
		return func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
			if p.Equal(parentPath) {
				return parentTfVal, nil
			}
			return v, nil
		}, diags
	}

	return p.setAttributeTransformFunc(ctx, parentPath, parentTfVal)
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
