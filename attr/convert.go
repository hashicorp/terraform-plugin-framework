package attr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Convert creates a new Value of the Type `typ`, using the data in `val`,
// which can be of any Type so long as its TerraformType method returns a
// tftypes.Type that `typ`'s ValueFromTerraform method can accept.
func Convert(ctx context.Context, val Value, typ Type) (Value, error) {
	tftype := typ.TerraformType(ctx)
	tfval, err := val.ToTerraformValue(ctx)
	if err != nil {
		return nil, err
	}
	err = tftypes.ValidateValue(tftype, tfval)
	if err != nil {
		return nil, err
	}
	newVal := tftypes.NewValue(tftype, tfval)
	return typ.ValueFromTerraform(ctx, newVal)
}
