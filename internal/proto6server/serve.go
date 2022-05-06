package proto6server

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/proto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tfprotov6.ProviderServer = &Server{}

// Provider server implementation.
type Server struct {
	FrameworkServer fwserver.Server

	// Provider will be migrated to FrameworkServer over time.
	Provider         tfsdk.Provider
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

func (s *Server) getResourceType(ctx context.Context, typ string) (tfsdk.ResourceType, diag.Diagnostics) {
	// TODO: Cache GetResources call in GetProviderSchema and reference cache instead
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/299
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetResources")
	resourceTypes, diags := s.Provider.GetResources(ctx)
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

func (s *Server) getDataSourceType(ctx context.Context, typ string) (tfsdk.DataSourceType, diag.Diagnostics) {
	// TODO: Cache GetDataSources call in GetProviderSchema and reference cache instead
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/299
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetDataSources")
	dataSourceTypes, diags := s.Provider.GetDataSources(ctx)
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

func (s *Server) GetProviderSchema(ctx context.Context, req *tfprotov6.GetProviderSchemaRequest) (*tfprotov6.GetProviderSchemaResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwReq := fromproto6.GetProviderSchemaRequest(ctx, req)
	fwResp := &fwserver.GetProviderSchemaResponse{}

	s.FrameworkServer.GetProviderSchema(ctx, fwReq, fwResp)

	return toproto6.GetProviderSchemaResponse(ctx, fwResp), nil
}

// validateProviderConfigResponse is a thin abstraction to allow native Diagnostics usage
type validateProviderConfigResponse struct {
	PreparedConfig *tfprotov6.DynamicValue
	Diagnostics    diag.Diagnostics
}

func (r validateProviderConfigResponse) toTfprotov6() *tfprotov6.ValidateProviderConfigResponse {
	return &tfprotov6.ValidateProviderConfigResponse{
		PreparedConfig: r.PreparedConfig,
		Diagnostics:    toproto6.Diagnostics(r.Diagnostics),
	}
}

func (s *Server) ValidateProviderConfig(ctx context.Context, req *tfprotov6.ValidateProviderConfigRequest) (*tfprotov6.ValidateProviderConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
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

func (s *Server) validateProviderConfig(ctx context.Context, req *tfprotov6.ValidateProviderConfigRequest, resp *validateProviderConfigResponse) {
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetSchema")
	schema, diags := s.Provider.GetSchema(ctx)
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

	vpcReq := tfsdk.ValidateProviderConfigRequest{
		Config: tfsdk.Config{
			Raw:    config,
			Schema: schema,
		},
	}

	if provider, ok := s.Provider.(tfsdk.ProviderWithConfigValidators); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithConfigValidators")
		for _, configValidator := range provider.ConfigValidators(ctx) {
			vpcRes := &tfsdk.ValidateProviderConfigResponse{
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

	if provider, ok := s.Provider.(tfsdk.ProviderWithValidateConfig); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithValidateConfig")
		vpcRes := &tfsdk.ValidateProviderConfigResponse{
			Diagnostics: resp.Diagnostics,
		}

		logging.FrameworkDebug(ctx, "Calling provider defined Provider ValidateConfig")
		provider.ValidateConfig(ctx, vpcReq, vpcRes)
		logging.FrameworkDebug(ctx, "Called provider defined Provider ValidateConfig")

		resp.Diagnostics = vpcRes.Diagnostics
	}

	validateSchemaReq := ValidateSchemaRequest{
		Config: tfsdk.Config{
			Raw:    config,
			Schema: schema,
		},
	}
	validateSchemaResp := ValidateSchemaResponse{
		Diagnostics: resp.Diagnostics,
	}

	SchemaValidate(ctx, schema, validateSchemaReq, &validateSchemaResp)

	resp.Diagnostics = validateSchemaResp.Diagnostics
}

// configureProviderResponse is a thin abstraction to allow native Diagnostics usage
type configureProviderResponse struct {
	Diagnostics diag.Diagnostics
}

func (r configureProviderResponse) toTfprotov6() *tfprotov6.ConfigureProviderResponse {
	return &tfprotov6.ConfigureProviderResponse{
		Diagnostics: toproto6.Diagnostics(r.Diagnostics),
	}
}

func (s *Server) ConfigureProvider(ctx context.Context, req *tfprotov6.ConfigureProviderRequest) (*tfprotov6.ConfigureProviderResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
	resp := &configureProviderResponse{}

	s.configureProvider(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *Server) configureProvider(ctx context.Context, req *tfprotov6.ConfigureProviderRequest, resp *configureProviderResponse) {
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetSchema")
	schema, diags := s.Provider.GetSchema(ctx)
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
	r := tfsdk.ConfigureProviderRequest{
		TerraformVersion: req.TerraformVersion,
		Config: tfsdk.Config{
			Raw:    config,
			Schema: schema,
		},
	}
	res := &tfsdk.ConfigureProviderResponse{}
	logging.FrameworkDebug(ctx, "Calling provider defined Provider Configure")
	s.Provider.Configure(ctx, r, res)
	logging.FrameworkDebug(ctx, "Called provider defined Provider Configure")
	resp.Diagnostics.Append(res.Diagnostics...)
}

func (s *Server) StopProvider(ctx context.Context, _ *tfprotov6.StopProviderRequest) (*tfprotov6.StopProviderResponse, error) {
	s.cancelRegisteredContexts(ctx)

	return &tfprotov6.StopProviderResponse{}, nil
}

// validateResourceConfigResponse is a thin abstraction to allow native Diagnostics usage
type validateResourceConfigResponse struct {
	Diagnostics diag.Diagnostics
}

func (r validateResourceConfigResponse) toTfprotov6() *tfprotov6.ValidateResourceConfigResponse {
	return &tfprotov6.ValidateResourceConfigResponse{
		Diagnostics: toproto6.Diagnostics(r.Diagnostics),
	}
}

func (s *Server) ValidateResourceConfig(ctx context.Context, req *tfprotov6.ValidateResourceConfigRequest) (*tfprotov6.ValidateResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
	resp := &validateResourceConfigResponse{}

	s.validateResourceConfig(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *Server) validateResourceConfig(ctx context.Context, req *tfprotov6.ValidateResourceConfigRequest, resp *validateResourceConfigResponse) {
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
	resource, diags := resourceType.NewResource(ctx, s.Provider)
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

	vrcReq := tfsdk.ValidateResourceConfigRequest{
		Config: tfsdk.Config{
			Raw:    config,
			Schema: resourceSchema,
		},
	}

	if resource, ok := resource.(tfsdk.ResourceWithConfigValidators); ok {
		logging.FrameworkTrace(ctx, "Resource implements ResourceWithConfigValidators")
		for _, configValidator := range resource.ConfigValidators(ctx) {
			vrcRes := &tfsdk.ValidateResourceConfigResponse{
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

	if resource, ok := resource.(tfsdk.ResourceWithValidateConfig); ok {
		logging.FrameworkTrace(ctx, "Resource implements ResourceWithValidateConfig")
		vrcRes := &tfsdk.ValidateResourceConfigResponse{
			Diagnostics: resp.Diagnostics,
		}

		logging.FrameworkDebug(ctx, "Calling provider defined Resource ValidateConfig")
		resource.ValidateConfig(ctx, vrcReq, vrcRes)
		logging.FrameworkDebug(ctx, "Called provider defined Resource ValidateConfig")

		resp.Diagnostics = vrcRes.Diagnostics
	}

	validateSchemaReq := ValidateSchemaRequest{
		Config: tfsdk.Config{
			Raw:    config,
			Schema: resourceSchema,
		},
	}
	validateSchemaResp := ValidateSchemaResponse{
		Diagnostics: resp.Diagnostics,
	}

	SchemaValidate(ctx, resourceSchema, validateSchemaReq, &validateSchemaResp)

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
		Diagnostics:   toproto6.Diagnostics(r.Diagnostics),
		UpgradedState: r.UpgradedState,
	}
}

func (s *Server) UpgradeResourceState(ctx context.Context, req *tfprotov6.UpgradeResourceStateRequest) (*tfprotov6.UpgradeResourceStateResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
	resp := &upgradeResourceStateResponse{}

	s.upgradeResourceState(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *Server) upgradeResourceState(ctx context.Context, req *tfprotov6.UpgradeResourceStateRequest, resp *upgradeResourceStateResponse) {
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
	resource, diags := resourceType.NewResource(ctx, s.Provider)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resourceWithUpgradeState, ok := resource.(tfsdk.ResourceWithUpgradeState)

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
		resourceStateUpgraders = make(map[int64]tfsdk.ResourceStateUpgrader, 0)
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

	upgradeResourceStateRequest := tfsdk.UpgradeResourceStateRequest{
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

		upgradeResourceStateRequest.State = &tfsdk.State{
			Raw:    rawStateValue,
			Schema: *resourceStateUpgrader.PriorSchema,
		}
	}

	upgradeResourceStateResponse := tfsdk.UpgradeResourceStateResponse{
		State: tfsdk.State{
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
		Diagnostics: toproto6.Diagnostics(r.Diagnostics),
		Private:     r.Private,
	}
}

func (s *Server) ReadResource(ctx context.Context, req *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
	resp := &readResourceResponse{}

	s.readResource(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *Server) readResource(ctx context.Context, req *tfprotov6.ReadResourceRequest, resp *readResourceResponse) {
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
	resource, diags := resourceType.NewResource(ctx, s.Provider)
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
	readReq := tfsdk.ReadResourceRequest{
		State: tfsdk.State{
			Raw:    state,
			Schema: resourceSchema,
		},
	}
	if pm, ok := s.Provider.(tfsdk.ProviderWithProviderMeta); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
		logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
		pmSchema, diags := pm.GetMetaSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		readReq.ProviderMeta = tfsdk.Config{
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
	readResp := tfsdk.ReadResourceResponse{
		State: tfsdk.State{
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

func markComputedNilsAsUnknown(ctx context.Context, config tftypes.Value, resourceSchema tfsdk.Schema) func(*tftypes.AttributePath, tftypes.Value) (tftypes.Value, error) {
	return func(path *tftypes.AttributePath, val tftypes.Value) (tftypes.Value, error) {
		ctx = logging.FrameworkWithAttributePath(ctx, path.String())

		// we are only modifying attributes, not the entire resource
		if len(path.Steps()) < 1 {
			return val, nil
		}
		configVal, _, err := tftypes.WalkAttributePath(config, path)
		if err != tftypes.ErrInvalidStep && err != nil {
			logging.FrameworkError(ctx, "error walking attribute path")
			return val, err
		} else if err != tftypes.ErrInvalidStep && !configVal.(tftypes.Value).IsNull() {
			logging.FrameworkTrace(ctx, "attribute not null in config, not marking unknown")
			return val, nil
		}
		attribute, err := resourceSchema.AttributeAtPath(path)
		if err != nil {
			if errors.Is(err, tfsdk.ErrPathInsideAtomicAttribute) {
				// ignore attributes/elements inside schema.Attributes, they have no schema of their own
				logging.FrameworkTrace(ctx, "attribute is a non-schema attribute, not marking unknown")
				return val, nil
			}
			logging.FrameworkError(ctx, "couldn't find attribute in resource schema")
			return tftypes.Value{}, fmt.Errorf("couldn't find attribute in resource schema: %w", err)
		}
		if !attribute.Computed {
			logging.FrameworkTrace(ctx, "attribute is not computed in schema, not marking unknown")
			return val, nil
		}
		logging.FrameworkDebug(ctx, "marking computed attribute that is null in the config as unknown")
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
		Diagnostics:     toproto6.Diagnostics(r.Diagnostics),
		RequiresReplace: r.RequiresReplace,
		PlannedPrivate:  r.PlannedPrivate,
	}
}

func (s *Server) PlanResourceChange(ctx context.Context, req *tfprotov6.PlanResourceChangeRequest) (*tfprotov6.PlanResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
	resp := &planResourceChangeResponse{}

	s.planResourceChange(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *Server) planResourceChange(ctx context.Context, req *tfprotov6.PlanResourceChangeRequest, resp *planResourceChangeResponse) {
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
	resource, diags := resourceType.NewResource(ctx, s.Provider)
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
			Config: tfsdk.Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			State: tfsdk.State{
				Schema: resourceSchema,
				Raw:    state,
			},
			Plan: tfsdk.Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
		}
		if pm, ok := s.Provider.(tfsdk.ProviderWithProviderMeta); ok {
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
			modifySchemaPlanReq.ProviderMeta = tfsdk.Config{
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
			Plan: tfsdk.Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
			Diagnostics: resp.Diagnostics,
		}

		SchemaModifyPlan(ctx, resourceSchema, modifySchemaPlanReq, &modifySchemaPlanResp)
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
	var modifyPlanResp tfsdk.ModifyResourcePlanResponse
	if resource, ok := resource.(tfsdk.ResourceWithModifyPlan); ok {
		logging.FrameworkTrace(ctx, "Resource implements ResourceWithModifyPlan")
		modifyPlanReq := tfsdk.ModifyResourcePlanRequest{
			Config: tfsdk.Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			State: tfsdk.State{
				Schema: resourceSchema,
				Raw:    state,
			},
			Plan: tfsdk.Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
		}
		if pm, ok := s.Provider.(tfsdk.ProviderWithProviderMeta); ok {
			logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
			logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
			pmSchema, diags := pm.GetMetaSchema(ctx)
			logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			modifyPlanReq.ProviderMeta = tfsdk.Config{
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

		modifyPlanResp = tfsdk.ModifyResourcePlanResponse{
			Plan: tfsdk.Plan{
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
		Diagnostics: toproto6.Diagnostics(r.Diagnostics),
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

func (s *Server) ApplyResourceChange(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
	resp := &applyResourceChangeResponse{
		// default to the prior state, so the state won't change unless
		// we choose to change it
		NewState: req.PriorState,
	}

	s.applyResourceChange(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *Server) applyResourceChange(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest, resp *applyResourceChangeResponse) {
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
	resource, diags := resourceType.NewResource(ctx, s.Provider)
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
		createReq := tfsdk.CreateResourceRequest{
			Config: tfsdk.Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			Plan: tfsdk.Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
		}
		if pm, ok := s.Provider.(tfsdk.ProviderWithProviderMeta); ok {
			logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
			logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
			pmSchema, diags := pm.GetMetaSchema(ctx)
			logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			createReq.ProviderMeta = tfsdk.Config{
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
		createResp := tfsdk.CreateResourceResponse{
			State: tfsdk.State{
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
		updateReq := tfsdk.UpdateResourceRequest{
			Config: tfsdk.Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			Plan: tfsdk.Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
			State: tfsdk.State{
				Schema: resourceSchema,
				Raw:    priorState,
			},
		}
		if pm, ok := s.Provider.(tfsdk.ProviderWithProviderMeta); ok {
			logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
			logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
			pmSchema, diags := pm.GetMetaSchema(ctx)
			logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			updateReq.ProviderMeta = tfsdk.Config{
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
		updateResp := tfsdk.UpdateResourceResponse{
			State: tfsdk.State{
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
		destroyReq := tfsdk.DeleteResourceRequest{
			State: tfsdk.State{
				Schema: resourceSchema,
				Raw:    priorState,
			},
		}
		if pm, ok := s.Provider.(tfsdk.ProviderWithProviderMeta); ok {
			logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
			logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
			pmSchema, diags := pm.GetMetaSchema(ctx)
			logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			destroyReq.ProviderMeta = tfsdk.Config{
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
		destroyResp := tfsdk.DeleteResourceResponse{
			State: tfsdk.State{
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
		Diagnostics: toproto6.Diagnostics(r.Diagnostics),
	}
}

func (s *Server) ValidateDataResourceConfig(ctx context.Context, req *tfprotov6.ValidateDataResourceConfigRequest) (*tfprotov6.ValidateDataResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
	resp := &validateDataResourceConfigResponse{}

	s.validateDataResourceConfig(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *Server) validateDataResourceConfig(ctx context.Context, req *tfprotov6.ValidateDataResourceConfigRequest, resp *validateDataResourceConfigResponse) {

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
	dataSource, diags := dataSourceType.NewDataSource(ctx, s.Provider)
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

	vrcReq := tfsdk.ValidateDataSourceConfigRequest{
		Config: tfsdk.Config{
			Raw:    config,
			Schema: dataSourceSchema,
		},
	}

	if dataSource, ok := dataSource.(tfsdk.DataSourceWithConfigValidators); ok {
		logging.FrameworkTrace(ctx, "DataSource implements DataSourceWithConfigValidators")
		for _, configValidator := range dataSource.ConfigValidators(ctx) {
			vrcRes := &tfsdk.ValidateDataSourceConfigResponse{
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

	if dataSource, ok := dataSource.(tfsdk.DataSourceWithValidateConfig); ok {
		logging.FrameworkTrace(ctx, "DataSource implements DataSourceWithValidateConfig")
		vrcRes := &tfsdk.ValidateDataSourceConfigResponse{
			Diagnostics: resp.Diagnostics,
		}

		logging.FrameworkDebug(ctx, "Calling provider defined DataSource ValidateConfig")
		dataSource.ValidateConfig(ctx, vrcReq, vrcRes)
		logging.FrameworkDebug(ctx, "Called provider defined DataSource ValidateConfig")

		resp.Diagnostics = vrcRes.Diagnostics
	}

	validateSchemaReq := ValidateSchemaRequest{
		Config: tfsdk.Config{
			Raw:    config,
			Schema: dataSourceSchema,
		},
	}
	validateSchemaResp := ValidateSchemaResponse{
		Diagnostics: resp.Diagnostics,
	}

	SchemaValidate(ctx, dataSourceSchema, validateSchemaReq, &validateSchemaResp)

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
		Diagnostics: toproto6.Diagnostics(r.Diagnostics),
	}
}

func (s *Server) ReadDataSource(ctx context.Context, req *tfprotov6.ReadDataSourceRequest) (*tfprotov6.ReadDataSourceResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
	resp := &readDataSourceResponse{}

	s.readDataSource(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *Server) readDataSource(ctx context.Context, req *tfprotov6.ReadDataSourceRequest, resp *readDataSourceResponse) {
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
	dataSource, diags := dataSourceType.NewDataSource(ctx, s.Provider)
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
	readReq := tfsdk.ReadDataSourceRequest{
		Config: tfsdk.Config{
			Raw:    config,
			Schema: dataSourceSchema,
		},
	}
	if pm, ok := s.Provider.(tfsdk.ProviderWithProviderMeta); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")
		logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
		pmSchema, diags := pm.GetMetaSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		readReq.ProviderMeta = tfsdk.Config{
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
	readResp := tfsdk.ReadDataSourceResponse{
		State: tfsdk.State{
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
