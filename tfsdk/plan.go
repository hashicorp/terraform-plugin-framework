package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/diagnostics"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Plan represents a Terraform plan.
type Plan struct {
	Raw    tftypes.Value
	Schema Schema
}

// Get populates the struct passed as `target` with the entire plan.
func (p Plan) Get(ctx context.Context, target interface{}) []*tfprotov6.Diagnostic {
	return reflect.Into(ctx, p.Schema.AttributeType(), p.Raw, target, reflect.Options{})
}

// GetAttribute retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (p Plan) GetAttribute(ctx context.Context, path *tftypes.AttributePath) (attr.Value, []*tfprotov6.Diagnostic) {
	var diags []*tfprotov6.Diagnostic

	attrType, err := p.Schema.AttributeTypeAtPath(path)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		return nil, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Plan Read Error",
			Detail:    "An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	// if the whole plan is nil, the value of a valid attribute is also nil
	if p.Raw.IsNull() {
		return nil, nil
	}

	tfValue, err := p.terraformValueAtPath(path)
	if err != nil {
		return nil, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Plan Read Error",
			Detail:    "An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	if attrTypeWithValidate, ok := attrType.(attr.TypeWithValidate); ok {
		diags = append(diags, attrTypeWithValidate.Validate(ctx, tfValue)...)

		if diagnostics.DiagsHasErrors(diags) {
			return nil, diags
		}
	}

	attrValue, err := attrType.ValueFromTerraform(ctx, tfValue)

	if err != nil {
		return nil, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Plan Read Error",
			Detail:    "An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	return attrValue, diags
}

// Set populates the entire plan using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (p *Plan) Set(ctx context.Context, val interface{}) []*tfprotov6.Diagnostic {
	newPlanAttrValue, diags := reflect.OutOf(ctx, p.Schema.AttributeType(), val)
	if diagnostics.DiagsHasErrors(diags) {
		return diags
	}

	newPlanVal, err := newPlanAttrValue.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on plan: %w", err)
		return append(diags, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Plan Write Error",
			Detail:   "An unexpected error was encountered trying to write the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
		})
	}

	newPlan := tftypes.NewValue(p.Schema.AttributeType().TerraformType(ctx), newPlanVal)

	p.Raw = newPlan
	return diags
}

// SetAttribute sets the attribute at `path` using the supplied Go value.
func (p *Plan) SetAttribute(ctx context.Context, path *tftypes.AttributePath, val interface{}) []*tfprotov6.Diagnostic {
	var diags []*tfprotov6.Diagnostic

	attrType, err := p.Schema.AttributeTypeAtPath(path)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		return append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Plan Write Error",
			Detail:    "An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	newVal, newValDiags := reflect.OutOf(ctx, attrType, val)
	diags = append(diags, newValDiags...)

	if diagnostics.DiagsHasErrors(diags) {
		return diags
	}

	newTfVal, err := newVal.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on new plan value: %w", err)
		return append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Plan Write Error",
			Detail:    "An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	transformFunc := func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if p.Equal(path) {
			tfVal := tftypes.NewValue(attrType.TerraformType(ctx), newTfVal)

			if attrTypeWithValidate, ok := attrType.(attr.TypeWithValidate); ok {
				diags = append(diags, attrTypeWithValidate.Validate(ctx, tfVal)...)

				if diagnostics.DiagsHasErrors(diags) {
					return v, nil
				}
			}

			return tfVal, nil
		}
		return v, nil
	}

	p.Raw, err = tftypes.Transform(p.Raw, transformFunc)
	if err != nil {
		return append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Plan Write Error",
			Detail:    "An unexpected error was encountered trying to write an attribute to the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	return diags
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
