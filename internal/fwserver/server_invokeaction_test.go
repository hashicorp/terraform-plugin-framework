// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerInvokeAction(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_required": tftypes.String,
		},
	}

	testConfigValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testUnlinkedSchema := schema.UnlinkedSchema{
		Attributes: map[string]schema.Attribute{
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testEmptyActionSchema := schema.UnlinkedSchema{
		Attributes: map[string]schema.Attribute{},
	}

	testLinkedResourceSchema := resourceschema.Schema{
		Attributes: map[string]resourceschema.Attribute{
			"test_computed": resourceschema.StringAttribute{
				Computed: true,
			},
			"test_required": resourceschema.StringAttribute{
				Required: true,
			},
		},
	}

	testLinkedResourceSchemaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testLinkedResourceIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}

	testLinkedResourceIdentitySchemaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id": tftypes.String,
		},
	}

	testUnlinkedConfig := &tfsdk.Config{
		Raw:    testConfigValue,
		Schema: testUnlinkedSchema,
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.InvokeActionRequest
		expectedResponse     *fwserver.InvokeActionResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.InvokeActionResponse{},
		},
		"unlinked-nil-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				ActionSchema: testUnlinkedSchema,
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						if !req.Config.Raw.IsNull() {
							resp.Diagnostics.AddError("Unexpected Config in action Invoke", "Expected Config to be null")
						}
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{},
			},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				Config:       testUnlinkedConfig,
				ActionSchema: testUnlinkedSchema,
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						var config struct {
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

						if config.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("unexpected req.Config value: %s", config.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{},
			},
		},
		"request-linkedresources": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.InvokeActionRequestLinkedResource{
					{
						Config: &tfsdk.Config{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, nil),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						var linkedResourceData struct {
							TestRequired types.String `tfsdk:"test_required"`
							TestComputed types.String `tfsdk:"test_computed"`
						}
						var linkedResourceIdentityData struct {
							TestID types.String `tfsdk:"test_id"`
						}

						if len(req.LinkedResources) != 1 {
							resp.Diagnostics.AddError("unexpected req.LinkedResources value", fmt.Sprintf("got %d, expected 1", len(req.LinkedResources)))
						}

						resp.Diagnostics.Append(req.LinkedResources[0].Plan.Get(ctx, &linkedResourceData)...)
						if resp.Diagnostics.HasError() {
							return
						}

						if !linkedResourceData.TestComputed.IsUnknown() {
							resp.Diagnostics.AddError(
								"unexpected req.LinkedResources value",
								fmt.Sprintf("expected linked resource data to be unknown, got: %s", linkedResourceData.TestComputed),
							)
							return
						}

						resp.Diagnostics.Append(req.LinkedResources[0].Config.Get(ctx, &linkedResourceData)...)
						if resp.Diagnostics.HasError() {
							return
						}

						if !linkedResourceData.TestComputed.IsNull() {
							resp.Diagnostics.AddError(
								"unexpected req.LinkedResources value",
								fmt.Sprintf("expected linked resource data to be null, got: %s", linkedResourceData.TestComputed),
							)
							return
						}

						resp.Diagnostics.Append(req.LinkedResources[0].State.Get(ctx, &linkedResourceData)...)
						if resp.Diagnostics.HasError() {
							return
						}

						if linkedResourceData.TestComputed.ValueString() != "test-state-value" {
							resp.Diagnostics.AddError(
								"unexpected req.LinkedResources value",
								fmt.Sprintf("expected linked resource data to be \"test-state-value\", got: %s", linkedResourceData.TestComputed),
							)
							return
						}

						resp.Diagnostics.Append(req.LinkedResources[0].Identity.Get(ctx, &linkedResourceIdentityData)...)
						if resp.Diagnostics.HasError() {
							return
						}

						if linkedResourceIdentityData.TestID.ValueString() != "id-123" {
							resp.Diagnostics.AddError(
								"unexpected req.LinkedResources value",
								fmt.Sprintf("expected linked resource data to be \"id-123\", got: %s", linkedResourceIdentityData.TestID),
							)
							return
						}
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{
					{
						NewState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						NewIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
			},
		},
		"action-configure-data": {
			server: &fwserver.Server{
				ActionConfigureData: "test-provider-configure-value",
				Provider:            &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				Config:       testUnlinkedConfig,
				ActionSchema: testUnlinkedSchema,
				Action: &testprovider.ActionWithConfigure{
					ConfigureMethod: func(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
						providerData, ok := req.ProviderData.(string)

						if !ok {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.ProviderData",
								fmt.Sprintf("Expected string, got: %T", req.ProviderData),
							)
							return
						}

						if providerData != "test-provider-configure-value" {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.ProviderData",
								fmt.Sprintf("Expected test-provider-configure-value, got: %q", providerData),
							)
						}
					},
					Action: &testprovider.Action{
						InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
							// In practice, the Configure method would save the
							// provider data to the Action implementation and
							// use it here. The fact that Configure is able to
							// read the data proves this can work.
						},
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{},
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				Config:       testUnlinkedConfig,
				ActionSchema: testUnlinkedSchema,
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{},
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic(
						"warning summary",
						"warning detail",
					),
					diag.NewErrorDiagnostic(
						"error summary",
						"error detail",
					),
				},
			},
		},
		"response-linkedresources": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.InvokeActionRequestLinkedResource{
					{
						Config: &tfsdk.Config{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, nil),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						// Should be copied over from request
						if len(resp.LinkedResources) != 1 {
							resp.Diagnostics.AddError("unexpected resp.LinkedResources value", fmt.Sprintf("got %d, expected 1", len(req.LinkedResources)))
						}

						resp.Diagnostics.Append(resp.LinkedResources[0].State.SetAttribute(ctx, path.Root("test_computed"), "new-state-value")...)
						if resp.Diagnostics.HasError() {
							return
						}

						resp.Diagnostics.Append(resp.LinkedResources[0].Identity.SetAttribute(ctx, path.Root("test_id"), "new-id-123")...)
						if resp.Diagnostics.HasError() {
							return
						}
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{
					{
						NewState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "new-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						NewIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "new-id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
			},
		},
		"response-linkedresources-removed": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.InvokeActionRequestLinkedResource{
					{
						Config: &tfsdk.Config{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, nil),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						resp.LinkedResources = make([]action.InvokeResponseLinkedResource, 0)
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Linked Resource State",
						"An unexpected error was encountered when invoking an action with linked resources. "+
							"The number of linked resource states produced by the action invoke cannot change: 0 linked resource(s) were planned, expected 1\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{
					{
						NewState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						NewIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
			},
		},
		"response-linkedresources-added": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.InvokeActionRequestLinkedResource{
					{
						Config: &tfsdk.Config{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, nil),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
					{
						Config: &tfsdk.Config{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, nil),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						resp.LinkedResources = append(resp.LinkedResources, action.InvokeResponseLinkedResource{})
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Linked Resource State",
						"An unexpected error was encountered when invoking an action with linked resources. "+
							"The number of linked resource states produced by the action invoke cannot change: 3 linked resource(s) were planned, expected 2\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{
					{
						NewState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						NewIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
					{
						NewState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						NewIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
			},
		},
		"response-linkedresources-valid-replacement": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.InvokeActionRequestLinkedResource{
					{
						Config: &tfsdk.Config{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, nil),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
					},
					{
						Config: &tfsdk.Config{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, nil),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
					},
				},
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						resp.LinkedResources[0].RequiresReplace = true
						resp.Diagnostics.AddError("error summary", "error detail")

						resp.Diagnostics.Append(resp.LinkedResources[1].State.SetAttribute(ctx, path.Root("test_computed"), "new-state-value")...)
						if resp.Diagnostics.HasError() {
							return
						}
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"error summary",
						"error detail",
					),
				},
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{
					{
						RequiresReplace: true,
						NewState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
					},
					{
						NewState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "new-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
					},
				},
			},
		},
		"response-linkedresources-invalid-replacement": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.InvokeActionRequestLinkedResource{
					{
						Config: &tfsdk.Config{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, nil),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
					},
					{
						Config: &tfsdk.Config{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, nil),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
					},
				},
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						// Not allowed to require replacement on a resource without a diagnostic
						resp.LinkedResources[0].RequiresReplace = true

						resp.Diagnostics.Append(resp.LinkedResources[1].State.SetAttribute(ctx, path.Root("test_computed"), "new-state-value")...)
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Linked Resource Replacement",
						"An unexpected error was encountered when invoking an action with linked resources. "+
							"The Terraform Provider returned a linked resource (at index 0) that "+
							"indicates that it needs to be replaced, but no error diagnostics were returned.\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{
					{
						RequiresReplace: true,
						NewState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
					},
					{
						NewState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "new-state-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.configureProviderReq != nil {
				configureProviderResp := &provider.ConfigureResponse{}
				testCase.server.ConfigureProvider(context.Background(), testCase.configureProviderReq, configureProviderResp)
			}

			response := &fwserver.InvokeActionResponse{}
			testCase.server.InvokeAction(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
