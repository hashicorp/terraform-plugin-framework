package tfsdk

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/internal/proto6"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tf6server "github.com/hashicorp/terraform-plugin-go/tfprotov6/server"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tfprotov6.ProviderServer = &server{}

type server struct {
	p                Provider
	contextCancels   []context.CancelFunc
	contextCancelsMu sync.Mutex
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

func diagsHasErrors(in []*tfprotov6.Diagnostic) bool {
	for _, diag := range in {
		if diag == nil {
			continue
		}
		if diag.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func (s *server) registerContext(in context.Context) context.Context {
	ctx, cancel := context.WithCancel(in)
	s.contextCancelsMu.Lock()
	defer s.contextCancelsMu.Unlock()
	s.contextCancels = append(s.contextCancels, cancel)
	return ctx
}

func (s *server) getResourceType(ctx context.Context, typ string) (ResourceType, []*tfprotov6.Diagnostic) {
	resourceTypes, diags := s.p.GetResources(ctx)
	if diagsHasErrors(diags) {
		return nil, diags
	}
	resourceType, ok := resourceTypes[typ]
	if !ok {
		return nil, append(diags, &tfprotov6.Diagnostic{
			Summary: "Resource not found",
			Detail:  fmt.Sprintf("No resource named %q is configured on the provider", typ),
		})
	}
	return resourceType, nil
}

func (s *server) getDataSourceType(ctx context.Context, typ string) (DataSourceType, []*tfprotov6.Diagnostic) {
	dataSourceTypes, diags := s.p.GetDataSources(ctx)
	if diagsHasErrors(diags) {
		return nil, diags
	}
	dataSourceType, ok := dataSourceTypes[typ]
	if !ok {
		return nil, append(diags, &tfprotov6.Diagnostic{
			Summary: "Data source not found",
			Detail:  fmt.Sprintf("No data source named %q is configured on the provider", typ),
		})
	}
	return dataSourceType, nil
}

func (s *server) GetProviderSchema(ctx context.Context, _ *tfprotov6.GetProviderSchemaRequest) (*tfprotov6.GetProviderSchemaResponse, error) {
	ctx = s.registerContext(ctx)

	resp := new(tfprotov6.GetProviderSchemaResponse)

	// get the provider schema
	providerSchema, diags := s.p.GetSchema(ctx)
	if diags != nil {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		if diagsHasErrors(resp.Diagnostics) {
			return resp, nil
		}
	}
	// convert the provider schema to a *tfprotov6.Schema
	provider6Schema, err := proto6.Schema(ctx, providerSchema)
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
		if diags != nil {
			resp.Diagnostics = append(resp.Diagnostics, diags...)
			if diagsHasErrors(resp.Diagnostics) {
				return resp, nil
			}
		}
		pm6Schema, err := proto6.Schema(ctx, providerMetaSchema)
		if err != nil {
			// TODO: convert to diag
			return resp, nil
		}
		providerMeta6Schema = pm6Schema
	}

	// get our resource schemas
	resourceSchemas, diags := s.p.GetResources(ctx)
	if diags != nil {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		if diagsHasErrors(resp.Diagnostics) {
			return resp, nil
		}
	}
	resource6Schemas := map[string]*tfprotov6.Schema{}
	for k, v := range resourceSchemas {
		schema, diags := v.GetSchema(ctx)
		if diags != nil {
			resp.Diagnostics = append(resp.Diagnostics, diags...)
			if diagsHasErrors(resp.Diagnostics) {
				return resp, nil
			}
		}
		schema6, err := proto6.Schema(ctx, schema)
		if err != nil {
			// TODO: convert to diag
			return resp, nil
		}
		resource6Schemas[k] = schema6
	}

	// get our data source schemas
	dataSourceSchemas, diags := s.p.GetDataSources(ctx)
	if diags != nil {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		if diagsHasErrors(resp.Diagnostics) {
			return resp, nil
		}
	}
	dataSource6Schemas := map[string]*tfprotov6.Schema{}
	for k, v := range dataSourceSchemas {
		schema, diags := v.GetSchema(ctx)
		if diags != nil {
			resp.Diagnostics = append(resp.Diagnostics, diags...)
			if diagsHasErrors(resp.Diagnostics) {
				return resp, nil
			}
		}
		schema6, err := proto6.Schema(ctx, schema)
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
	ctx = s.registerContext(ctx)

	// We don't actually do anything as part of this. In theory, we could
	// validate the configuration for the provider block? Need to check in
	// again with the core team about the goal of this RPC.
	return &tfprotov6.ValidateProviderConfigResponse{
		PreparedConfig: req.Config,
	}, nil
}

func (s *server) ConfigureProvider(ctx context.Context, req *tfprotov6.ConfigureProviderRequest) (*tfprotov6.ConfigureProviderResponse, error) {
	ctx = s.registerContext(ctx)

	resp := &tfprotov6.ConfigureProviderResponse{}
	schema, diags := s.p.GetSchema(ctx)
	if diags != nil {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		if diagsHasErrors(resp.Diagnostics) {
			return resp, nil
		}
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
	s.contextCancelsMu.Lock()
	defer s.contextCancelsMu.Unlock()
	for _, cancel := range s.contextCancels {
		cancel()
	}
	s.contextCancels = nil
	return &tfprotov6.StopProviderResponse{}, nil
}

func (s *server) ValidateResourceConfig(ctx context.Context, _ *tfprotov6.ValidateResourceConfigRequest) (*tfprotov6.ValidateResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)

	// TODO: support validation
	return &tfprotov6.ValidateResourceConfigResponse{}, nil
}

func (s *server) UpgradeResourceState(ctx context.Context, _ *tfprotov6.UpgradeResourceStateRequest) (*tfprotov6.UpgradeResourceStateResponse, error) {
	ctx = s.registerContext(ctx)

	// TODO: support state upgrades
	return &tfprotov6.UpgradeResourceStateResponse{}, nil
}

func (s *server) ReadResource(ctx context.Context, req *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &tfprotov6.ReadResourceResponse{}

	resourceType, diags := s.getResourceType(ctx, req.TypeName)
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}
	/*
		TODO: eventually we'll have a State on the ReadResourceRequest type, and we'll want to fill in the schema on that. Until then, this is unused, so let's comment it out.
		resourceSchema, diags := resourceType.GetSchema(ctx)
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		if diagsHasErrors(resp.Diagnostics) {
			return resp, nil
		}
	*/
	resource, diags := resourceType.NewResource(s.p)
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}
	readReq := ReadResourceRequest{
		// TODO: when we get a state type, populate it
	}
	readResp := ReadResourceResponse{
		Diagnostics: resp.Diagnostics,
	}
	resource.Read(ctx, &readReq, &readResp)
	resp.Diagnostics = readResp.Diagnostics
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}
	// TODO: at some point we're going to need to handle the State that's returned. Not today, though!
	return resp, nil
}

