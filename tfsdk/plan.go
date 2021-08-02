package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Plan represents a Terraform plan.
type Plan struct {
	Raw    tftypes.Value
	Schema Schema
}

// Get populates the struct passed as `target` with the entire plan.
func (p Plan) Get(ctx context.Context, target interface{}) error {
	return reflect.Into(ctx, p.Schema.AttributeType(), p.Raw, target, reflect.Options{})
}

// GetAttribute retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (p Plan) GetAttribute(ctx context.Context, path *tftypes.AttributePath) (attr.Value, error) {
	attrType, err := p.Schema.AttributeTypeAtPath(path)
	if err != nil {
		return nil, fmt.Errorf("error walking schema: %w", err)
	}

	attrValue, err := p.terraformValueAtPath(path)
	if err != nil {
		return nil, fmt.Errorf("error walking plan: %w", err)
	}

	return attrType.ValueFromTerraform(ctx, attrValue)
}

// Set populates the entire plan using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (p *Plan) Set(ctx context.Context, val interface{}) error {
	newPlanAttrValue, err := reflect.OutOf(ctx, p.Schema.AttributeType(), val)
	if err != nil {
		return fmt.Errorf("error creating new plan value: %w", err)
	}

	newPlanVal, err := newPlanAttrValue.ToTerraformValue(ctx)
	if err != nil {
		return fmt.Errorf("error running ToTerraformValue on plan: %w", err)
	}

	newPlan := tftypes.NewValue(p.Schema.AttributeType().TerraformType(ctx), newPlanVal)

	p.Raw = newPlan
	return nil
}

// SetAttribute sets the attribute at `path` using the supplied Go value.
func (p *Plan) SetAttribute(ctx context.Context, path *tftypes.AttributePath, val interface{}) error {
	attrType, err := p.Schema.AttributeTypeAtPath(path)
	if err != nil {
		return fmt.Errorf("error getting attribute type at path %s in schema: %w", path, err)
	}

	newVal, err := reflect.OutOf(ctx, attrType, val)
	if err != nil {
		return fmt.Errorf("error creating new plan value: %w", err)
	}

	newTfVal, err := newVal.ToTerraformValue(ctx)
	if err != nil {
		return fmt.Errorf("error running ToTerraformValue on new plan value: %w", err)
	}

	transformFunc := func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if p.Equal(path) {
			return tftypes.NewValue(attrType.TerraformType(ctx), newTfVal), nil
		}
		return v, nil
	}

	p.Raw, err = tftypes.Transform(p.Raw, transformFunc)
	if err != nil {
		return fmt.Errorf("error setting attribute in plan: %w", err)
	}

	return nil
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
