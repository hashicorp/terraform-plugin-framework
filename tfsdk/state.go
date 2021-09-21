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

// State represents a Terraform state.
type State struct {
	Raw    tftypes.Value
	Schema Schema
}

// Get populates the struct passed as `target` with the entire state.
func (s State) Get(ctx context.Context, target interface{}) diag.Diagnostics {
	return reflect.Into(ctx, s.Schema.AttributeType(), s.Raw, target, reflect.Options{})
}

// GetAttribute retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (s State) GetAttribute(ctx context.Context, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrType, err := s.Schema.AttributeTypeAtPath(path)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		diags.AddAttributeError(
			path,
			"State Read Error",
			"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	// if the whole state is nil, the value of a valid attribute is also nil
	if s.Raw.IsNull() {
		return nil, nil
	}

	tfValue, err := s.terraformValueAtPath(path)
	if err != nil {
		diags.AddAttributeError(
			path,
			"State Read Error",
			"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
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
			"State Read Error",
			"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	return attrValue, diags
}

// Set populates the entire state using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (s *State) Set(ctx context.Context, val interface{}) diag.Diagnostics {
	if val == nil {
		err := fmt.Errorf("cannot set nil as entire state; to remove a resource from state, call State.RemoveResource, instead")
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"State Read Error",
				"An unexpected error was encountered trying to write the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			),
		}
	}
	newStateAttrValue, diags := reflect.FromValue(ctx, s.Schema.AttributeType(), val, tftypes.NewAttributePath())
	if diags.HasError() {
		return diags
	}

	newStateVal, err := newStateAttrValue.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on state: %w", err)
		diags.AddError(
			"State Write Error",
			"An unexpected error was encountered trying to write the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	newState := tftypes.NewValue(s.Schema.AttributeType().TerraformType(ctx), newStateVal)

	s.Raw = newState
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
func (s *State) SetAttribute(ctx context.Context, path *tftypes.AttributePath, val interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	attrType, err := s.Schema.AttributeTypeAtPath(path)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		diags.AddAttributeError(
			path,
			"State Write Error",
			"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
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
		err = fmt.Errorf("error running ToTerraformValue on new state value: %w", err)
		diags.AddAttributeError(
			path,
			"State Write Error",
			"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
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

	transformFunc, transformFuncDiags := s.setAttributeTransformFunc(ctx, path, tfVal, nil)
	diags.Append(transformFuncDiags...)

	if diags.HasError() {
		return diags
	}

	s.Raw, err = tftypes.Transform(s.Raw, transformFunc)
	if err != nil {
		err = fmt.Errorf("Cannot transform state: %w", err)
		diags.AddAttributeError(
			path,
			"State Write Error",
			"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	return diags
}

// pathExists walks the current state and returns true if the path can be reached.
// The value at the path may be null or unknown.
func (s State) pathExists(ctx context.Context, path *tftypes.AttributePath) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	_, remaining, err := tftypes.WalkAttributePath(s.Raw, path)

	if err != nil {
		if errors.Is(err, tftypes.ErrInvalidStep) {
			return false, diags
		}

		diags.AddAttributeError(
			path,
			"State Read Error",
			"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Cannot walk attribute path in state: %s", err),
		)
		return false, diags
	}

	return len(remaining.Steps()) == 0, diags
}

func (s State) setAttributeTransformFunc(ctx context.Context, path *tftypes.AttributePath, tfVal tftypes.Value, diags diag.Diagnostics) (func(*tftypes.AttributePath, tftypes.Value) (tftypes.Value, error), diag.Diagnostics) {
	exists, pathExistsDiags := s.pathExists(ctx, path)
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

	var parentTfVal tftypes.Value
	parentPath := path.WithoutLastStep()
	parentAttrType, err := s.Schema.AttributeTypeAtPath(parentPath)

	if err != nil {
		err = fmt.Errorf("error getting parent attribute type in schema: %w", err)
		diags.AddAttributeError(
			parentPath,
			"State Write Error",
			"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	parentTfType := parentAttrType.TerraformType(ctx)
	parentValue, err := s.terraformValueAtPath(parentPath)

	if err != nil && !errors.Is(err, tftypes.ErrInvalidStep) {
		diags.AddAttributeError(
			parentPath,
			"State Read Error",
			"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
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

	switch step := path.Steps()[len(path.Steps())-1].(type) {
	case tftypes.AttributeName:
		// Add to Object
		if !parentValue.Type().Is(tftypes.Object{}) {
			diags.AddAttributeError(
				parentPath,
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Cannot add attribute into parent type: %s", parentValue.Type()),
			)
			return nil, diags
		}

		var parentAttrs map[string]tftypes.Value
		err = parentValue.Copy().As(&parentAttrs)

		if err != nil {
			diags.AddAttributeError(
				parentPath,
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
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
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Cannot add list element into parent type: %s", parentValue.Type()),
			)
			return nil, diags
		}

		var parentElems []tftypes.Value
		err = parentValue.Copy().As(&parentElems)

		if err != nil {
			diags.AddAttributeError(
				parentPath,
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Unable to extract list elements from parent value: %s", err),
			)
			return nil, diags
		}

		if int(step) > len(parentElems) {
			diags.AddAttributeError(
				parentPath,
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
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
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Cannot add map value into parent type: %s", parentValue.Type()),
			)
			return nil, diags
		}

		var parentElems map[string]tftypes.Value
		err = parentValue.Copy().As(&parentElems)

		if err != nil {
			diags.AddAttributeError(
				parentPath,
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
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
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Cannot add set element into parent type: %s", parentValue.Type()),
			)
			return nil, diags
		}

		var parentElems []tftypes.Value
		err = parentValue.Copy().As(&parentElems)

		if err != nil {
			diags.AddAttributeError(
				parentPath,
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
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

	return s.setAttributeTransformFunc(ctx, parentPath, parentTfVal, diags)
}

// RemoveResource removes the entire resource from state.
func (s *State) RemoveResource(ctx context.Context) {
	s.Raw = tftypes.NewValue(s.Schema.TerraformType(ctx), nil)
}

func (s State) terraformValueAtPath(path *tftypes.AttributePath) (tftypes.Value, error) {
	rawValue, remaining, err := tftypes.WalkAttributePath(s.Raw, path)
	if err != nil {
		return tftypes.Value{}, fmt.Errorf("%v still remains in the path: %w", remaining, err)
	}
	attrValue, ok := rawValue.(tftypes.Value)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("got non-tftypes.Value result %v", rawValue)
	}
	return attrValue, err
}
