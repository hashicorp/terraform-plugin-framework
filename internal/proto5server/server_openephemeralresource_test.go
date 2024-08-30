// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

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
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
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

	testStateDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
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
		request          *tfprotov5.OpenEphemeralResourceRequest
		expectedError    error
		expectedResponse *tfprotov5.OpenEphemeralResourceResponse
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
			request: &tfprotov5.OpenEphemeralResourceRequest{
				Config:   testEmptyDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov5.OpenEphemeralResourceResponse{
				State: testEmptyDynamicValue,
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
			request: &tfprotov5.OpenEphemeralResourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov5.OpenEphemeralResourceResponse{
				State: testConfigDynamicValue,
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
			request: &tfprotov5.OpenEphemeralResourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov5.OpenEphemeralResourceResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
				},
				State: testConfigDynamicValue,
			},
		},
		"response-is-closable": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									// Implementing ephemeral.EphemeralResourceWithClose will set IsClosable to true
									return &testprovider.EphemeralResourceWithClose{
										EphemeralResource: &testprovider.EphemeralResource{
											SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
												resp.Schema = schema.Schema{}
											},
											MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
												resp.TypeName = "test_ephemeral_resource"
											},
										},
										CloseMethod: func(ctx context.Context, _ ephemeral.CloseRequest, _ *ephemeral.CloseResponse) {},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.OpenEphemeralResourceRequest{
				Config:   testEmptyDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov5.OpenEphemeralResourceResponse{
				State:      testEmptyDynamicValue,
				IsClosable: true,
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
											resp.RenewAt = time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.OpenEphemeralResourceRequest{
				Config:   testEmptyDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov5.OpenEphemeralResourceResponse{
				State:   testEmptyDynamicValue,
				RenewAt: time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC),
			},
		},
		"response-state": {
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

											data.TestComputed = types.StringValue("test-state-value")

											resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.OpenEphemeralResourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov5.OpenEphemeralResourceResponse{
				State: testStateDynamicValue,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

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
