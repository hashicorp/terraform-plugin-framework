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

func TestServerPlanAction(t *testing.T) {
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

	testDeferralAllowed := action.ModifyPlanClientCapabilities{
		DeferralAllowed: true,
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.PlanActionRequest
		expectedResponse     *fwserver.PlanActionResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.PlanActionResponse{},
		},
		"unlinked-nil-config-no-modifyplan": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanActionRequest{
				ActionSchema: testUnlinkedSchema,
				Action:       &testprovider.Action{},
			},
			expectedResponse: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{},
			},
		},
		"request-client-capabilities-deferral-allowed": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanActionRequest{
				ClientCapabilities: testDeferralAllowed,
				Config:             testUnlinkedConfig,
				ActionSchema:       testUnlinkedSchema,
				Action: &testprovider.ActionWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
						if req.ClientCapabilities.DeferralAllowed != true {
							resp.Diagnostics.AddError("Unexpected req.ClientCapabilities.DeferralAllowed value",
								"expected: true but got: false")
						}

						var config struct {
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
					},
				},
			},
			expectedResponse: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{},
			},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanActionRequest{
				Config:       testUnlinkedConfig,
				ActionSchema: testUnlinkedSchema,
				Action: &testprovider.ActionWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
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
			expectedResponse: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{},
			},
		},
		"request-linkedresources": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.PlanActionLinkedResourceRequest{
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
						PriorIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
				Action: &testprovider.ActionWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
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
			expectedResponse: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{
					{
						PlannedState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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
			},
		},
		"action-configure-data": {
			server: &fwserver.Server{
				ActionConfigureData: "test-provider-configure-value",
				Provider:            &testprovider.Provider{},
			},
			request: &fwserver.PlanActionRequest{
				Config:       testUnlinkedConfig,
				ActionSchema: testUnlinkedSchema,
				Action: &testprovider.ActionWithConfigureAndModifyPlan{
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
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
						// In practice, the Configure method would save the
						// provider data to the Action implementation and
						// use it here. The fact that Configure is able to
						// read the data proves this can work.
					},
				},
			},
			expectedResponse: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{},
			},
		},
		"response-deferral-automatic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
						resp.Deferred = &provider.Deferred{Reason: provider.DeferredReasonProviderConfigUnknown}
					},
				},
			},
			configureProviderReq: &provider.ConfigureRequest{
				ClientCapabilities: provider.ConfigureProviderClientCapabilities{
					DeferralAllowed: true,
				},
			},
			request: &fwserver.PlanActionRequest{
				Config:       testUnlinkedConfig,
				ActionSchema: testUnlinkedSchema,
				Action: &testprovider.ActionWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
						resp.Diagnostics.AddError("Test assertion failed: ", "ModifyPlan shouldn't be called")
					},
				},
				ClientCapabilities: testDeferralAllowed,
			},
			expectedResponse: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{},
				Deferred:        &action.Deferred{Reason: action.DeferredReasonProviderConfigUnknown},
			},
		},
		"response-deferral-manual": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanActionRequest{
				Config:       testUnlinkedConfig,
				ActionSchema: testUnlinkedSchema,
				Action: &testprovider.ActionWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
						var config struct {
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

						resp.Deferred = &action.Deferred{Reason: action.DeferredReasonAbsentPrereq}

						if config.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("unexpected req.Config value: %s", config.TestRequired.ValueString())
						}
					},
				},
				ClientCapabilities: testDeferralAllowed,
			},
			expectedResponse: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{},
				Deferred:        &action.Deferred{Reason: action.DeferredReasonAbsentPrereq},
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanActionRequest{
				Config:       testUnlinkedConfig,
				ActionSchema: testUnlinkedSchema,
				Action: &testprovider.ActionWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{},
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
			request: &fwserver.PlanActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.PlanActionLinkedResourceRequest{
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
						PriorIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
				Action: &testprovider.ActionWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
						// Should be copied over from request
						if len(resp.LinkedResources) != 1 {
							resp.Diagnostics.AddError("unexpected resp.LinkedResources value", fmt.Sprintf("got %d, expected 1", len(req.LinkedResources)))
						}

						resp.Diagnostics.Append(resp.LinkedResources[0].Plan.SetAttribute(ctx, path.Root("test_computed"), "new-plan-value")...)
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
			expectedResponse: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{
					{
						PlannedState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, "new-plan-value"),
								"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
							}),
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
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
			request: &fwserver.PlanActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.PlanActionLinkedResourceRequest{
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
						PriorIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
				Action: &testprovider.ActionWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
						resp.LinkedResources = make([]action.ModifyPlanResponseLinkedResource, 0)
					},
				},
			},
			expectedResponse: &fwserver.PlanActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Linked Resource Plan",
						"An unexpected error was encountered when planning an action with linked resources. "+
							"The number of linked resources produced by the action plan cannot change: 0 linked resource(s) were produced in the plan, expected 1\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{
					{
						PlannedState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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
			},
		},
		"response-linkedresources-added": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanActionRequest{
				ActionSchema: testEmptyActionSchema,
				LinkedResources: []*fwserver.PlanActionLinkedResourceRequest{
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
						PriorIdentity: &tfsdk.ResourceIdentity{
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
						PriorIdentity: &tfsdk.ResourceIdentity{
							Raw: tftypes.NewValue(testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
				Action: &testprovider.ActionWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
						resp.LinkedResources = append(resp.LinkedResources, action.ModifyPlanResponseLinkedResource{})
					},
				},
			},
			expectedResponse: &fwserver.PlanActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Linked Resource Plan",
						"An unexpected error was encountered when planning an action with linked resources. "+
							"The number of linked resources produced by the action plan cannot change: 3 linked resource(s) were produced in the plan, expected 2\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
				LinkedResources: []*fwserver.PlanActionLinkedResourceResponse{
					{
						PlannedState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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
						PlannedState: &tfsdk.State{
							Raw: tftypes.NewValue(testLinkedResourceSchemaType, map[string]tftypes.Value{
								"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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

			response := &fwserver.PlanActionResponse{}
			testCase.server.PlanAction(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
