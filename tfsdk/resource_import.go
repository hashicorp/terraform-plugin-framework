package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ResourceImportStateNotImplemented is a helper function to return an error
// diagnostic about the resource not supporting import. The details defaults
// to a generic message to contact the provider developer, but can be
// customized to provide specific information or recommendations.
func ResourceImportStateNotImplemented(ctx context.Context, details string, resp *ImportResourceStateResponse) {
	if details == "" {
		details = "This resource does not support import. Please contact the provider developer for additional information."
	}

	resp.Diagnostics.AddError(
		"Resource Import Not Implemented",
		details,
	)
}

// ResourceImportStatePassthroughID is a helper function to set the import
// identifier to a given state attribute path. The attribute must accept a
// string value.
func ResourceImportStatePassthroughID(ctx context.Context, path *tftypes.AttributePath, req ImportResourceStateRequest, resp *ImportResourceStateResponse) {
	if path == nil || tftypes.NewAttributePath().Equal(path) {
		resp.Diagnostics.AddError(
			"Resource Import Passthrough Missing Attribute Path",
			"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				"Resource ImportState method call to ResourceImportStatePassthroughID path must be set to a valid attribute path that can accept a string value.",
		)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path, req.ID)...)
}
