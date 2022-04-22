package tfsdk

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/proto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tfprotov6.ProviderServer = &server{}

type server struct {
	p                Provider
	contextCancels   []context.CancelFunc
	contextCancelsMu sync.Mutex
}

// NewProtocol6Server returns a tfprotov6.ProviderServer implementation based
// on the passed Provider implementation.
//
// Deprecated: Use NewProtocol6ProviderServer instead. This will be removed
// in an upcoming minor version before v1.0.0.
func NewProtocol6Server(p Provider) tfprotov6.ProviderServer {
	return &server{
		p: p,
	}
}

// Returns a protocol version 6 ProviderServer implementation based on the
// given Provider and suitable for usage with the terraform-plugin-go
// tf6server.Serve() function and various terraform-plugin-mux functions.
func NewProtocol6ProviderServer(p Provider) func() tfprotov6.ProviderServer {
	return func() tfprotov6.ProviderServer {
		return &server{
			p: p,
		}
	}
}

// Returns a protocol version 6 ProviderServer implementation based on the
// given Provider and suitable for usage with the terraform-plugin-sdk
// acceptance testing helper/resource.TestCase.ProtoV6ProviderFactories.
//
// The error return is not currently used, but it may be in the future.
func NewProtocol6ProviderServerWithError(p Provider) func() (tfprotov6.ProviderServer, error) {
	return func() (tfprotov6.ProviderServer, error) {
		return &server{
			p: p,
		}, nil
	}
}

// Serve serves a provider, blocking until the context is canceled.
func Serve(ctx context.Context, providerFunc func() Provider, opts ServeOpts) error {
	err := opts.validate(ctx)

	if err != nil {
		return fmt.Errorf("unable to validate ServeOpts: %w", err)
	}

	var tf6serverOpts []tf6server.ServeOpt

	if opts.Debug {
		tf6serverOpts = append(tf6serverOpts, tf6server.WithManagedDebug())
	}

	return tf6server.Serve(opts.address(ctx), func() tfprotov6.ProviderServer {
		return &server{
			p: providerFunc(),
		}
	}, tf6serverOpts...)
}

func (s *server) registerContext(in context.Context) context.Context {
	ctx, cancel := context.WithCancel(in)
	s.contextCancelsMu.Lock()
	defer s.contextCancelsMu.Unlock()
	s.contextCancels = append(s.contextCancels, cancel)
	return ctx
}

func (s *server) cancelRegisteredContexts(_ context.Context) {
	s.contextCancelsMu.Lock()
	defer s.contextCancelsMu.Unlock()
	for _, cancel := range s.contextCancels {
		cancel()
	}
	s.contextCancels = nil
}

