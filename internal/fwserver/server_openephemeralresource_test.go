// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerOpenEphemeralResource(t *testing.T) {
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

	testResultValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-result-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testResultUnknownValue := tftypes.NewValue(testType, tftypes.UnknownValue)

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

	testConfig := &tfsdk.Config{
		Raw:    testConfigValue,
		Schema: testSchema,
	}

	testResultUnchanged := &tfsdk.EphemeralResultData{
		Raw:    testConfigValue,
		Schema: testSchema,
	}

	testResultUnknown := &tfsdk.EphemeralResultData{
		Raw:    testResultUnknownValue,
		Schema: testSchema,
	}

	testResult := &tfsdk.EphemeralResultData{
		Raw:    testResultValue,
		Schema: testSchema,
	}

	testDeferralAllowed := ephemeral.OpenClientCapabilities{
		DeferralAllowed: true,
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testPrivateProvider := &privatestate.Data{
		Provider: testProviderData,
	}

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testEmptyPrivate := &privatestate.Data{
		Provider: testEmptyProviderData,
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.OpenEphemeralResourceRequest
		expectedResponse     *fwserver.OpenEphemeralResourceResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{},
		},
		"request-client-capabilities-deferral-allowed": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.OpenEphemeralResourceRequest{
				ClientCapabilities:      testDeferralAllowed,
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResource{
					OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
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
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
				Result:  testResultUnchanged,
				Private: testEmptyPrivate,
			},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.OpenEphemeralResourceRequest{
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResource{
					OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
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
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
				Result:  testResultUnchanged,
				Private: testEmptyPrivate,
			},
		},
		"ephemeralresource-configure-data": {
			server: &fwserver.Server{
				EphemeralResourceConfigureData: "test-provider-configure-value",
				Provider:                       &testprovider.Provider{},
			},
			request: &fwserver.OpenEphemeralResourceRequest{
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithConfigure{
					ConfigureMethod: func(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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
					EphemeralResource: &testprovider.EphemeralResource{
						OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
							// In practice, the Configure method would save the
							// provider data to the EphemeralResource implementation and
							// use it here. The fact that Configure is able to
							// read the data proves this can work.
						},
					},
				},
			},
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
				Result:  testResultUnchanged,
				Private: testEmptyPrivate,
			},
		},
		"response-default-values": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.OpenEphemeralResourceRequest{
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResource{
					OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {},
				},
			},
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
				Result:  testResultUnchanged,
				Private: testEmptyPrivate,
				RenewAt: *new(time.Time),
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
			request: &fwserver.OpenEphemeralResourceRequest{
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResource{
					OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
						resp.Diagnostics.AddError("Test assertion failed: ", "open shouldn't be called")
					},
				},
				ClientCapabilities: testDeferralAllowed,
			},
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
				Result:   testResultUnknown,
				Deferred: &ephemeral.Deferred{Reason: ephemeral.DeferredReasonProviderConfigUnknown},
			},
		},
		"response-deferral-manual": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.OpenEphemeralResourceRequest{
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResource{
					OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
						var config struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

						resp.Deferred = &ephemeral.Deferred{Reason: ephemeral.DeferredReasonAbsentPrereq}

						if config.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("unexpected req.Config value: %s", config.TestRequired.ValueString())
						}
					},
				},
				ClientCapabilities: testDeferralAllowed,
			},
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
				Result:   testResultUnchanged,
				Private:  testEmptyPrivate,
				Deferred: &ephemeral.Deferred{Reason: ephemeral.DeferredReasonAbsentPrereq},
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.OpenEphemeralResourceRequest{
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResource{
					OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
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
				Result:  testResultUnchanged,
				Private: testEmptyPrivate,
			},
		},
		"response-renew-at": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.OpenEphemeralResourceRequest{
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResource{
					OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
						resp.RenewAt = time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC)
					},
				},
			},
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
				Result:  testResultUnchanged,
				Private: testEmptyPrivate,
				RenewAt: time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC),
			},
		},
		"response-result": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.OpenEphemeralResourceRequest{
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResource{
					OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
						var data struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						data.TestComputed = types.StringValue("test-result-value")

						resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
				Result:  testResult,
				Private: testEmptyPrivate,
			},
		},
		"response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.OpenEphemeralResourceRequest{
				Config:                  testConfig,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResource{
					OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.OpenEphemeralResourceResponse{
				Result:  testResultUnchanged,
				Private: testPrivateProvider,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.configureProviderReq != nil {
				configureProviderResp := &provider.ConfigureResponse{}
				testCase.server.ConfigureProvider(context.Background(), testCase.configureProviderReq, configureProviderResp)
			}

			response := &fwserver.OpenEphemeralResourceResponse{}
			testCase.server.OpenEphemeralResource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
