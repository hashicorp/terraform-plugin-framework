package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// pathMatches returns all matching path.Paths from the given path.Expression.
//
// TODO: This function should be part of a internal/schemadata package
// except that doing so would currently introduce an import cycle due to the
// Schema parameter here and Config/Plan/State.PathMatches needing to
// call this function until the schema data is migrated to attr.Value.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/172
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/365
func pathMatches(ctx context.Context, schema Schema, tfTypeValue tftypes.Value, pathExpr path.Expression) (path.Paths, diag.Diagnostics) {
	var diags diag.Diagnostics
	var paths path.Paths

	_ = tftypes.Walk(tfTypeValue, func(tfTypePath *tftypes.AttributePath, tfTypeValue tftypes.Value) (bool, error) {
		fwPath, fwPathDiags := attributePath(ctx, tfTypePath, schema)

		diags.Append(fwPathDiags...)

		if diags.HasError() {
			return false, nil
		}

		if pathExpr.Matches(fwPath) {
			paths = append(paths, fwPath)
		}

		return true, nil
	})

	return paths, diags
}