func (s *server) getResourceType(ctx context.Context, typ string) (ResourceType, diag.Diagnostics) {
	// TODO: Cache GetResources call in GetProviderSchema and reference cache instead
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/299
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetResources")
	resourceTypes, diags := s.p.GetResources(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetResources")
	if diags.HasError() {
		return nil, diags
	}
	resourceType, ok := resourceTypes[typ]
	if !ok {
		diags.AddError(
			"Resource not found",
			fmt.Sprintf("No resource named %q is configured on the provider", typ),
		)
		return nil, diags
	}
	return resourceType, diags
}

func (s *server) getDataSourceType(ctx context.Context, typ string) (DataSourceType, diag.Diagnostics) {
	// TODO: Cache GetDataSources call in GetProviderSchema and reference cache instead
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/299
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetDataSources")
	dataSourceTypes, diags := s.p.GetDataSources(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetDataSources")
	if diags.HasError() {
		return nil, diags
	}
	dataSourceType, ok := dataSourceTypes[typ]
	if !ok {
		diags.AddError(
			"Data source not found",
			fmt.Sprintf("No data source named %q is configured on the provider", typ),
		)
		return nil, diags
	}
	return dataSourceType, diags
}

// getProviderSchemaResponse is a thin abstraction to allow native Diagnostics usage
type getProviderSchemaResponse struct {
	Provider          *tfprotov6.Schema
	ProviderMeta      *tfprotov6.Schema
	ResourceSchemas   map[string]*tfprotov6.Schema
	DataSourceSchemas map[string]*tfprotov6.Schema
	Diagnostics       diag.Diagnostics
}

func (r getProviderSchemaResponse) toTfprotov6() *tfprotov6.GetProviderSchemaResponse {
	return &tfprotov6.GetProviderSchemaResponse{
		Provider:          r.Provider,
		ProviderMeta:      r.ProviderMeta,
		ResourceSchemas:   r.ResourceSchemas,
		DataSourceSchemas: r.DataSourceSchemas,
		Diagnostics:       r.Diagnostics.ToTfprotov6Diagnostics(),
	}
}

func (s *server) GetProviderSchema(ctx context.Context, _ *tfprotov6.GetProviderSchemaRequest) (*tfprotov6.GetProviderSchemaResponse, error) {
	ctx = s.registerContext(ctx)
	resp := new(getProviderSchemaResponse)

	s.getProviderSchema(ctx, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) getProviderSchema(ctx context.Context, resp *getProviderSchemaResponse) {
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetSchema")
	providerSchema, diags := s.p.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetSchema")
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	// convert the provider schema to a *tfprotov6.Schema
	provider6Schema, err := providerSchema.tfprotov6Schema(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting provider schema",
			"The provider schema couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	// don't set the schema on the response yet, we want it to be able to
	// accrue warning diagnostics and return them on the first error
	// diagnostic without returning a partial schema, so we need to wait
	// until the very end to set the schemas on the response

	// if we have a provider_meta schema, get it
	var providerMeta6Schema *tfprotov6.Schema
	if pm, ok := s.p.(ProviderWithProviderMeta); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
		logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
		providerMetaSchema, diags := pm.GetMetaSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		pm6Schema, err := providerMetaSchema.tfprotov6Schema(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error converting provider_meta schema",
				"The provider_meta schema couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return
		}
		providerMeta6Schema = pm6Schema
	}

	// TODO: Cache GetDataSources call
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/299
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetResources")
	resourceSchemas, diags := s.p.GetResources(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetResources")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resource6Schemas := map[string]*tfprotov6.Schema{}
	for k, v := range resourceSchemas {
		// KeyResourceType field only necessary here since we are in GetProviderSchema RPC
		logging.FrameworkTrace(ctx, "Found resource type", map[string]interface{}{logging.KeyResourceType: k})
		logging.FrameworkDebug(ctx, "Calling provider defined ResourceType GetSchema", map[string]interface{}{logging.KeyResourceType: k})
		schema, diags := v.GetSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined ResourceType GetSchema", map[string]interface{}{logging.KeyResourceType: k})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		schema6, err := schema.tfprotov6Schema(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error converting resource schema",
				"The schema for the resource \""+k+"\" couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return
		}
		resource6Schemas[k] = schema6
	}

	// TODO: Cache GetDataSources call
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/299
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetDataSources")
	dataSourceSchemas, diags := s.p.GetDataSources(ctx)
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetDataSources")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dataSource6Schemas := map[string]*tfprotov6.Schema{}
	for k, v := range dataSourceSchemas {
		// KeyDataSourceType field only necessary here since we are in GetProviderSchema RPC
		logging.FrameworkTrace(ctx, "Found data source type", map[string]interface{}{logging.KeyDataSourceType: k})
		logging.FrameworkDebug(ctx, "Calling provider defined DataSourceType GetSchema", map[string]interface{}{logging.KeyDataSourceType: k})
		schema, diags := v.GetSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined DataSourceType GetSchema", map[string]interface{}{logging.KeyDataSourceType: k})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		schema6, err := schema.tfprotov6Schema(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error converting data sourceschema",
				"The schema for the data source \""+k+"\" couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return
		}
		dataSource6Schemas[k] = schema6
	}

	// ok, we didn't get any error diagnostics, populate the schemas and
	// send the response
	resp.Provider = provider6Schema
	resp.ProviderMeta = providerMeta6Schema
	resp.ResourceSchemas = resource6Schemas
	resp.DataSourceSchemas = dataSource6Schemas
}

// validateProviderConfigResponse is a thin abstraction to allow native Diagnostics usage
type validateProviderConfigResponse struct {
	PreparedConfig *tfprotov6.DynamicValue
	Diagnostics    diag.Diagnostics
}

func (r validateProviderConfigResponse) toTfprotov6() *tfprotov6.ValidateProviderConfigResponse {
	return &tfprotov6.ValidateProviderConfigResponse{
		PreparedConfig: r.PreparedConfig,
		Diagnostics:    r.Diagnostics.ToTfprotov6Diagnostics(),
	}
}

func (s *server) ValidateProviderConfig(ctx context.Context, req *tfprotov6.ValidateProviderConfigRequest) (*tfprotov6.ValidateProviderConfigResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &validateProviderConfigResponse{
		// This RPC allows a modified configuration to be returned. This was
		// previously used to allow a "required" provider attribute (as defined
		// by a schema) to still be "optional" with a default value, typically
		// through an environment variable. Other tooling based on the provider
		// schema information could not determine this implementation detail.
		// To ensure accuracy going forward, this implementation is opinionated
		// towards accurate provider schema definitions and optional values
		// can be filled in or return errors during ConfigureProvider().
		PreparedConfig: req.Config,
	}

	s.validateProviderConfig(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) validateProviderConfig(ctx context.Context, req *tfprotov6.ValidateProviderConfigRequest, resp *validateProviderConfigResponse) {
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetSchema")
	schema, diags := s.p.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetSchema")
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, err := req.Config.Unmarshal(schema.TerraformType(ctx))

	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing config",
			"The provider had a problem parsing the config. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}

	vpcReq := ValidateProviderConfigRequest{
		Config: Config{
			Raw:    config,
			Schema: schema,
		},
	}

	if provider, ok := s.p.(ProviderWithConfigValidators); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithConfigValidators")
		for _, configValidator := range provider.ConfigValidators(ctx) {
			vpcRes := &ValidateProviderConfigResponse{
				Diagnostics: resp.Diagnostics,
			}

			logging.FrameworkDebug(
				ctx,
				"Calling provider defined ProviderConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)
			configValidator.Validate(ctx, vpcReq, vpcRes)
			logging.FrameworkDebug(
				ctx,
				"Called provider defined ProviderConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)

			resp.Diagnostics = vpcRes.Diagnostics
		}
	}

	if provider, ok := s.p.(ProviderWithValidateConfig); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithValidateConfig")
		vpcRes := &ValidateProviderConfigResponse{
			Diagnostics: resp.Diagnostics,
		}

		logging.FrameworkDebug(ctx, "Calling provider defined Provider ValidateConfig")
		provider.ValidateConfig(ctx, vpcReq, vpcRes)
		logging.FrameworkDebug(ctx, "Called provider defined Provider ValidateConfig")

		resp.Diagnostics = vpcRes.Diagnostics
	}

	validateSchemaReq := ValidateSchemaRequest{
		Config: Config{
			Raw:    config,
			Schema: schema,
		},
	}
	validateSchemaResp := ValidateSchemaResponse{
		Diagnostics: resp.Diagnostics,
	}

	schema.validate(ctx, validateSchemaReq, &validateSchemaResp)

	resp.Diagnostics = validateSchemaResp.Diagnostics
}

// configureProviderResponse is a thin abstraction to allow native Diagnostics usage
type configureProviderResponse struct {
	Diagnostics diag.Diagnostics
}

func (r configureProviderResponse) toTfprotov6() *tfprotov6.ConfigureProviderResponse {
	return &tfprotov6.ConfigureProviderResponse{
		Diagnostics: r.Diagnostics.ToTfprotov6Diagnostics(),
	}
}

func (s *server) ConfigureProvider(ctx context.Context, req *tfprotov6.ConfigureProviderRequest) (*tfprotov6.ConfigureProviderResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &configureProviderResponse{}

	s.configureProvider(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) configureProvider(ctx context.Context, req *tfprotov6.ConfigureProviderRequest, resp *configureProviderResponse) {
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetSchema")
	schema, diags := s.p.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetSchema")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config, err := req.Config.Unmarshal(schema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing config",
			"The provider had a problem parsing the config. Report this to the provider developer:\n\n"+err.Error(),
		)
		return
	}
	r := ConfigureProviderRequest{
		TerraformVersion: req.TerraformVersion,
		Config: Config{
			Raw:    config,
			Schema: schema,
		},
	}
	res := &ConfigureProviderResponse{}
	logging.FrameworkDebug(ctx, "Calling provider defined Provider Configure")
	s.p.Configure(ctx, r, res)
	logging.FrameworkDebug(ctx, "Called provider defined Provider Configure")
	resp.Diagnostics.Append(res.Diagnostics...)
}

func (s *server) StopProvider(ctx context.Context, _ *tfprotov6.StopProviderRequest) (*tfprotov6.StopProviderResponse, error) {
	s.cancelRegisteredContexts(ctx)

	return &tfprotov6.StopProviderResponse{}, nil
}

// validateResourceConfigResponse is a thin abstraction to allow native Diagnostics usage
type validateResourceConfigResponse struct {
	Diagnostics diag.Diagnostics
}

func (r validateResourceConfigResponse) toTfprotov6() *tfprotov6.ValidateResourceConfigResponse {
	return &tfprotov6.ValidateResourceConfigResponse{
		Diagnostics: r.Diagnostics.ToTfprotov6Diagnostics(),
	}
}

func (s *server) ValidateResourceConfig(ctx context.Context, req *tfprotov6.ValidateResourceConfigRequest) (*tfprotov6.ValidateResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &validateResourceConfigResponse{}

	s.validateResourceConfig(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) validateResourceConfig(ctx context.Context, req *tfprotov6.ValidateResourceConfigRequest, resp *validateResourceConfigResponse) {
	// Get the type of resource, so we can get its schema and create an
	// instance
	resourceType, diags := s.getResourceType(ctx, req.TypeName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the schema from the resource type, so we can embed it in the
	// config
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType GetSchema")
	resourceSchema, diags := resourceType.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType GetSchema")
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the resource instance, so we can call its methods and handle
	// the request
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType NewResource")
	resource, diags := resourceType.NewResource(ctx, s.p)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, err := req.Config.Unmarshal(resourceSchema.TerraformType(ctx))

	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing config",
			"The provider had a problem parsing the config. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}

	vrcReq := ValidateResourceConfigRequest{
		Config: Config{
			Raw:    config,
			Schema: resourceSchema,
		},
	}

	if resource, ok := resource.(ResourceWithConfigValidators); ok {
		logging.FrameworkTrace(ctx, "Resource implements ResourceWithConfigValidators")
		for _, configValidator := range resource.ConfigValidators(ctx) {
			vrcRes := &ValidateResourceConfigResponse{
				Diagnostics: resp.Diagnostics,
			}

			logging.FrameworkDebug(
				ctx,
				"Calling provider defined ResourceConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)
			configValidator.Validate(ctx, vrcReq, vrcRes)
			logging.FrameworkDebug(
				ctx,
				"Called provider defined ResourceConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)

			resp.Diagnostics = vrcRes.Diagnostics
		}
	}

	if resource, ok := resource.(ResourceWithValidateConfig); ok {
		logging.FrameworkTrace(ctx, "Resource implements ResourceWithValidateConfig")
		vrcRes := &ValidateResourceConfigResponse{
			Diagnostics: resp.Diagnostics,
		}

		logging.FrameworkDebug(ctx, "Calling provider defined Resource ValidateConfig")
		resource.ValidateConfig(ctx, vrcReq, vrcRes)
		logging.FrameworkDebug(ctx, "Called provider defined Resource ValidateConfig")

		resp.Diagnostics = vrcRes.Diagnostics
	}

	validateSchemaReq := ValidateSchemaRequest{
		Config: Config{
			Raw:    config,
			Schema: resourceSchema,
		},
	}
	validateSchemaResp := ValidateSchemaResponse{
		Diagnostics: resp.Diagnostics,
	}

	resourceSchema.validate(ctx, validateSchemaReq, &validateSchemaResp)

	resp.Diagnostics = validateSchemaResp.Diagnostics
}

// upgradeResourceStateResponse is a thin abstraction to allow native
// Diagnostics usage.
type upgradeResourceStateResponse struct {
	Diagnostics   diag.Diagnostics
	UpgradedState *tfprotov6.DynamicValue
}

func (r upgradeResourceStateResponse) toTfprotov6() *tfprotov6.UpgradeResourceStateResponse {
	return &tfprotov6.UpgradeResourceStateResponse{
		Diagnostics:   r.Diagnostics.ToTfprotov6Diagnostics(),
		UpgradedState: r.UpgradedState,
	}
}

func (s *server) UpgradeResourceState(ctx context.Context, req *tfprotov6.UpgradeResourceStateRequest) (*tfprotov6.UpgradeResourceStateResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &upgradeResourceStateResponse{}

	s.upgradeResourceState(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) upgradeResourceState(ctx context.Context, req *tfprotov6.UpgradeResourceStateRequest, resp *upgradeResourceStateResponse) {
	if req == nil {
		return
	}

	resourceType, diags := s.getResourceType(ctx, req.TypeName)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// No UpgradedState to return. This could return an error diagnostic about
	// the odd scenario, but seems best to allow Terraform CLI to handle the
	// situation itself in case it might be expected behavior.
	if req.RawState == nil {
		return
	}

	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType GetSchema")
	resourceSchema, diags := resourceType.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType GetSchema")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Terraform CLI can call UpgradeResourceState even if the stored state
	// version matches the current schema. Presumably this is to account for
	// the previous terraform-plugin-sdk implementation, which handled some
	// state fixups on behalf of Terraform CLI. When this happens, we do not
	// want to return errors for a missing ResourceWithUpgradeState
	// implementation or an undefined version within an existing
	// ResourceWithUpgradeState implementation as that would be confusing
	// detail for provider developers. Instead, the framework will attempt to
	// roundtrip the prior RawState to a State matching the current Schema.
	//
	// TODO: To prevent provider developers from accidentially implementing
	// ResourceWithUpgradeState with a version matching the current schema
	// version which would never get called, the framework can introduce a
	// unit test helper.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/113
	if req.Version == resourceSchema.Version {
		logging.FrameworkTrace(ctx, "UpgradeResourceState request version matches current Schema version, using framework defined passthrough implementation")

		resourceSchemaType := resourceSchema.TerraformType(ctx)

		rawStateValue, err := req.RawState.Unmarshal(resourceSchemaType)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Previously Saved State for UpgradeResourceState",
				"There was an error reading the saved resource state using the current resource schema.\n\n"+
					"If this resource state was last refreshed with Terraform CLI 0.11 and earlier, it must be refreshed or applied with an older provider version first. "+
					"If you manually modified the resource state, you will need to manually modify it to match the current resource schema. "+
					"Otherwise, please report this to the provider developer:\n\n"+err.Error(),
			)
			return
		}

		// NewDynamicValue will ensure the Msgpack field is set for Terraform CLI
		// 0.12 through 0.14 compatibility when using terraform-plugin-mux tf6to5server.
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/262
		upgradedStateValue, err := tfprotov6.NewDynamicValue(resourceSchemaType, rawStateValue)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Convert Previously Saved State for UpgradeResourceState",
				"There was an error converting the saved resource state using the current resource schema. "+
					"This is always an issue in the Terraform Provider SDK used to implement the resource and should be reported to the provider developers.\n\n"+
					"Please report this to the provider developer:\n\n"+err.Error(),
			)
			return
		}

		resp.UpgradedState = &upgradedStateValue

		return
	}

	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType NewResource")
	resource, diags := resourceType.NewResource(ctx, s.p)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resourceWithUpgradeState, ok := resource.(ResourceWithUpgradeState)

	if !ok {
		resp.Diagnostics.AddError(
			"Unable to Upgrade Resource State",
			"This resource was implemented without an UpgradeState() method, "+
				fmt.Sprintf("however Terraform was expecting an implementation for version %d upgrade.\n\n", req.Version)+
				"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
		)
		return
	}

	logging.FrameworkTrace(ctx, "Resource implements ResourceWithUpgradeState")

	logging.FrameworkDebug(ctx, "Calling provider defined Resource UpgradeState")
	resourceStateUpgraders := resourceWithUpgradeState.UpgradeState(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Resource UpgradeState")

	// Panic prevention
	if resourceStateUpgraders == nil {
		resourceStateUpgraders = make(map[int64]ResourceStateUpgrader, 0)
	}

	resourceStateUpgrader, ok := resourceStateUpgraders[req.Version]

	if !ok {
		resp.Diagnostics.AddError(
			"Unable to Upgrade Resource State",
			"This resource was implemented with an UpgradeState() method, "+
				fmt.Sprintf("however Terraform was expecting an implementation for version %d upgrade.\n\n", req.Version)+
				"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
		)
		return
	}

	upgradeResourceStateRequest := UpgradeResourceStateRequest{
		RawState: req.RawState,
	}

	if resourceStateUpgrader.PriorSchema != nil {
		logging.FrameworkTrace(ctx, "Initializing populated UpgradeResourceStateRequest state from provider defined prior schema and request RawState")

		priorSchemaType := resourceStateUpgrader.PriorSchema.TerraformType(ctx)

		rawStateValue, err := req.RawState.Unmarshal(priorSchemaType)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Previously Saved State for UpgradeResourceState",
				fmt.Sprintf("There was an error reading the saved resource state using the prior resource schema defined for version %d upgrade.\n\n", req.Version)+
					"Please report this to the provider developer:\n\n"+err.Error(),
			)
			return
		}

		upgradeResourceStateRequest.State = &State{
			Raw:    rawStateValue,
			Schema: *resourceStateUpgrader.PriorSchema,
		}
	}

	upgradeResourceStateResponse := UpgradeResourceStateResponse{
		State: State{
			Schema: resourceSchema,
		},
	}

	// To simplify provider logic, this could perform a best effort attempt
	// to populate the response State by looping through all Attribute/Block
	// by calling the equivalent of SetAttribute(GetAttribute()) and skipping
	// any errors.

	logging.FrameworkDebug(ctx, "Calling provider defined StateUpgrader")
	resourceStateUpgrader.StateUpgrader(ctx, upgradeResourceStateRequest, &upgradeResourceStateResponse)
	logging.FrameworkDebug(ctx, "Called provider defined StateUpgrader")

	resp.Diagnostics.Append(upgradeResourceStateResponse.Diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	if upgradeResourceStateResponse.DynamicValue != nil {
		logging.FrameworkTrace(ctx, "UpgradeResourceStateResponse DynamicValue set, overriding State")
		resp.UpgradedState = upgradeResourceStateResponse.DynamicValue
		return
	}

	if upgradeResourceStateResponse.State.Raw.Type() == nil || upgradeResourceStateResponse.State.Raw.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Upgraded Resource State",
			fmt.Sprintf("After attempting a resource state upgrade to version %d, the provider did not return any state data. ", req.Version)+
				"Preventing the unexpected loss of resource state data. "+
				"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
		)
		return
	}

	upgradedStateValue, err := tfprotov6.NewDynamicValue(upgradeResourceStateResponse.State.Schema.TerraformType(ctx), upgradeResourceStateResponse.State.Raw)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Convert Upgraded Resource State",
			fmt.Sprintf("An unexpected error was encountered when converting the state returned for version %d upgrade to a usable type. ", req.Version)+
				"This is always an issue in the Terraform Provider SDK used to implement the resource and should be reported to the provider developers.\n\n"+
				"Please report this to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	resp.UpgradedState = &upgradedStateValue
}

// readResourceResponse is a thin abstraction to allow native Diagnostics usage
type readResourceResponse struct {
	NewState    *tfprotov6.DynamicValue
	Diagnostics diag.Diagnostics
	Private     []byte
}

func (r readResourceResponse) toTfprotov6() *tfprotov6.ReadResourceResponse {
	return &tfprotov6.ReadResourceResponse{
		NewState:    r.NewState,
		Diagnostics: r.Diagnostics.ToTfprotov6Diagnostics(),
		Private:     r.Private,
	}
}

func (s *server) ReadResource(ctx context.Context, req *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &readResourceResponse{}

	s.readResource(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) readResource(ctx context.Context, req *tfprotov6.ReadResourceRequest, resp *readResourceResponse) {
	resourceType, diags := s.getResourceType(ctx, req.TypeName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType GetSchema")
	resourceSchema, diags := resourceType.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType GetSchema")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType NewResource")
	resource, diags := resourceType.NewResource(ctx, s.p)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, err := req.CurrentState.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing current state",
			"There was an error parsing the current state. Please report this to the provider developer:\n\n"+err.Error(),
		)
		return
	}
	readReq := ReadResourceRequest{
		State: State{
			Raw:    state,
			Schema: resourceSchema,
		},
	}
	if pm, ok := s.p.(ProviderWithProviderMeta); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
		logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
		pmSchema, diags := pm.GetMetaSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		readReq.ProviderMeta = Config{
			Schema: pmSchema,
			Raw:    tftypes.NewValue(pmSchema.TerraformType(ctx), nil),
		}

		if req.ProviderMeta != nil {
			pmValue, err := req.ProviderMeta.Unmarshal(pmSchema.TerraformType(ctx))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error parsing provider_meta",
					"There was an error parsing the provider_meta block. Please report this to the provider developer:\n\n"+err.Error(),
				)
				return
			}
			readReq.ProviderMeta.Raw = pmValue
		}
	}
	readResp := ReadResourceResponse{
		State: State{
			Raw:    state,
			Schema: resourceSchema,
		},
		Diagnostics: resp.Diagnostics,
	}
	logging.FrameworkDebug(ctx, "Calling provider defined Resource Read")
	resource.Read(ctx, readReq, &readResp)
	logging.FrameworkDebug(ctx, "Called provider defined Resource Read")
	resp.Diagnostics = readResp.Diagnostics
	// don't return even if we have error diagnostics, we need to set the
	// state on the response, first

	newState, err := tfprotov6.NewDynamicValue(resourceSchema.TerraformType(ctx), readResp.State.Raw)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting read response",
			"An unexpected error was encountered when converting the read response to a usable type. This is always a problem with the provider. Please give the following information to the provider developer:\n\n"+err.Error(),
		)
		return
	}
	resp.NewState = &newState
}

