package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ValidateResourceConfigRequest returns the *fwserver.ValidateResourceConfigRequest
// equivalent of a *tfprotov6.ValidateResourceConfigRequest.
func ValidateResourceConfigRequest(ctx context.Context, proto6 *tfprotov6.ValidateResourceConfigRequest, resourceType tfsdk.ResourceType, resourceSchema *tfsdk.Schema) (*fwserver.ValidateResourceConfigRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	fw := &fwserver.ValidateResourceConfigRequest{}

	config, diags := Config(ctx, proto6.Config, resourceSchema)

	fw.Config = config
	fw.ResourceType = resourceType

	return fw, diags
}
