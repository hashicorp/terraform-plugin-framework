package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// State represents a Terraform state.
type State struct {
	Raw    tftypes.Value
	Schema Schema
}

// Get populates the struct passed as `target` with the entire state.
func (s State) Get(ctx context.Context, target interface{}) []*tfprotov6.Diagnostic {
	return reflect.Into(ctx, s.Schema.AttributeType(), s.Raw, target, reflect.Options{})
}

// GetAttribute retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (s State) GetAttribute(ctx context.Context, path *tftypes.AttributePath) (attr.Value, []*tfprotov6.Diagnostic) {
	var diags []*tfprotov6.Diagnostic

	attrType, err := s.Schema.AttributeTypeAtPath(path)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		return nil, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "State Read Error",
			Detail:    "An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	tfValue, err := s.terraformValueAtPath(path)
	if err != nil {
		return nil, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "State Read Error",
			Detail:    "An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	if attrTypeWithValidate, ok := attrType.(attr.TypeWithValidate); ok {
		diags = append(diags, attrTypeWithValidate.Validate(ctx, tfValue)...)

		if diagsHasErrors(diags) {
			return nil, diags
		}
	}

	attrValue, err := attrType.ValueFromTerraform(ctx, tfValue)

	if err != nil {
		return nil, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "State Read Error",
			Detail:    "An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	return attrValue, diags
}

// Set populates the entire state using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (s *State) Set(ctx context.Context, val interface{}) []*tfprotov6.Diagnostic {
	if val == nil {
		err := fmt.Errorf("cannot set nil as entire state; to remove a resource from state, call State.RemoveResource, instead")
		return []*tfprotov6.Diagnostic{
			{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "State Read Error",
				Detail:   "An unexpected error was encountered trying to write the state. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			},
		}
	}
	newStateAttrValue, diags := reflect.OutOf(ctx, s.Schema.AttributeType(), val)
	if diagsHasErrors(diags) {
		return diags
	}

	newStateVal, err := newStateAttrValue.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on state: %w", err)
		return append(diags, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "State Write Error",
			Detail:   "An unexpected error was encountered trying to write the state. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
		})
	}

	newState := tftypes.NewValue(s.Schema.AttributeType().TerraformType(ctx), newStateVal)

	s.Raw = newState
	return nil
}

// SetAttribute sets the attribute at `path` using the supplied Go value.
func (s *State) SetAttribute(ctx context.Context, path *tftypes.AttributePath, val interface{}) []*tfprotov6.Diagnostic {
	var diags []*tfprotov6.Diagnostic

	attrType, err := s.Schema.AttributeTypeAtPath(path)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		return append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "State Write Error",
			Detail:    "An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	newVal, newValDiags := reflect.OutOf(ctx, attrType, val)
	diags = append(diags, newValDiags...)

	if diagsHasErrors(diags) {
		return diags
	}

	newTfVal, err := newVal.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on new state value: %w", err)
		return append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "State Write Error",
			Detail:    "An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	transformFunc := func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if p.Equal(path) {
			tfVal := tftypes.NewValue(attrType.TerraformType(ctx), newTfVal)

			if attrTypeWithValidate, ok := attrType.(attr.TypeWithValidate); ok {
				diags = append(diags, attrTypeWithValidate.Validate(ctx, tfVal)...)

				if diagsHasErrors(diags) {
					return v, nil
				}
			}

			return tfVal, nil
		}
		return v, nil
	}

	s.Raw, err = tftypes.Transform(s.Raw, transformFunc)
	if err != nil {
		return append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "State Write Error",
			Detail:    "An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
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