func markComputedNilsAsUnknown(ctx context.Context, config tftypes.Value, resourceSchema Schema) func(*tftypes.AttributePath, tftypes.Value) (tftypes.Value, error) {
	return func(path *tftypes.AttributePath, val tftypes.Value) (tftypes.Value, error) {
		// we are only modifying attributes, not the entire resource
		if len(path.Steps()) < 1 {
			return val, nil
		}
		configVal, _, err := tftypes.WalkAttributePath(config, path)
		if err != tftypes.ErrInvalidStep && err != nil {
			logging.FrameworkError(ctx, "error walking attribute path", map[string]interface{}{logging.KeyAttributePath: path})
			return val, err
		} else if err != tftypes.ErrInvalidStep && !configVal.(tftypes.Value).IsNull() {
			logging.FrameworkTrace(ctx, "attribute not null in config, not marking unknown", map[string]interface{}{logging.KeyAttributePath: path})
			return val, nil
		}
		attribute, err := resourceSchema.AttributeAtPath(path)
		if err != nil {
			if errors.Is(err, ErrPathInsideAtomicAttribute) {
				// ignore attributes/elements inside schema.Attributes, they have no schema of their own
				logging.FrameworkTrace(ctx, "attribute is a non-schema attribute, not marking unknown", map[string]interface{}{logging.KeyAttributePath: path})
				return val, nil
			}
			logging.FrameworkError(ctx, "couldn't find attribute in resource schema", map[string]interface{}{logging.KeyAttributePath: path})
			return tftypes.Value{}, fmt.Errorf("couldn't find attribute in resource schema: %w", err)
		}
		if !attribute.Computed {
			logging.FrameworkTrace(ctx, "attribute is not computed in schema, not marking unknown", map[string]interface{}{logging.KeyAttributePath: path})
			return val, nil
		}
		logging.FrameworkDebug(ctx, "marking computed attribute that is null in the config as unknown", map[string]interface{}{logging.KeyAttributePath: path})
		return tftypes.NewValue(val.Type(), tftypes.UnknownValue), nil
	}
}

