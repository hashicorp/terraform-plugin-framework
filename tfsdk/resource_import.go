package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Optional interface on top of Resource that enables provider control over
// the ImportResourceState RPC. This RPC is called by Terraform when the
// `terraform import` command is executed. Afterwards, the ReadResource RPC
// is executed to allow providers to fully populate the resource state.
type ResourceWithImportState interface {
	// ImportState is called when the provider must import the state of a
	// resource instance. This method must return enough state so the Read
	// method can properly refresh the full resource.
	//
	// If setting an attribute with the import identifier, it is recommended
	// to use the ResourceImportStatePassthroughID() call in this method.
	ImportState(context.Context, ImportResourceStateRequest, *ImportResourceStateResponse)
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
