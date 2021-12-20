package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
)

// ValueAs populates the Go value passed as `target` with
// the contents of `val`, using the reflection rules
// defined for `Get` and `GetAttribute`.
func ValueAs(ctx context.Context, val attr.Value, target interface{}) diag.Diagnostics {
	if reflect.IsGenericAttrValue(ctx, target) {
		*(target.(*attr.Value)) = val
		return nil
	}
	raw, err := val.ToTerraformValue(ctx)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Error converting value",
			fmt.Sprintf("An unexpected error was encountered converting a %T to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", val, err))}
	}
	return reflect.Into(ctx, val.Type(ctx), raw, target, reflect.Options{})
}