// planResourceChangeResponse is a thin abstraction to allow native Diagnostics usage
type planResourceChangeResponse struct {
	PlannedState    *tfprotov6.DynamicValue
	Diagnostics     diag.Diagnostics
	RequiresReplace []*tftypes.AttributePath
	PlannedPrivate  []byte
}

func (r planResourceChangeResponse) toTfprotov6() *tfprotov6.PlanResourceChangeResponse {
	return &tfprotov6.PlanResourceChangeResponse{
		PlannedState:    r.PlannedState,
		Diagnostics:     r.Diagnostics.ToTfprotov6Diagnostics(),
		RequiresReplace: r.RequiresReplace,
		PlannedPrivate:  r.PlannedPrivate,
	}
}

func (s *server) PlanResourceChange(ctx context.Context, req *tfprotov6.PlanResourceChangeRequest) (*tfprotov6.PlanResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &planResourceChangeResponse{}

	s.planResourceChange(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) planResourceChange(ctx context.Context, req *tfprotov6.PlanResourceChangeRequest, resp *planResourceChangeResponse) {
	// get the type of resource, so we can get its schema and create an
	// instance
	resourceType, diags := s.getResourceType(ctx, req.TypeName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get the schema from the resource type, so we can embed it in the
	// config and plan
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType GetSchema")
	resourceSchema, diags := resourceType.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType GetSchema")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := req.Config.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing configuration",
			"An unexpected error was encountered trying to parse the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	plan, err := req.ProposedNewState.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing plan",
			"There was an unexpected error parsing the plan. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	state, err := req.PriorState.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing prior state",
			"An unexpected error was encountered trying to parse the prior state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	resp.PlannedState = req.ProposedNewState

	// create the resource instance, so we can call its methods and handle
	// the request
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType NewResource")
	resource, diags := resourceType.NewResource(ctx, s.p)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Execute any AttributePlanModifiers.
	//
	// This pass is before any Computed-only attributes are marked as unknown
	// to ensure any plan changes will trigger that behavior. These plan
	// modifiers are run again after that marking to allow setting values
	// and preventing extraneous plan differences.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	//
	// TODO: Enabling this pass will generate the following test error:
	//
	//     --- FAIL: TestServerPlanResourceChange/two_modifyplan_add_list_elem (0.00s)
	// serve_test.go:3303: An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:
	//
	// ElementKeyInt(1).AttributeName("name") still remains in the path: step cannot be applied to this value
	//
	// To fix this, (Config).GetAttribute() should return nil instead of the error.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/183
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/150
	// See also: https://github.com/hashicorp/terraform-plugin-framework/pull/167

	// Execute any resource-level ModifyPlan method.
	//
	// This pass is before any Computed-only attributes are marked as unknown
	// to ensure any plan changes will trigger that behavior. These plan
	// modifiers be run again after that marking to allow setting values and
	// preventing extraneous plan differences.
	//
	// TODO: Enabling this pass will generate the following test error:
	//
	//     --- FAIL: TestServerPlanResourceChange/two_modifyplan_add_list_elem (0.00s)
	// serve_test.go:3303: An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:
	//
	// ElementKeyInt(1).AttributeName("name") still remains in the path: step cannot be applied to this value
	//
	// To fix this, (Config).GetAttribute() should return nil instead of the error.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/183
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/150
	// See also: https://github.com/hashicorp/terraform-plugin-framework/pull/167

	// After ensuring there are proposed changes, mark any computed attributes
	// that are null in the config as unknown in the plan, so providers have
	// the choice to update them.
	//
	// Later attribute and resource plan modifier passes can override the
	// unknown with a known value using any plan modifiers.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	if !plan.IsNull() && !plan.Equal(state) {
		logging.FrameworkTrace(ctx, "marking computed null values as unknown")
		modifiedPlan, err := tftypes.Transform(plan, markComputedNilsAsUnknown(ctx, config, resourceSchema))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error modifying plan",
				"There was an unexpected error updating the plan. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return
		}
		if !plan.Equal(modifiedPlan) {
			logging.FrameworkTrace(ctx, "at least one value was changed to unknown")
		}
		plan = modifiedPlan
	}

	// Execute any AttributePlanModifiers again. This allows overwriting
	// any unknown values.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	if !plan.IsNull() {
		modifySchemaPlanReq := ModifySchemaPlanRequest{
			Config: Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			State: State{
				Schema: resourceSchema,
				Raw:    state,
			},
			Plan: Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
		}
		if pm, ok := s.p.(ProviderWithProviderMeta); ok {
			logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
			logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
			pmSchema, diags := pm.GetMetaSchema(ctx)
			logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
			if diags != nil {
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
			modifySchemaPlanReq.ProviderMeta = Config{
				Schema: pmSchema,
				Raw:    tftypes.NewValue(pmSchema.TerraformType(ctx), nil),
			}

			if req.ProviderMeta != nil {
				pmValue, err := req.ProviderMeta.Unmarshal(pmSchema.TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"Error parsing provider_meta",
						"There was an error parsing the provider_meta block. Please report this to the provider developer:\n\n"+err.Error(),
					)
					return
				}
				modifySchemaPlanReq.ProviderMeta.Raw = pmValue
			}
		}

		modifySchemaPlanResp := ModifySchemaPlanResponse{
			Plan: Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
			Diagnostics: resp.Diagnostics,
		}

		resourceSchema.modifyPlan(ctx, modifySchemaPlanReq, &modifySchemaPlanResp)
		resp.RequiresReplace = append(resp.RequiresReplace, modifySchemaPlanResp.RequiresReplace...)
		plan = modifySchemaPlanResp.Plan.Raw
		resp.Diagnostics = modifySchemaPlanResp.Diagnostics
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Execute any resource-level ModifyPlan method again. This allows
	// overwriting any unknown values.
	//
	// We do this regardless of whether the plan is null or not, because we
	// want resources to be able to return diagnostics when planning to
	// delete resources, e.g. to inform practitioners that the resource
	// _can't_ be deleted in the API and will just be removed from
	// Terraform's state
	var modifyPlanResp ModifyResourcePlanResponse
	if resource, ok := resource.(ResourceWithModifyPlan); ok {
		logging.FrameworkTrace(ctx, "Resource implements ResourceWithModifyPlan")
		modifyPlanReq := ModifyResourcePlanRequest{
			Config: Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			State: State{
				Schema: resourceSchema,
				Raw:    state,
			},
			Plan: Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
		}
		if pm, ok := s.p.(ProviderWithProviderMeta); ok {
			logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
			logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
			pmSchema, diags := pm.GetMetaSchema(ctx)
			logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			modifyPlanReq.ProviderMeta = Config{
				Schema: pmSchema,
				Raw:    tftypes.NewValue(pmSchema.TerraformType(ctx), nil),
			}

			if req.ProviderMeta != nil {
				pmValue, err := req.ProviderMeta.Unmarshal(pmSchema.TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"Error parsing provider_meta",
						"There was an error parsing the provider_meta block. Please report this to the provider developer:\n\n"+err.Error(),
					)
					return
				}
				modifyPlanReq.ProviderMeta.Raw = pmValue
			}
		}

		modifyPlanResp = ModifyResourcePlanResponse{
			Plan: Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
			RequiresReplace: []*tftypes.AttributePath{},
			Diagnostics:     resp.Diagnostics,
		}
		logging.FrameworkDebug(ctx, "Calling provider defined Resource ModifyPlan")
		resource.ModifyPlan(ctx, modifyPlanReq, &modifyPlanResp)
		logging.FrameworkDebug(ctx, "Called provider defined Resource ModifyPlan")
		resp.Diagnostics = modifyPlanResp.Diagnostics
		plan = modifyPlanResp.Plan.Raw
	}

	plannedState, err := tfprotov6.NewDynamicValue(plan.Type(), plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting response",
			"There was an unexpected error converting the state in the response to a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}
	resp.PlannedState = &plannedState
	resp.RequiresReplace = append(resp.RequiresReplace, modifyPlanResp.RequiresReplace...)

	// ensure deterministic RequiresReplace by sorting and deduplicating
	resp.RequiresReplace = normaliseRequiresReplace(ctx, resp.RequiresReplace)
}

