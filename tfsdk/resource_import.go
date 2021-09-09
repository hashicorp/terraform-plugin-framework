package tfsdk

import "context"

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
