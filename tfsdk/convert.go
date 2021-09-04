package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ConvertValue creates a new attr.Value of the attr.Type `typ`, using the data
// in `val`, which can be of any attr.Type so long as its TerraformType method
// returns a tftypes.Type that `typ`'s ValueFromTerraform method can accept.
func ConvertValue(ctx context.Context, val attr.Value, typ attr.Type) (attr.Value, error) {
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