// applyResourceChangeResponse is a thin abstraction to allow native Diagnostics usage
type applyResourceChangeResponse struct {
	NewState    *tfprotov6.DynamicValue
	Private     []byte
	Diagnostics diag.Diagnostics
}

func (r applyResourceChangeResponse) toTfprotov6() *tfprotov6.ApplyResourceChangeResponse {
	return &tfprotov6.ApplyResourceChangeResponse{
		NewState:    r.NewState,
		Private:     r.Private,
		Diagnostics: r.Diagnostics.ToTfprotov6Diagnostics(),
	}
}

// normaliseRequiresReplace sorts and deduplicates the slice of AttributePaths
// used in the RequiresReplace response field.
// Sorting is lexical based on the string representation of each AttributePath.
func normaliseRequiresReplace(ctx context.Context, rs []*tftypes.AttributePath) []*tftypes.AttributePath {
	if len(rs) < 2 {
		return rs
	}

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].String() < rs[j].String()
	})

	ret := make([]*tftypes.AttributePath, len(rs))
	ret[0] = rs[0]

	// deduplicate
	j := 1
	for i := 1; i < len(rs); i++ {
		if rs[i].Equal(ret[j-1]) {
			logging.FrameworkDebug(ctx, "attribute found multiple times in RequiresReplace, removing duplicate", map[string]interface{}{logging.KeyAttributePath: rs[i]})
			continue
		}
		ret[j] = rs[i]
		j++
	}
	return ret[:j]
}

