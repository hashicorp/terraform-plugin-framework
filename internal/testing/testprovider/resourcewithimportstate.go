package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Resource = &ResourceWithImportState{}
var _ tfsdk.ResourceWithImportState = &ResourceWithImportState{}

// Declarative tfsdk.ResourceWithImportState for unit testing.
type ResourceWithImportState struct {
	*Resource

	// ResourceWithImportState interface methods
	ImportStateMethod func(context.Context, tfsdk.ImportResourceStateRequest, *tfsdk.ImportResourceStateResponse)
}

// ImportState satisfies the tfsdk.ResourceWithImportState interface.
func (p *ResourceWithImportState) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	if p.ImportStateMethod == nil {
		return
	}

	p.ImportStateMethod(ctx, req, resp)
}
