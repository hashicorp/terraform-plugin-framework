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
	"github.com/hashicorp/terraform-plugin-framework/provider"
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
			expectedResponse: &fwserver.PlanActionResponse{},
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
			expectedResponse: &fwserver.PlanActionResponse{},
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
			expectedResponse: &fwserver.PlanActionResponse{},
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
			expectedResponse: &fwserver.PlanActionResponse{},
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
				Deferred: &action.Deferred{Reason: action.DeferredReasonProviderConfigUnknown},
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
				Deferred: &action.Deferred{Reason: action.DeferredReasonAbsentPrereq},
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