func (s *server) ApplyResourceChange(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &applyResourceChangeResponse{
		// default to the prior state, so the state won't change unless
		// we choose to change it
		NewState: req.PriorState,
	}

	s.applyResourceChange(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) applyResourceChange(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest, resp *applyResourceChangeResponse) {
	// get the type of resource, so we can get its schema and create an
	// instance
	resourceType, diags := s.getResourceType(ctx, req.TypeName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get the schema from the resource type, so we can embed it in the
	// config and plan
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType GetSchema")
	resourceSchema, diags := resourceType.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType GetSchema")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the resource instance, so we can call its methods and handle
	// the request
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType NewResource")
	resource, diags := resourceType.NewResource(ctx, s.p)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := req.Config.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing configuration",
			"An unexpected error was encountered trying to parse the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	plan, err := req.PlannedState.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing plan",
			"An unexpected error was encountered trying to parse the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	priorState, err := req.PriorState.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing prior state",
			"An unexpected error was encountered trying to parse the prior state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	// figure out what kind of request we're serving
	create, err := proto6.IsCreate(ctx, req, resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error understanding request",
			"An unexpected error was encountered trying to understand the type of request being made. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}
	update, err := proto6.IsUpdate(ctx, req, resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error understanding request",
			"An unexpected error was encountered trying to understand the type of request being made. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}
	destroy, err := proto6.IsDestroy(ctx, req, resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error understanding request",
			"An unexpected error was encountered trying to understand the type of request being made. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	switch {
	case create && !update && !destroy:
		logging.FrameworkTrace(ctx, "running create")
		createReq := CreateResourceRequest{
			Config: Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			Plan: Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
		}
		if pm, ok := s.p.(ProviderWithProviderMeta); ok {
			logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
			logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
			pmSchema, diags := pm.GetMetaSchema(ctx)
			logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			createReq.ProviderMeta = Config{
				Schema: pmSchema,
				Raw:    tftypes.NewValue(pmSchema.TerraformType(ctx), nil),
			}

			if req.ProviderMeta != nil {
				pmValue, err := req.ProviderMeta.Unmarshal(pmSchema.TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"Error parsing provider_meta",
						"There was an error parsing the provider_meta block. Please report this to the provider developer:\n\n"+err.Error(),
					)
					return
				}
				createReq.ProviderMeta.Raw = pmValue
			}
		}
		createResp := CreateResourceResponse{
			State: State{
				Schema: resourceSchema,
				Raw:    priorState,
			},
			Diagnostics: resp.Diagnostics,
		}
		logging.FrameworkDebug(ctx, "Calling provider defined Resource Create")
		resource.Create(ctx, createReq, &createResp)
		logging.FrameworkDebug(ctx, "Called provider defined Resource Create")
		resp.Diagnostics = createResp.Diagnostics
		newState, err := tfprotov6.NewDynamicValue(resourceSchema.TerraformType(ctx), createResp.State.Raw)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error converting create response",
				"An unexpected error was encountered when converting the create response to a usable type. This is always a problem with the provider. Please give the following information to the provider developer:\n\n"+err.Error(),
			)
			return
		}
		resp.NewState = &newState
	case !create && update && !destroy:
		logging.FrameworkTrace(ctx, "running update")
		updateReq := UpdateResourceRequest{
			Config: Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			Plan: Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
			State: State{
				Schema: resourceSchema,
				Raw:    priorState,
			},
		}
		if pm, ok := s.p.(ProviderWithProviderMeta); ok {
			logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
			logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
			pmSchema, diags := pm.GetMetaSchema(ctx)
			logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			updateReq.ProviderMeta = Config{
				Schema: pmSchema,
				Raw:    tftypes.NewValue(pmSchema.TerraformType(ctx), nil),
			}

			if req.ProviderMeta != nil {
				pmValue, err := req.ProviderMeta.Unmarshal(pmSchema.TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"Error parsing provider_meta",
						"There was an error parsing the provider_meta block. Please report this to the provider developer:\n\n"+err.Error(),
					)
					return
				}
				updateReq.ProviderMeta.Raw = pmValue
			}
		}
		updateResp := UpdateResourceResponse{
			State: State{
				Schema: resourceSchema,
				Raw:    priorState,
			},
			Diagnostics: resp.Diagnostics,
		}
		logging.FrameworkDebug(ctx, "Calling provider defined Resource Update")
		resource.Update(ctx, updateReq, &updateResp)
		logging.FrameworkDebug(ctx, "Called provider defined Resource Update")
		resp.Diagnostics = updateResp.Diagnostics
		newState, err := tfprotov6.NewDynamicValue(resourceSchema.TerraformType(ctx), updateResp.State.Raw)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error converting update response",
				"An unexpected error was encountered when converting the update response to a usable type. This is always a problem with the provider. Please give the following information to the provider developer:\n\n"+err.Error(),
			)
			return
		}
		resp.NewState = &newState
	case !create && !update && destroy:
		logging.FrameworkTrace(ctx, "running delete")
		destroyReq := DeleteResourceRequest{
			State: State{
				Schema: resourceSchema,
				Raw:    priorState,
			},
		}
		if pm, ok := s.p.(ProviderWithProviderMeta); ok {
			logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
			logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
			pmSchema, diags := pm.GetMetaSchema(ctx)
			logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			destroyReq.ProviderMeta = Config{
				Schema: pmSchema,
				Raw:    tftypes.NewValue(pmSchema.TerraformType(ctx), nil),
			}

			if req.ProviderMeta != nil {
				pmValue, err := req.ProviderMeta.Unmarshal(pmSchema.TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"Error parsing provider_meta",
						"There was an error parsing the provider_meta block. Please report this to the provider developer:\n\n"+err.Error(),
					)
					return
				}
				destroyReq.ProviderMeta.Raw = pmValue
			}
		}
		destroyResp := DeleteResourceResponse{
			State: State{
				Schema: resourceSchema,
				Raw:    priorState,
			},
			Diagnostics: resp.Diagnostics,
		}
		logging.FrameworkDebug(ctx, "Calling provider defined Resource Delete")
		resource.Delete(ctx, destroyReq, &destroyResp)
		logging.FrameworkDebug(ctx, "Called provider defined Resource Delete")
		resp.Diagnostics = destroyResp.Diagnostics

		if !resp.Diagnostics.HasError() {
			logging.FrameworkTrace(ctx, "No provider defined Delete errors detected, ensuring State is cleared")
			destroyResp.State.RemoveResource(ctx)
		}

		newState, err := tfprotov6.NewDynamicValue(resourceSchema.TerraformType(ctx), destroyResp.State.Raw)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error converting delete response",
				"An unexpected error was encountered when converting the delete response to a usable type. This is always a problem with the provider. Please give the following information to the provider developer:\n\n"+err.Error(),
			)
			return
		}
		resp.NewState = &newState
	default:
		resp.Diagnostics.AddError(
			"Error understanding request",
			fmt.Sprintf("An unexpected error was encountered trying to understand the type of request being made. This is always an error in the provider. Please report the following to the provider developer:\n\nRequest matched unexpected number of methods: (create: %v, update: %v, delete: %v)", create, update, destroy),
		)
	}
}