func (s *server) PlanResourceChange(ctx context.Context, req *tfprotov6.PlanResourceChangeRequest) (*tfprotov6.PlanResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &tfprotov6.PlanResourceChangeResponse{}

	// get the type of resource, so we can get its schema and create an
	// instance
	resourceType, diags := s.getResourceType(ctx, req.TypeName)
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}

	// get the schema from the resource type, so we can embed it in the
	// config and plan
	resourceSchema, diags := resourceType.GetSchema(ctx)
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}

	plan, err := req.ProposedNewState.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		// TODO: return error
	}

	modifiedPlan, err := tftypes.Transform(plan, func(path *tftypes.AttributePath, val tftypes.Value) (tftypes.Value, error) {
		if !val.IsNull() {
			return val, nil
		}
		// TODO: if the Attribute at this path in the resource schema is not computed, return val instead of modifying it
		return tftypes.NewValue(val.Type(), tftypes.UnknownValue), nil
	})

	plannedState, err := tfprotov6.NewDynamicValue(modifiedPlan.Type(), modifiedPlan)
	if err != nil {
		// TODO: return error
	}
	resp.PlannedState = &plannedState

	// TODO: implement customizable plan modifications later
	// TODO: implement RequiresReplace behavior later
	return resp, nil
}

func (s *server) ApplyResourceChange(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &tfprotov6.ApplyResourceChangeResponse{}

	// get the type of resource, so we can get its scheman and create an
	// instance
	resourceType, diags := s.getResourceType(ctx, req.TypeName)
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}

	// get the schema from the resource type, so we can embed it in the
	// config and plan
	resourceSchema, diags := resourceType.GetSchema(ctx)
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}

	// create the resource instance, so we can call its methods and handle
	// the request
	resource, diags := resourceType.NewResource(s.p)
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}

	config, err := req.Config.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		// TODO: return error
	}

	// figure out what kind of request we're serving
	create, err := proto6.IsCreate(ctx, req, resourceSchema.TerraformType(ctx))
	if err != nil {
		// TODO: return error
	}
	update, err := proto6.IsUpdate(ctx, req, resourceSchema.TerraformType(ctx))
	if err != nil {
		// TODO: return error
	}
	destroy, err := proto6.IsDestroy(ctx, req, resourceSchema.TerraformType(ctx))
	if err != nil {
		// TODO: return error
	}

	switch {
	case create && !update && !destroy:
		createReq := &CreateResourceRequest{
			Config: Config{
				Schema: resourceSchema,
				Raw:    config,
			},
		}
		createResp := &CreateResourceResponse{
			Diagnostics: resp.Diagnostics,
		}
		resource.Create(ctx, createReq, createResp)
		resp.Diagnostics = createResp.Diagnostics
		// TODO: set partial state before returning error
		if diagsHasErrors(resp.Diagnostics) {
			return resp, nil
		}
		// TODO: set state on resp
	case !create && update && !destroy:
		updateReq := &UpdateResourceRequest{}
		updateResp := &UpdateResourceResponse{
			Diagnostics: resp.Diagnostics,
		}
		resource.Update(ctx, updateReq, updateResp)
		resp.Diagnostics = updateResp.Diagnostics
		// TODO: set partial state before returning error
		if diagsHasErrors(resp.Diagnostics) {
			return resp, nil
		}
		// TODO: set state on resp
	case !create && !update && destroy:
		destroyReq := &DeleteResourceRequest{}
		destroyResp := &DeleteResourceResponse{
			Diagnostics: resp.Diagnostics,
		}
		resource.Delete(ctx, destroyReq, destroyResp)
		resp.Diagnostics = destroyResp.Diagnostics
		// TODO: set partial state before returning error
		if diagsHasErrors(resp.Diagnostics) {
			return resp, nil
		}
		// TODO: set state on resp
	default:
		// TODO: return error
	}
	return resp, nil
}

