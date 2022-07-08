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
			// If there was an error with conversion of the path at this level,
			// no need to traverse further since a deeper path will error.
			return false, nil
		}

		if pathExpr.Matches(fwPath) {
			paths.Append(fwPath)

			// If we matched, there is no need to traverse further since a
			// deeper path will never match.
			return false, nil
		}

		// If current path cannot be parent path, there is no need to traverse
		// further since a deeper path will never match.
		if !pathExpr.MatchesParent(fwPath) {
			return false, nil
		}

		// If value at current path (now known to be a parent path of the
		// expression) is null or unknown, return it as a valid path match
		// since Walk will stop traversing deeper anyways and we want
		// consumers to know about the path with the null or unknown value.
		//
		// This behavior may be confusing for consumers as fetching the value
		// at this parent path will return a potentially unexpected type,
		// however this is an implementation tradeoff to prevent false
		// positives of missing null or unknown values.
		if tfTypeValue.IsNull() || !tfTypeValue.IsKnown() {
			paths.Append(fwPath)

			return false, nil
		}

		return true, nil
	})

	// If there were no matches, including no parent path matches, it should be
	// safe to assume the path expression was invalid for the schema.
	if len(paths) == 0 {
		diags.AddError(
			"Invalid Path Expression for Schema Data",
			"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
				"This can happen if the path expression does not correctly follow the schema in structure or types. "+
				"Please report this to the provider developers.\n\n"+
				"Path Expression: "+pathExpr.String(),
		)
	}

	return paths, diags
}
