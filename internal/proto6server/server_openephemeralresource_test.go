// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerOpenEphemeralResource(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testConfigDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, nil),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testEmptyDynamicValue := testNewDynamicValue(t, tftypes.Object{}, nil)

	testResultDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-result-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

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

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.OpenEphemeralResourceRequest
		expectedError    error
		expectedResponse *tfprotov6.OpenEphemeralResourceResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
											resp.Schema = schema.Schema{}
										},
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.OpenEphemeralResourceRequest{
				Config:   testEmptyDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov6.OpenEphemeralResourceResponse{
				Result: testEmptyDynamicValue,
			},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource"
										},
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
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.OpenEphemeralResourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov6.OpenEphemeralResourceResponse{
				Result: testConfigDynamicValue,
			},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource"
										},
										OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
											resp.Diagnostics.AddWarning("warning summary", "warning detail")
											resp.Diagnostics.AddError("error summary", "error detail")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.OpenEphemeralResourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov6.OpenEphemeralResourceResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
				},
				Result: testConfigDynamicValue,
			},
		},
		"response-renew-at": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
											resp.Schema = schema.Schema{}
										},
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource"
										},
										OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
											resp.RenewAt = time.Date(2024, 8, 29, 6, 10, 32, 0, time.UTC)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.OpenEphemeralResourceRequest{
				Config:   testEmptyDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov6.OpenEphemeralResourceResponse{
				Result:  testEmptyDynamicValue,
				RenewAt: time.Date(2024, 8, 29, 6, 10, 32, 0, time.UTC),
			},
		},
		"response-result": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource"
										},
										OpenMethod: func(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
											var data struct {
												TestComputed types.String `tfsdk:"test_computed"`
												TestRequired types.String `tfsdk:"test_required"`
											}

											resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

											data.TestComputed = types.StringValue("test-result-value")

											resp.Diagnostics.Append(resp.Result.Set(ctx, data)...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.OpenEphemeralResourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov6.OpenEphemeralResourceResponse{
				Result: testResultDynamicValue,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.OpenEphemeralResource(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
