package tfsdk

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ValueFrom populates the attr value passed as `val` with
// the contents of `target`, using the reflection rules
// defined for `Get` and `GetAttribute`.
func ValueFrom(ctx context.Context, typ attr.Type, val *attr.Value, target interface{}) diag.Diagnostics {
	v, diags := reflect.FromValue(ctx, typ, target, tftypes.NewAttributePath())
	if diags.HasError() {
		return diags
	}

	*val = v
	return diags
}
