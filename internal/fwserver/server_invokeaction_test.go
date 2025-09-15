// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testConfig := &tfsdk.Config{
		Raw:    testConfigValue,
		Schema: testSchema,
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
		"nil-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				ActionSchema: testSchema,
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						if !req.Config.Raw.IsNull() {
							resp.Diagnostics.AddError("Unexpected Config in action Invoke", "Expected Config to be null")
						}
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				Config:       testConfig,
				ActionSchema: testSchema,
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
			expectedResponse: &fwserver.InvokeActionResponse{},
		},
		"action-configure-data": {
			server: &fwserver.Server{
				ActionConfigureData: "test-provider-configure-value",
				Provider:            &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				Config:       testConfig,
				ActionSchema: testSchema,
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
			expectedResponse: &fwserver.InvokeActionResponse{},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.InvokeActionRequest{
				Config:       testConfig,
				ActionSchema: testSchema,
				Action: &testprovider.Action{
					InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.InvokeActionResponse{
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

			response := &fwserver.InvokeActionResponse{}
			testCase.server.InvokeAction(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
