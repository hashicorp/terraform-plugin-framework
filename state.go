package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// State represents a Terraform state.
type State struct {
	Raw    tftypes.Value
	Schema schema.Schema
}

// Get populates the struct passed as `target` with the entire state.
func (s State) Get(ctx context.Context, target interface{}) error {
	return reflect.Into(ctx, s.Schema.AttributeType(), s.Raw, target, reflect.Options{})
}

// GetAttribute retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (s State) GetAttribute(ctx context.Context, path *tftypes.AttributePath) (attr.Value, error) {
	attrType, err := s.Schema.AttributeTypeAtPath(path)
	if err != nil {
		return nil, fmt.Errorf("error walking schema: %w", err)
	}

	attrValue, err := s.terraformValueAtPath(path)
	if err != nil {
		return nil, fmt.Errorf("error walking state: %w", err)
	}

	return attrType.ValueFromTerraform(ctx, attrValue)
}

// Set populates the entire state using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (s *State) Set(ctx context.Context, val interface{}) error {
	newStateAttrValue, err := reflect.OutOf(ctx, s.Schema.AttributeType(), val)
	if err != nil {
		return fmt.Errorf("error creating new state value: %w", err)
	}

	newStateVal, err := newStateAttrValue.ToTerraformValue(ctx)
	if err != nil {
		return fmt.Errorf("error running ToTerraformValue on state: %w", err)
	}

	newState := tftypes.NewValue(s.Schema.AttributeType().TerraformType(ctx), newStateVal)

	s.Raw = newState
	return nil
}

// SetAttribute sets the attribute at `path` using the supplied Go value.
func (s *State) SetAttribute(ctx context.Context, path *tftypes.AttributePath, val interface{}) error {
	attrType, err := s.Schema.AttributeTypeAtPath(path)
	if err != nil {
		return fmt.Errorf("error getting attribute type at path %s in schema: %w", path, err)
	}

	newVal, err := reflect.OutOf(ctx, attrType, val)
	if err != nil {
		return fmt.Errorf("error creating new state value: %w", err)
	}

	newTfVal, err := newVal.ToTerraformValue(ctx)
	if err != nil {
		return fmt.Errorf("error running ToTerraformValue on new state value: %w", err)
	}

	transformFunc := func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if p.Equal(path) {
			return tftypes.NewValue(attrType.TerraformType(ctx), newTfVal), nil
		}
		return v, nil
	}

	s.Raw, err = tftypes.Transform(s.Raw, transformFunc)
	if err != nil {
		return fmt.Errorf("error setting attribute in state: %w", err)
	}

	return nil
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
