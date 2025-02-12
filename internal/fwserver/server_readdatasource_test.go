// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerReadDataSource(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testConfigValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, nil),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testStateUnknownValue := tftypes.NewValue(testType, tftypes.UnknownValue)

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testSchemaWithSemanticEquals := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				CustomType: testtypes.StringTypeWithSemanticEquals{
					SemanticEquals: true,
				},
				Required: true,
			},
		},
	}

	testSchemaWithSemanticEqualsDiagnostics := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				CustomType: testtypes.StringTypeWithSemanticEquals{
					SemanticEquals: true,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Required: true,
			},
		},
	}

	testConfig := &tfsdk.Config{
		Raw:    testConfigValue,
		Schema: testSchema,
	}

	testStateUnchanged := &tfsdk.State{
		Raw:    testConfigValue,
		Schema: testSchema,
	}

	testStateUnknown := &tfsdk.State{
		Raw:    testStateUnknownValue,
		Schema: testSchema,
	}

	testState := &tfsdk.State{
		Raw:    testStateValue,
		Schema: testSchema,
	}

	testDeferralAllowed := datasource.ReadClientCapabilities{
		DeferralAllowed: true,
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.ReadDataSourceRequest
		expectedResponse     *fwserver.ReadDataSourceResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{},
		},
		"request-client-capabilities-deferral-allowed": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				ClientCapabilities: testDeferralAllowed,
				Config:             testConfig,
				DataSourceSchema:   testSchema,
				DataSource: &testprovider.DataSource{
					ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
						if req.ClientCapabilities.DeferralAllowed != true {
							resp.Diagnostics.AddError("Unexpected req.ClientCapabilities.DeferralAllowed value",
								"expected: true but got: false")
						}

						var config struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State: testStateUnchanged,
			},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSource: &testprovider.DataSource{
					ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
						var config struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

						if config.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("unexpected req.Config value: %s", config.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State: testStateUnchanged,
			},
		},
		"request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSource: &testprovider.DataSource{
					ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
						var config struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &config)...)

						if config.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", config.TestRequired.ValueString())
						}
					},
				},
				ProviderMeta: testConfig,
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State: testStateUnchanged,
			},
		},
		"resource-configure-data": {
			server: &fwserver.Server{
				DataSourceConfigureData: "test-provider-configure-value",
				Provider:                &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSource: &testprovider.DataSourceWithConfigure{
					ConfigureMethod: func(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
					DataSource: &testprovider.DataSource{
						ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
							// In practice, the Configure method would save the
							// provider data to the Resource implementation and
							// use it here. The fact that Configure is able to
							// read the data proves this can work.
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State: testStateUnchanged,
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
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSource: &testprovider.DataSource{
					ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
						resp.Diagnostics.AddError("Test assertion failed: ", "read shouldn't be called")
					},
				},
				ClientCapabilities: testDeferralAllowed,
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State:    testStateUnknown,
				Deferred: &datasource.Deferred{Reason: datasource.DeferredReasonProviderConfigUnknown},
			},
		},
		"response-deferral-manual": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSource: &testprovider.DataSource{
					ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
						var config struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

						resp.Deferred = &datasource.Deferred{Reason: datasource.DeferredReasonAbsentPrereq}

						if config.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("unexpected req.Config value: %s", config.TestRequired.ValueString())
						}
					},
				},
				ClientCapabilities: testDeferralAllowed,
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State:    testStateUnchanged,
				Deferred: &datasource.Deferred{Reason: datasource.DeferredReasonAbsentPrereq},
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSource: &testprovider.DataSource{
					ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
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
				State: testStateUnchanged,
			},
		},
		"response-diagnostics-semantic-equality": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaWithSemanticEqualsDiagnostics,
				},
				DataSourceSchema: testSchemaWithSemanticEqualsDiagnostics,
				DataSource: &testprovider.DataSource{
					ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
						var data struct {
							TestComputed types.String                            `tfsdk:"test_computed"`
							TestRequired testtypes.StringValueWithSemanticEquals `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						data.TestRequired = testtypes.StringValueWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
							StringValue: types.StringValue("test-semantic-equal-value"),
						}

						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						// The response state is intentionally not updated when there are diagnostics
						"test_required": tftypes.NewValue(tftypes.String, "test-semantic-equal-value"),
					}),
					Schema: testSchemaWithSemanticEqualsDiagnostics,
				},
			},
		},
		"response-state": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSource: &testprovider.DataSource{
					ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
						var data struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						data.TestComputed = types.StringValue("test-state-value")

						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State: testState,
			},
		},
		"response-state-semantic-equality": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaWithSemanticEquals,
				},
				DataSourceSchema: testSchemaWithSemanticEquals,
				DataSource: &testprovider.DataSource{
					ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
						var data struct {
							TestComputed types.String                            `tfsdk:"test_computed"`
							TestRequired testtypes.StringValueWithSemanticEquals `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						// This value should be overwritten back to the config value.
						data.TestRequired = testtypes.StringValueWithSemanticEquals{
							SemanticEquals: true,
							StringValue:    types.StringValue("test-semantic-equal-value"),
						}

						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State: &tfsdk.State{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaWithSemanticEquals,
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

			response := &fwserver.ReadDataSourceResponse{}
			testCase.server.ReadDataSource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
