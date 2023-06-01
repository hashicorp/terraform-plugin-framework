// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerUpgradeResourceState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"optional_attribute": schema.StringAttribute{
				Optional: true,
			},
			"required_attribute": schema.StringAttribute{
				Required: true,
			},
		},
		Version: 1, // Must be above 0
	}
	schemaType := schema.Type().TerraformType(ctx)

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.UpgradeResourceStateRequest
		expectedResponse *tfprotov6.UpgradeResourceStateResponse
		expectedError    error
	}{
		"nil": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request:          nil,
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{},
		},
		"request-RawState": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithUpgradeState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = schema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
											return map[int64]resource.StateUpgrader{
												0: {
													StateUpgrader: func(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
														expectedRawState := testNewRawState(t, map[string]interface{}{
															"id":                 "test-id-value",
															"required_attribute": true,
														})

														if diff := cmp.Diff(req.RawState, expectedRawState); diff != "" {
															resp.Diagnostics.AddError("unexpected req.RawState difference: %s", diff)
														}

														// Prevent Missing Upgraded Resource State error
														resp.State = tfsdk.State{
															Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
																"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
																"optional_attribute": tftypes.NewValue(tftypes.String, nil),
																"required_attribute": tftypes.NewValue(tftypes.String, "true"),
															}),
															Schema: schema,
														}
													},
												},
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_resource",
				Version:  0,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				UpgradedState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-TypeName-missing": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request: &tfprotov6.UpgradeResourceStateRequest{},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Resource Type Not Found",
						Detail:   "No resource type named \"\" was found in the provider.",
					},
				},
			},
		},
		"request-TypeName-unknown": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request: &tfprotov6.UpgradeResourceStateRequest{
				TypeName: "unknown",
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Resource Type Not Found",
						Detail:   "No resource type named \"unknown\" was found in the provider.",
					},
				},
			},
		},
		"response-Diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithUpgradeState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = schema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
											return map[int64]resource.StateUpgrader{
												0: {
													StateUpgrader: func(_ context.Context, _ resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
														resp.Diagnostics.AddWarning("warning summary", "warning detail")
														resp.Diagnostics.AddError("error summary", "error detail")
													},
												},
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_resource",
				Version:  0,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
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
			},
		},
		"response-UpgradedState": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithUpgradeState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = schema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
											return map[int64]resource.StateUpgrader{
												0: {
													StateUpgrader: func(_ context.Context, _ resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
														resp.State = tfsdk.State{
															Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
																"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
																"optional_attribute": tftypes.NewValue(tftypes.String, nil),
																"required_attribute": tftypes.NewValue(tftypes.String, "true"),
															}),
															Schema: schema,
														}
													},
												},
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_resource",
				Version:  0,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				UpgradedState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.UpgradeResourceState(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