func (s *server) ImportResourceState(ctx context.Context, _ *tfprotov6.ImportResourceStateRequest) (*tfprotov6.ImportResourceStateResponse, error) {
	ctx = s.registerContext(ctx)

	// TODO: support resource importing
	return &tfprotov6.ImportResourceStateResponse{}, nil
}

func (s *server) ValidateDataResourceConfig(ctx context.Context, _ *tfprotov6.ValidateDataResourceConfigRequest) (*tfprotov6.ValidateDataResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)

	// TODO: support validation
	return &tfprotov6.ValidateDataResourceConfigResponse{}, nil
}

func (s *server) ReadDataSource(ctx context.Context, req *tfprotov6.ReadDataSourceRequest) (*tfprotov6.ReadDataSourceResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &tfprotov6.ReadDataSourceResponse{}

	dataSourceType, diags := s.getDataSourceType(ctx, req.TypeName)
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}
	/*
		TODO: eventually we'll have a State on the ReadDataSourceRequest type, and we'll want to fill in the schema on that. Until then, this is unused, so let's comment it out.
		dataSourceSchema, diags := dataSourceType.GetSchema(ctx)
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		if diagsHasErrors(resp.Diagnostics) {
			return resp, nil
		}
	*/
	dataSource, diags := dataSourceType.NewDataSource(s.p)
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}
	readReq := ReadResourceRequest{
		// TODO: when we get a state type, populate it
	}
	readResp := ReadResourceResponse{
		Diagnostics: resp.Diagnostics,
	}
	dataSource.Read(ctx, &readReq, &readResp)
	resp.Diagnostics = readResp.Diagnostics
	if diagsHasErrors(resp.Diagnostics) {
		return resp, nil
	}
	// TODO: at some point we're going to need to handle the State that's returned. Not today, though!
	return resp, nil
}
