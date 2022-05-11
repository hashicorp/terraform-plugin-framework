package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ConfigureProvider implements the framework server ConfigureProvider RPC.
func (s *Server) ConfigureProvider(ctx context.Context, req *tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	logging.FrameworkDebug(ctx, "Calling provider defined Provider Configure")

	if req != nil {
		s.Provider.Configure(ctx, *req, resp)
	} else {
		s.Provider.Configure(ctx, tfsdk.ConfigureProviderRequest{}, resp)
	}

	logging.FrameworkDebug(ctx, "Called provider defined Provider Configure")
}
