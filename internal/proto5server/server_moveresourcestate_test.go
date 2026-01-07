// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerMoveResourceState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
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
	}
	schemaType := testSchema.Type().TerraformType(ctx)

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}

	testIdentityType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id": tftypes.String,
		},
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.MoveResourceStateRequest
		expectedResponse *tfprotov5.MoveResourceStateResponse
		expectedError    error
	}{
		"nil": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request:          nil,
			expectedResponse: &tfprotov5.MoveResourceStateResponse{},
		},
		"request-SourcePrivate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`
														got, diags := req.SourcePrivate.GetKey(ctx, "providerKey")

														resp.Diagnostics.Append(diags...)

														if string(got) != expected {
															resp.Diagnostics.AddError(
																"Unexpected req.SourcePrivate Value",
																fmt.Sprintf("expected %q, got %q", expected, got),
															)
														}

														// Prevent missing implementation error, the values do not matter except for response assertion
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				SourcePrivate: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKey":   []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
				// SourceState required to prevent error
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				// TargetPrivate intentionally not set by the framework
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-SourceProviderAddress": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														expected := `example.com/namespace/type`

														if diff := cmp.Diff(req.SourceProviderAddress, expected); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceProviderAddress difference", diff)
														}

														// Prevent missing implementation error, the values do not matter except for response assertion
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				SourceProviderAddress: "example.com/namespace/type",
				// SourceState required to prevent error
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-SourceSchemaVersion": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														expected := int64(123)

														if diff := cmp.Diff(req.SourceSchemaVersion, expected); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceSchemaVersion difference", diff)
														}

														// Prevent missing implementation error, the values do not matter except for response assertion
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				SourceSchemaVersion: 123,
				// SourceState required to prevent error
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-SourceRawState": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														expectedSourceRawState := testNewTfprotov6RawState(t, map[string]interface{}{
															"id":                 "test-id-value",
															"required_attribute": true,
														})

														if diff := cmp.Diff(req.SourceRawState, expectedSourceRawState); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceRawState difference", diff)
														}

														// Prevent missing implementation error, the values do not matter except for response assertion
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-SourceState": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													SourceSchema: &testSchema,
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														var id, requiredAttribute types.String

														resp.Diagnostics.Append(req.SourceState.GetAttribute(ctx, path.Root("id"), &id)...)
														resp.Diagnostics.Append(req.SourceState.GetAttribute(ctx, path.Root("required_attribute"), &requiredAttribute)...)

														if diff := cmp.Diff(id, types.StringValue("test-id-value")); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceState difference", diff)
														}

														if diff := cmp.Diff(requiredAttribute, types.StringValue("true")); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceState difference", diff)
														}

														// Prevent missing implementation error, the values do not matter except for response assertion
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-SourceTypeName": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														expected := `test_source_resource`

														if diff := cmp.Diff(req.SourceTypeName, expected); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceTypeName difference", diff)
														}

														// Prevent missing implementation error, the values do not matter except for response assertion
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				// SourceState required to prevent error
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				SourceTypeName: "test_source_resource",
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-SourceIdentitySchemaVersion": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														expected := int64(123)

														if diff := cmp.Diff(req.SourceIdentitySchemaVersion, expected); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceIdentitySchemaVersion difference", diff)
														}

														// Prevent missing implementation error, the values do not matter except for response assertion
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				SourceIdentitySchemaVersion: 123,
				// SourceState required to prevent error
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-SourceIdentity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithIdentityAndMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														expectedSourceRawState := testNewTfprotov6RawState(t, map[string]interface{}{
															"id":                 "test-id-value",
															"required_attribute": true,
														})

														if diff := cmp.Diff(req.SourceRawState, expectedSourceRawState); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceRawState difference", diff)
														}

														expectedSourceIdentity := testNewTfprotov6RawState(t, map[string]interface{}{
															"test_id": "test-id-value",
														})

														if diff := cmp.Diff(req.SourceIdentity, expectedSourceIdentity); diff != "" {
															resp.Diagnostics.AddError("Unexpected req.SourceIdentity difference", diff)
														}

														resp.Diagnostics.Append(resp.TargetIdentity.SetAttribute(ctx, path.Root("test_id"), "test-id-value")...)

														// Prevent missing implementation error, the values do not matter except for response assertion
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				SourceIdentity: testNewTfprotov5RawState(t, map[string]interface{}{
					"test_id": "test-id-value",
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetIdentity: &tfprotov5.ResourceIdentityData{IdentityData: testNewDynamicValue(t, testIdentityType, map[string]tftypes.Value{
					"test_id": tftypes.NewValue(tftypes.String, "test-id-value"),
				})},
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-TargetTypeName-missing": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request: &tfprotov5.MoveResourceStateRequest{},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Resource Type Not Found",
						Detail:   "No resource type named \"\" was found in the provider.",
					},
				},
			},
		},
		"request-TargetTypeName-unknown": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request: &tfprotov5.MoveResourceStateRequest{
				TargetTypeName: "unknown",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Resource Type Not Found",
						Detail:   "No resource type named \"unknown\" was found in the provider.",
					},
				},
			},
		},
		"request-TargetTypeName-unimplemented-interface": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.MoveResourceStateRequest{
				SourceProviderAddress: "example.com/namespace/type",
				// SourceState required to prevent error
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				SourceTypeName: "test_source_resource",
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Unable to Move Resource State",
						Detail: "The target resource implementation does not include move resource state support. " +
							"The resource implementation can be updated by the provider developers to include this support with the ResourceWithMoveState interface.\n\n" +
							"Source Provider Address: example.com/namespace/type\n" +
							"Source Resource Type: test_source_resource\n" +
							"Source Resource Schema Version: 0\n" +
							"Target Resource Type: test_resource",
					},
				},
			},
		},
		"request-TargetTypeName-unimplemented-no-responses": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.MoveResourceStateRequest{
				SourceProviderAddress: "example.com/namespace/type",
				// SourceState required to prevent error
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				SourceTypeName: "test_source_resource",
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Unable to Move Resource State",
						Detail: "The target resource implementation does not include support for the given source resource. " +
							"The resource implementation can be updated by the provider developers to include this support by returning the moved state when the request matches this source.\n\n" +
							"Source Provider Address: example.com/namespace/type\n" +
							"Source Resource Type: test_source_resource\n" +
							"Source Resource Schema Version: 0\n" +
							"Target Resource Type: test_resource",
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
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
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
			request: &tfprotov5.MoveResourceStateRequest{
				// SourceState required to prevent error
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
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
		"response-TargetPrivate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														resp.Diagnostics.Append(resp.TargetPrivate.SetKey(ctx, "providerKey", []byte(`{"key": "value"}`))...)

														// Prevent missing implementation error, the values do not matter except for response assertion
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				// SourceState required to prevent error
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetPrivate: privatestate.MustMarshalToJson(map[string][]byte{
					"providerKey": []byte(`{"key": "value"}`),
				}),
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"response-TargetState": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"response-TargetIdentity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithIdentityAndMoveState{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										MoveStateMethod: func(ctx context.Context) []resource.StateMover {
											return []resource.StateMover{
												{
													StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
														resp.Diagnostics.Append(resp.TargetIdentity.SetAttribute(ctx, path.Root("test_id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
														resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
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
			request: &tfprotov5.MoveResourceStateRequest{
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				SourceIdentity: testNewTfprotov5RawState(t, map[string]interface{}{
					"test_id": "test-id-value",
				}),
				TargetTypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.MoveResourceStateResponse{
				TargetState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
				TargetIdentity: &tfprotov5.ResourceIdentityData{IdentityData: testNewDynamicValue(t, testIdentityType, map[string]tftypes.Value{
					"test_id": tftypes.NewValue(tftypes.String, "test-id-value"),
				})},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.MoveResourceState(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
