package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tf6server "github.com/hashicorp/terraform-plugin-go/tfprotov6/server"
)

var _ tfprotov6.ProviderServer = &server{}

type server struct {
	p Provider
}

type ServeOpts struct {
	Name string
}

func Serve(ctx context.Context, factory func() Provider, opts ServeOpts) error {
	return tf6server.Serve(opts.Name, func() tfprotov6.ProviderServer {
		return &server{
			p: factory(),
		}
	}) // TODO: set up debug serving if the --debug flag is passed
}

func proto6Schema(ctx context.Context, s schema.Schema) (*tfprotov6.Schema, error) {
	// TODO: convert schema from our type to *tfprotov6.Schema
	return nil, nil
}

func (s *server) GetProviderSchema(ctx context.Context, _ *tfprotov6.GetProviderSchemaRequest) (*tfprotov6.GetProviderSchemaResponse, error) {
	resp := new(tfprotov6.GetProviderSchemaResponse)

	// get the provider schema
	providerSchema, diags := s.p.GetSchema(ctx)
	if diags != nil { // TODO: don't return if no errors
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return resp, nil
	}
	// convert the provider schema to a *tfprotov6.Schema
	provider6Schema, err := proto6Schema(ctx, providerSchema)
	if err != nil {
		// TODO: convert to diag
		return resp, nil
	}

	// don't set the schema on the response yet, we want it to be able to
	// accrue warning diagnostics and return them on the first error
	// diagnostic without returning a partial schema, so we need to wait
	// until the very end to set the schemas on the response

	// if we have a provider_meta schema, get it
	var providerMeta6Schema *tfprotov6.Schema
	if pm, ok := s.p.(ProviderWithProviderMeta); ok {
		providerMetaSchema, diags := pm.GetMetaSchema(ctx)
		if diags != nil { // TODO: don't return if no errors
			resp.Diagnostics = append(resp.Diagnostics, diags...)
			return resp, nil
		}
		pm6Schema, err := proto6Schema(ctx, providerMetaSchema)
		if err != nil {
			// TODO: convert to diag
			return resp, nil
		}
		providerMeta6Schema = pm6Schema
	}

	// get our resource schemas
	resourceSchemas, diags := s.p.GetResources(ctx)
	if diags != nil { // TODO: don't return if no errors
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return resp, nil
	}
	resource6Schemas := map[string]*tfprotov6.Schema{}
	for k, v := range resourceSchemas {
		schema, diags := v.GetSchema(ctx)
		if diags != nil { // TODO: don't return if no errors
			resp.Diagnostics = append(resp.Diagnostics, diags...)
			return resp, nil
		}
		schema6, err := proto6Schema(ctx, schema)
		if err != nil {
			// TODO: convert to diag
			return resp, nil
		}
		resource6Schemas[k] = schema6
	}

	// get our data source schemas
	dataSourceSchemas, diags := s.p.GetDataSources(ctx)
	if diags != nil { // TODO: don't return if no errors
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return resp, nil
	}
	dataSource6Schemas := map[string]*tfprotov6.Schema{}
	for k, v := range dataSourceSchemas {
		schema, diags := v.GetSchema(ctx)
		if diags != nil { // TODO: don't return if no errors
			resp.Diagnostics = append(resp.Diagnostics, diags...)
			return resp, nil
		}
		schema6, err := proto6Schema(ctx, schema)
		if err != nil {
			// TODO: convert to diag
			return resp, nil
		}
		dataSource6Schemas[k] = schema6
	}

	// ok, we didn't get any error diagnostics, populate the schemas and
	// send the response
	resp.Provider = provider6Schema
	resp.ProviderMeta = providerMeta6Schema
	resp.ResourceSchemas = resource6Schemas
	resp.DataSourceSchemas = dataSource6Schemas
	return resp, nil
}

func (s *server) ValidateProviderConfig(ctx context.Context, req *tfprotov6.ValidateProviderConfigRequest) (*tfprotov6.ValidateProviderConfigResponse, error) {
	return &tfprotov6.ValidateProviderConfigResponse{
		PreparedConfig: req.Config,
	}, nil
}

func (s *server) ConfigureProvider(ctx context.Context, req *tfprotov6.ConfigureProviderRequest) (*tfprotov6.ConfigureProviderResponse, error) {
	resp := &tfprotov6.ConfigureProviderResponse{}
	schema, diags := s.p.GetSchema(ctx)
	if diags != nil { // TODO: only return if error diags
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return resp, nil
	}
	config, err := req.Config.Unmarshal(schema.TerraformType(ctx))
	if err != nil {
		// TODO: convert to diagnostic
		return resp, nil
	}
	r := &ConfigureProviderRequest{
		TerraformVersion: req.TerraformVersion,
		Config: Config{
			Raw:    config,
			Schema: schema,
		},
	}
	res := &ConfigureProviderResponse{}
	s.p.Configure(ctx, r, res)
	resp.Diagnostics = append(resp.Diagnostics, res.Diagnostics...)
	return resp, nil
}

func (s *server) StopProvider(ctx context.Context, _ *tfprotov6.StopProviderRequest) (*tfprotov6.StopProviderResponse, error) {
	// TODO: cancel all contexts
	return &tfprotov6.StopProviderResponse{}, nil
}

func (s *server) ValidateResourceConfig(ctx context.Context, _ *tfprotov6.ValidateResourceConfigRequest) (*tfprotov6.ValidateResourceConfigResponse, error) {
	// TODO: support validation
	return &tfprotov6.ValidateResourceConfigResponse{}, nil
}

func (s *server) UpgradeResourceState(ctx context.Context, _ *tfprotov6.UpgradeResourceStateRequest) (*tfprotov6.UpgradeResourceStateResponse, error) {
	// TODO: support state upgrades
	return &tfprotov6.UpgradeResourceStateResponse{}, nil
}

func (s *server) ReadResource(ctx context.Context, _ *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (s *server) PlanResourceChange(ctx context.Context, _ *tfprotov6.PlanResourceChangeRequest) (*tfprotov6.PlanResourceChangeResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (s *server) ApplyResourceChange(ctx context.Context, _ *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (s *server) ImportResourceState(ctx context.Context, _ *tfprotov6.ImportResourceStateRequest) (*tfprotov6.ImportResourceStateResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (s *server) ValidateDataResourceConfig(ctx context.Context, _ *tfprotov6.ValidateDataResourceConfigRequest) (*tfprotov6.ValidateDataResourceConfigResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (s *server) ReadDataSource(ctx context.Context, _ *tfprotov6.ReadDataSourceRequest) (*tfprotov6.ReadDataSourceResponse, error) {
	panic("not implemented") // TODO: Implement
}
