package tfsdk

import (
	"context"
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
		diags.Append(attrTypeWithValidate.Validate(ctx, tfValue)...)

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
	newStateAttrValue, diags := reflect.OutOf(ctx, s.Schema.AttributeType(), val)
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

	newVal, newValDiags := reflect.OutOf(ctx, attrType, val)
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

	transformFunc := func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if p.Equal(path) {
			tfVal := tftypes.NewValue(attrType.TerraformType(ctx), newTfVal)

			if attrTypeWithValidate, ok := attrType.(attr.TypeWithValidate); ok {
				diags.Append(attrTypeWithValidate.Validate(ctx, tfVal)...)

				if diags.HasError() {
					return v, nil
				}
			}

			return tfVal, nil
		}
		return v, nil
	}

	s.Raw, err = tftypes.Transform(s.Raw, transformFunc)
	if err != nil {
		diags.AddAttributeError(
			path,
			"State Write Error",
			"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	return diags
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
