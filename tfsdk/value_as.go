package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ValueAs populates the Go value passed as `target` with
// the contents of `val`, using the reflection rules
// defined for `Get` and `GetAttribute`.
func ValueAs(ctx context.Context, val attr.Value, target interface{}) diag.Diagnostics {
	raw, err := val.ToTerraformValue(ctx)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Error converting value",
			fmt.Sprintf("An unexpected error was encountered converting a %T to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", val, err))}
	}
	typ := val.Type(ctx).TerraformType(ctx)
	err = tftypes.ValidateValue(typ, raw)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Invalid value conversion",
			fmt.Sprintf("An unexpected error was encountered converting a %T to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", val, err))}
	}
	v := tftypes.NewValue(typ, raw)
	return reflect.Into(ctx, val.Type(ctx), v, target, reflect.Options{})
}
