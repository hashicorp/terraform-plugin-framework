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
