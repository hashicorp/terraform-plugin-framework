package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Config represents a Terraform config.
type Config struct {
	Raw    tftypes.Value
	Schema schema.Schema
}

// Get populates the struct passed as `target` with the entire config.
func (c Config) Get(ctx context.Context, target interface{}) error {
	return reflect.Into(ctx, c.Schema.AttributeType(), c.Raw, target, reflect.Options{})
}

// GetAttribute retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (c Config) GetAttribute(ctx context.Context, path *tftypes.AttributePath) (attr.Value, error) {
	attrType, err := c.Schema.AttributeTypeAtPath(path)
	if err != nil {
		return nil, fmt.Errorf("error walking schema: %w", err)
	}

	attrValue, err := c.terraformValueAtPath(path)
	if err != nil {
		return nil, fmt.Errorf("error walking config: %w", err)
	}

	return attrType.ValueFromTerraform(ctx, attrValue)
}

func (c Config) terraformValueAtPath(path *tftypes.AttributePath) (tftypes.Value, error) {
	rawValue, remaining, err := tftypes.WalkAttributePath(c.Raw, path)
	if err != nil {
		return tftypes.Value{}, fmt.Errorf("%v still remains in the path: %w", remaining, err)
	}
	attrValue, ok := rawValue.(tftypes.Value)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("got non-tftypes.Value result %v", rawValue)
	}
	return attrValue, err
}
