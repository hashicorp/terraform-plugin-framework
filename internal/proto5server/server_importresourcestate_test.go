// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerImportResourceState(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":       tftypes.String,
			"optional": tftypes.String,
			"required": tftypes.String,
		},
	}

	testIdentityType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id":       tftypes.String,
			"test_other_id": tftypes.String,
		},
	}

	testRequestIdentityValue := testNewDynamicValue(t, testIdentityType, map[string]tftypes.Value{
		"test_id":       tftypes.NewValue(tftypes.String, "id-123"),
		"test_other_id": tftypes.NewValue(tftypes.String, nil),
	})

	testImportedResourceIdentityDynamicValue := testNewDynamicValue(t, testIdentityType, map[string]tftypes.Value{
		"test_id":       tftypes.NewValue(tftypes.String, "id-123"),
		"test_other_id": tftypes.NewValue(tftypes.String, "new-value-123"),
	})

	testStateDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, "test-id"),
		"optional": tftypes.NewValue(tftypes.String, nil),
		"required": tftypes.NewValue(tftypes.String, nil),
	})

	testEmptyStateDynamicValue, err := tfprotov5.NewDynamicValue(testType, tftypes.NewValue(testType, nil))
	if err != nil {
		t.Fatalf("unexpected error calling tfprotov5.NewDynamicValue(): %s", err)
	}

	testEmptyPrivateBytes := privatestate.MustMarshalToJson(map[string][]byte{
		privatestate.ImportBeforeReadKey: []byte(`true`),
	})

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"optional": schema.StringAttribute{
				Optional: true,
			},
			"required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"test_other_id": identityschema.StringAttribute{
				OptionalForImport: true,
			},
		},
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.ImportResourceStateRequest
		expectedError    error
		expectedResponse *tfprotov5.ImportResourceStateResponse
	}{
		"request-id": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithImportState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
											if req.ID != "test-id" {
												resp.Diagnostics.AddError("unexpected req.ID value: %s", req.ID)
											}

											resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ImportResourceStateResponse{
				ImportedResources: []*tfprotov5.ImportedResource{
					{
						State:    testStateDynamicValue,
						TypeName: "test_resource",
						Private:  testEmptyPrivateBytes,
					},
				},
			},
		},
		"request-identity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithIdentityAndImportState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
											var identityData struct {
												TestID      types.String `tfsdk:"test_id"`
												TestOtherID types.String `tfsdk:"test_other_id"`
											}

											resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)

											if identityData.TestID.ValueString() != "id-123" {
												resp.Diagnostics.AddError("Unexpected req.Identity", identityData.TestID.ValueString())
											}

											resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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
			request: &tfprotov5.ImportResourceStateRequest{
				TypeName: "test_resource",
				Identity: &tfprotov5.ResourceIdentityData{
					IdentityData: testRequestIdentityValue,
				},
			},
			expectedResponse: &tfprotov5.ImportResourceStateResponse{
				ImportedResources: []*tfprotov5.ImportedResource{
					{
						State:   &testEmptyStateDynamicValue,
						Private: testEmptyPrivateBytes,
						Identity: &tfprotov5.ResourceIdentityData{
							IdentityData: testRequestIdentityValue,
						},
						TypeName: "test_resource",
					},
				},
			},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithImportState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
			request: &tfprotov5.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ImportResourceStateResponse{
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
		"response-importedresources": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithImportState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
											resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ImportResourceStateResponse{
				ImportedResources: []*tfprotov5.ImportedResource{
					{
						Private:  testEmptyPrivateBytes,
						State:    testStateDynamicValue,
						TypeName: "test_resource",
					},
				},
			},
		},
		"response-importedresources-identity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithIdentityAndImportState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
											resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("test_other_id"), types.StringValue("new-value-123"))...)

											resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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
			request: &tfprotov5.ImportResourceStateRequest{
				TypeName: "test_resource",
				Identity: &tfprotov5.ResourceIdentityData{
					IdentityData: testRequestIdentityValue,
				},
			},
			expectedResponse: &tfprotov5.ImportResourceStateResponse{
				ImportedResources: []*tfprotov5.ImportedResource{
					{
						State:   &testEmptyStateDynamicValue,
						Private: testEmptyPrivateBytes,
						Identity: &tfprotov5.ResourceIdentityData{
							IdentityData: testImportedResourceIdentityDynamicValue,
						},
						TypeName: "test_resource",
					},
				},
			},
		},
		"response-importedresources-private": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithImportState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
											diags := resp.Private.SetKey(ctx, "providerKey", []byte(`{"key": "value"}`))

											resp.Diagnostics.Append(diags...)

											resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ImportResourceStateResponse{
				ImportedResources: []*tfprotov5.ImportedResource{
					{
						State:    testStateDynamicValue,
						TypeName: "test_resource",
						Private: privatestate.MustMarshalToJson(map[string][]byte{
							privatestate.ImportBeforeReadKey: []byte(`true`),
							"providerKey":                    []byte(`{"key": "value"}`),
						}),
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ImportResourceState(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