// validateDataResourceConfigResponse is a thin abstraction to allow native Diagnostics usage
type validateDataResourceConfigResponse struct {
	Diagnostics diag.Diagnostics
}

func (r validateDataResourceConfigResponse) toTfprotov6() *tfprotov6.ValidateDataResourceConfigResponse {
	return &tfprotov6.ValidateDataResourceConfigResponse{
		Diagnostics: r.Diagnostics.ToTfprotov6Diagnostics(),
	}
}

func (s *server) ValidateDataResourceConfig(ctx context.Context, req *tfprotov6.ValidateDataResourceConfigRequest) (*tfprotov6.ValidateDataResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &validateDataResourceConfigResponse{}

	s.validateDataResourceConfig(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) validateDataResourceConfig(ctx context.Context, req *tfprotov6.ValidateDataResourceConfigRequest, resp *validateDataResourceConfigResponse) {

	// Get the type of data source, so we can get its schema and create an
	// instance
	dataSourceType, diags := s.getDataSourceType(ctx, req.TypeName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the schema from the data source type, so we can embed it in the
	// config
	logging.FrameworkDebug(ctx, "Calling provider defined DataSourceType GetSchema")
	dataSourceSchema, diags := dataSourceType.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined DataSourceType GetSchema")
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the data source instance, so we can call its methods and handle
	// the request
	logging.FrameworkDebug(ctx, "Calling provider defined DataSourceType NewDataSource")
	dataSource, diags := dataSourceType.NewDataSource(ctx, s.p)
	logging.FrameworkDebug(ctx, "Called provider defined DataSourceType NewDataSource")
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, err := req.Config.Unmarshal(dataSourceSchema.TerraformType(ctx))

	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing config",
			"The provider had a problem parsing the config. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}

	vrcReq := ValidateDataSourceConfigRequest{
		Config: Config{
			Raw:    config,
			Schema: dataSourceSchema,
		},
	}

	if dataSource, ok := dataSource.(DataSourceWithConfigValidators); ok {
		logging.FrameworkTrace(ctx, "DataSource implements DataSourceWithConfigValidators")
		for _, configValidator := range dataSource.ConfigValidators(ctx) {
			vrcRes := &ValidateDataSourceConfigResponse{
				Diagnostics: resp.Diagnostics,
			}

			logging.FrameworkDebug(
				ctx,
				"Calling provider defined DataSourceConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)
			configValidator.Validate(ctx, vrcReq, vrcRes)
			logging.FrameworkDebug(
				ctx,
				"Called provider defined DataSourceConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)

			resp.Diagnostics = vrcRes.Diagnostics
		}
	}

	if dataSource, ok := dataSource.(DataSourceWithValidateConfig); ok {
		logging.FrameworkTrace(ctx, "DataSource implements DataSourceWithValidateConfig")
		vrcRes := &ValidateDataSourceConfigResponse{
			Diagnostics: resp.Diagnostics,
		}

		logging.FrameworkDebug(ctx, "Calling provider defined DataSource ValidateConfig")
		dataSource.ValidateConfig(ctx, vrcReq, vrcRes)
		logging.FrameworkDebug(ctx, "Called provider defined DataSource ValidateConfig")

		resp.Diagnostics = vrcRes.Diagnostics
	}

	validateSchemaReq := ValidateSchemaRequest{
		Config: Config{
			Raw:    config,
			Schema: dataSourceSchema,
		},
	}
	validateSchemaResp := ValidateSchemaResponse{
		Diagnostics: resp.Diagnostics,
	}

	dataSourceSchema.validate(ctx, validateSchemaReq, &validateSchemaResp)

	resp.Diagnostics = validateSchemaResp.Diagnostics
}

// readDataSourceResponse is a thin abstraction to allow native Diagnostics usage
type readDataSourceResponse struct {
	State       *tfprotov6.DynamicValue
	Diagnostics diag.Diagnostics
}

func (r readDataSourceResponse) toTfprotov6() *tfprotov6.ReadDataSourceResponse {
	return &tfprotov6.ReadDataSourceResponse{
		State:       r.State,
		Diagnostics: r.Diagnostics.ToTfprotov6Diagnostics(),
	}
}

func (s *server) ReadDataSource(ctx context.Context, req *tfprotov6.ReadDataSourceRequest) (*tfprotov6.ReadDataSourceResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &readDataSourceResponse{}

	s.readDataSource(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *server) readDataSource(ctx context.Context, req *tfprotov6.ReadDataSourceRequest, resp *readDataSourceResponse) {
	dataSourceType, diags := s.getDataSourceType(ctx, req.TypeName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	logging.FrameworkDebug(ctx, "Calling provider defined DataSourceType GetSchema")
	dataSourceSchema, diags := dataSourceType.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined DataSourceType GetSchema")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	logging.FrameworkDebug(ctx, "Calling provider defined DataSourceType NewDataSource")
	dataSource, diags := dataSourceType.NewDataSource(ctx, s.p)
	logging.FrameworkDebug(ctx, "Called provider defined DataSourceType NewDataSource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config, err := req.Config.Unmarshal(dataSourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing current state",
			"There was an error parsing the current state. Please report this to the provider developer:\n\n"+err.Error(),
		)
		return
	}
	readReq := ReadDataSourceRequest{
		Config: Config{
			Raw:    config,
			Schema: dataSourceSchema,
		},
	}
	if pm, ok := s.p.(ProviderWithProviderMeta); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
		logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
		pmSchema, diags := pm.GetMetaSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		readReq.ProviderMeta = Config{
			Schema: pmSchema,
			Raw:    tftypes.NewValue(pmSchema.TerraformType(ctx), nil),
		}

		if req.ProviderMeta != nil {
			pmValue, err := req.ProviderMeta.Unmarshal(pmSchema.TerraformType(ctx))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error parsing provider_meta",
					"There was an error parsing the provider_meta block. Please report this to the provider developer:\n\n"+err.Error(),
				)
				return
			}
			readReq.ProviderMeta.Raw = pmValue
		}
	}
	readResp := ReadDataSourceResponse{
		State: State{
			Schema: dataSourceSchema,
			// default to the config values
			// they should be of the same type
			// we just want SetAttribute to not find an empty value
			Raw: config,
		},
		Diagnostics: resp.Diagnostics,
	}
	logging.FrameworkDebug(ctx, "Calling provider defined DataSource Read")
	dataSource.Read(ctx, readReq, &readResp)
	logging.FrameworkDebug(ctx, "Called provider defined DataSource Read")
	resp.Diagnostics = readResp.Diagnostics
	// don't return even if we have error diagnostics, we need to set the
	// state on the response, first

	state, err := tfprotov6.NewDynamicValue(dataSourceSchema.TerraformType(ctx), readResp.State.Raw)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting read response",
			"An unexpected error was encountered when converting the read response to a usable type. This is always a problem with the provider. Please give the following information to the provider developer:\n\n"+err.Error(),
		)
		return
	}
	resp.State = &state
}
