// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerUpgradeResourceIdentity(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
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

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
		Version: 1, // Must be above 0
	}

	ctx := context.Background()
	testIdentityType := testIdentitySchema.Type().TerraformType(ctx)

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.UpgradeResourceIdentityRequest
		expectedResponse *tfprotov5.UpgradeResourceIdentityResponse
		expectedError    error
	}{
		"nil": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request:          nil,
			expectedResponse: &tfprotov5.UpgradeResourceIdentityResponse{},
		},
		"request-RawIdentity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithUpgradeResourceIdentity{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
											return map[int64]resource.IdentityUpgrader{
												0: {
													IdentityUpgrader: func(_ context.Context, req resource.UpgradeIdentityRequest, resp *resource.UpgradeIdentityResponse) {
														expectedSourceIdentity := testNewTfprotov6RawState(t, map[string]interface{}{
															"test_id": "test-id-value",
														})

														if diff := cmp.Diff(req.RawIdentity, expectedSourceIdentity); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceIdentity difference", diff)
														}

														// Prevent Missing Upgraded Resource Identity error
														resp.Identity = &tfsdk.ResourceIdentity{
															Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
																"test_id": tftypes.NewValue(tftypes.String, "test-id-value"),
															}),
															Schema: testIdentitySchema,
														}
													},
												},
											}
										},
										IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
											resp.IdentitySchema = testIdentitySchema
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.UpgradeResourceIdentityRequest{
				RawIdentity: testNewTfprotov5RawState(t, map[string]interface{}{
					"test_id": "test-id-value",
				}),
				TypeName: "test_resource",
				Version:  0,
			},
			expectedResponse: &tfprotov5.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfprotov5.ResourceIdentityData{
					IdentityData: testNewDynamicValue(t, testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "test-id-value"),
					}),
				},
			},
		},
		"request-TypeName-missing": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request: &tfprotov5.UpgradeResourceIdentityRequest{},
			expectedResponse: &tfprotov5.UpgradeResourceIdentityResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
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
			request: &tfprotov5.UpgradeResourceIdentityRequest{
				TypeName: "unknown",
			},
			expectedResponse: &tfprotov5.UpgradeResourceIdentityResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
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
									return &testprovider.ResourceWithUpgradeResourceIdentity{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
											return map[int64]resource.IdentityUpgrader{
												0: {
													IdentityUpgrader: func(_ context.Context, _ resource.UpgradeIdentityRequest, resp *resource.UpgradeIdentityResponse) {
														resp.Diagnostics.AddWarning("warning summary", "warning detail")
														resp.Diagnostics.AddError("error summary", "error detail")
													},
												},
											}
										},
										IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
											resp.IdentitySchema = testIdentitySchema
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.UpgradeResourceIdentityRequest{
				RawIdentity: testNewTfprotov5RawState(t, map[string]interface{}{
					"test_id":            "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_resource",
				Version:  0,
			},
			expectedResponse: &tfprotov5.UpgradeResourceIdentityResponse{
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
			},
		},
		"response-UpgradedIdentity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithUpgradeResourceIdentity{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
											return map[int64]resource.IdentityUpgrader{
												0: {
													IdentityUpgrader: func(_ context.Context, _ resource.UpgradeIdentityRequest, resp *resource.UpgradeIdentityResponse) {
														resp.Identity = &tfsdk.ResourceIdentity{
															Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
																"test_id": tftypes.NewValue(tftypes.String, "test-id-value"),
															}),
															Schema: testIdentitySchema,
														}
													},
												},
											}
										},
										IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
											resp.IdentitySchema = testIdentitySchema
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.UpgradeResourceIdentityRequest{
				RawIdentity: testNewTfprotov5RawState(t, map[string]interface{}{
					"test_id": "test-id-value",
				}),
				TypeName: "test_resource",
				Version:  0,
			},
			expectedResponse: &tfprotov5.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfprotov5.ResourceIdentityData{
					IdentityData: testNewDynamicValue(t, testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "test-id-value"),
					}),
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.UpgradeResourceIdentity(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
