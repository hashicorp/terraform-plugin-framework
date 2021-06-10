package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Plan represents a Terraform plan.
type Plan struct {
	Raw    tftypes.Value
	Schema schema.Schema
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
