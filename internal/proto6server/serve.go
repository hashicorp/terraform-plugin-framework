package proto6server

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var _ tfprotov6.ProviderServer = &Server{}

// Provider server implementation.
type Server struct {
	FrameworkServer fwserver.Server

	contextCancels   []context.CancelFunc
	contextCancelsMu sync.Mutex
}

func (s *Server) registerContext(in context.Context) context.Context {
	ctx, cancel := context.WithCancel(in)
	s.contextCancelsMu.Lock()
	defer s.contextCancelsMu.Unlock()
	s.contextCancels = append(s.contextCancels, cancel)
	return ctx
}

func (s *Server) cancelRegisteredContexts(_ context.Context) {
	s.contextCancelsMu.Lock()
	defer s.contextCancelsMu.Unlock()
	for _, cancel := range s.contextCancels {
		cancel()
	}
	s.contextCancels = nil
}

func (s *Server) GetProviderSchema(ctx context.Context, proto6Req *tfprotov6.GetProviderSchemaRequest) (*tfprotov6.GetProviderSchemaResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwReq := fromproto6.GetProviderSchemaRequest(ctx, proto6Req)
	fwResp := &fwserver.GetProviderSchemaResponse{}

	s.FrameworkServer.GetProviderSchema(ctx, fwReq, fwResp)

	return toproto6.GetProviderSchemaResponse(ctx, fwResp), nil
}

func (s *Server) ValidateProviderConfig(ctx context.Context, proto6Req *tfprotov6.ValidateProviderConfigRequest) (*tfprotov6.ValidateProviderConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ValidateProviderConfigResponse{}

	providerSchema, diags := s.FrameworkServer.ProviderSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateProviderConfigResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ValidateProviderConfigRequest(ctx, proto6Req, providerSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateProviderConfigResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ValidateProviderConfig(ctx, fwReq, fwResp)

	return toproto6.ValidateProviderConfigResponse(ctx, fwResp), nil
}

func (s *Server) ConfigureProvider(ctx context.Context, proto6Req *tfprotov6.ConfigureProviderRequest) (*tfprotov6.ConfigureProviderResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &tfsdk.ConfigureProviderResponse{}

	providerSchema, diags := s.FrameworkServer.ProviderSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ConfigureProviderResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ConfigureProviderRequest(ctx, proto6Req, providerSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ConfigureProviderResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ConfigureProvider(ctx, fwReq, fwResp)

	return toproto6.ConfigureProviderResponse(ctx, fwResp), nil
}

func (s *Server) StopProvider(ctx context.Context, _ *tfprotov6.StopProviderRequest) (*tfprotov6.StopProviderResponse, error) {
	s.cancelRegisteredContexts(ctx)

	return &tfprotov6.StopProviderResponse{}, nil
}

func (s *Server) ValidateResourceConfig(ctx context.Context, proto6Req *tfprotov6.ValidateResourceConfigRequest) (*tfprotov6.ValidateResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ValidateResourceConfigResponse{}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ValidateResourceConfigRequest(ctx, proto6Req, resourceType, resourceSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ValidateResourceConfig(ctx, fwReq, fwResp)

	return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
}

func (s *Server) UpgradeResourceState(ctx context.Context, proto6Req *tfprotov6.UpgradeResourceStateRequest) (*tfprotov6.UpgradeResourceStateResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.UpgradeResourceStateResponse{}

	if proto6Req == nil {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.UpgradeResourceStateRequest(ctx, proto6Req, resourceType, resourceSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.UpgradeResourceState(ctx, fwReq, fwResp)

	return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
}

func (s *Server) ReadResource(ctx context.Context, proto6Req *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ReadResourceResponse{}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadResourceResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadResourceResponse(ctx, fwResp), nil
	}

	providerMetaSchema, diags := s.FrameworkServer.ProviderMetaSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadResourceResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ReadResourceRequest(ctx, proto6Req, resourceType, resourceSchema, providerMetaSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadResourceResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ReadResource(ctx, fwReq, fwResp)

	return toproto6.ReadResourceResponse(ctx, fwResp), nil
}

func (s *Server) PlanResourceChange(ctx context.Context, proto6Req *tfprotov6.PlanResourceChangeRequest) (*tfprotov6.PlanResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.PlanResourceChangeResponse{}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.PlanResourceChangeResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.PlanResourceChangeResponse(ctx, fwResp), nil
	}

	providerMetaSchema, diags := s.FrameworkServer.ProviderMetaSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.PlanResourceChangeResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.PlanResourceChangeRequest(ctx, proto6Req, resourceType, resourceSchema, providerMetaSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.PlanResourceChangeResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.PlanResourceChange(ctx, fwReq, fwResp)

	return toproto6.PlanResourceChangeResponse(ctx, fwResp), nil
}

func (s *Server) ApplyResourceChange(ctx context.Context, proto6Req *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ApplyResourceChangeResponse{}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
	}

	providerMetaSchema, diags := s.FrameworkServer.ProviderMetaSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ApplyResourceChangeRequest(ctx, proto6Req, resourceType, resourceSchema, providerMetaSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ApplyResourceChange(ctx, fwReq, fwResp)

	return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
}

func (s *Server) ValidateDataResourceConfig(ctx context.Context, proto6Req *tfprotov6.ValidateDataResourceConfigRequest) (*tfprotov6.ValidateDataResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ValidateDataSourceConfigResponse{}

	dataSourceType, diags := s.FrameworkServer.DataSourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateDataSourceConfigResponse(ctx, fwResp), nil
	}

	dataSourceSchema, diags := s.FrameworkServer.DataSourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateDataSourceConfigResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ValidateDataSourceConfigRequest(ctx, proto6Req, dataSourceType, dataSourceSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateDataSourceConfigResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ValidateDataSourceConfig(ctx, fwReq, fwResp)

	return toproto6.ValidateDataSourceConfigResponse(ctx, fwResp), nil
}

func (s *Server) ReadDataSource(ctx context.Context, proto6Req *tfprotov6.ReadDataSourceRequest) (*tfprotov6.ReadDataSourceResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ReadDataSourceResponse{}

	dataSourceType, diags := s.FrameworkServer.DataSourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
	}

	dataSourceSchema, diags := s.FrameworkServer.DataSourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
	}

	providerMetaSchema, diags := s.FrameworkServer.ProviderMetaSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ReadDataSourceRequest(ctx, proto6Req, dataSourceType, dataSourceSchema, providerMetaSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ReadDataSource(ctx, fwReq, fwResp)

	return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
}
