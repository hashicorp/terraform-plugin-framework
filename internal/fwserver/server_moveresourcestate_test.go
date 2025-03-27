// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"optionalforimport_attribute": identityschema.StringAttribute{
				OptionalForImport: true,
			},
			"requiredforimport_attribute": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
	schemaType := testSchema.Type().TerraformType(ctx)

	schemaIdentityType := testIdentitySchema.Type().TerraformType(ctx)

	testSchemaWriteOnly := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"write_only_attribute": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"required_attribute": schema.StringAttribute{
				Required: true,
			},
		},
	}
	schemaTypeWriteOnly := testSchemaWriteOnly.Type().TerraformType(ctx)

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.MoveResourceStateRequest
		expectedResponse *fwserver.MoveResourceStateResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request:          nil,
			expectedResponse: &fwserver.MoveResourceStateResponse{},
		},
		"request-SourcePrivate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourcePrivate: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					},
					Provider: privatestate.MustProviderData(ctx, privatestate.MustMarshalToJson(map[string][]byte{
						"providerKey": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
					})),
				},
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				// TargetPrivate intentionally not set by the framework
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"request-SourceProviderAddress": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceProviderAddress: "example.com/namespace/type",
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"request-SourceSchemaVersion": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceSchemaVersion: 123,
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"request-SourceRawState": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
					MoveStateMethod: func(ctx context.Context) []resource.StateMover {
						return []resource.StateMover{
							{
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									expectedSourceRawState := testNewRawState(t, map[string]interface{}{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"request-SourceState": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"request-SourceState-conversion-errors-ignored": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
					MoveStateMethod: func(ctx context.Context) []resource.StateMover {
						return []resource.StateMover{
							{
								// Intentionally invalid SourceSchema to cause conversion errors
								SourceSchema: &schema.Schema{
									Attributes: map[string]schema.Attribute{
										"id": schema.BoolAttribute{
											Computed: true,
										},
									},
								},
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									if req.SourceState != nil {
										resp.Diagnostics.AddError("Unexpected req.SourceState", "expected nil, got non-nil")
									}
								},
							},
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"request-SourceState-IgnoreUndefinedAttributes": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
					MoveStateMethod: func(ctx context.Context) []resource.StateMover {
						return []resource.StateMover{
							{
								// Intentionally different SourceSchema to cause null state
								SourceSchema: &schema.Schema{
									Attributes: map[string]schema.Attribute{
										"nonexistent": schema.BoolAttribute{
											Computed: true,
										},
									},
								},
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									// Intentionally populated SourceState due to IgnoreUndefinedAttributes
									if req.SourceState == nil {
										resp.Diagnostics.AddError("Expected req.SourceState", "expected non-nil, got nil")
									}
								},
							},
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"request-SourceTypeName": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				SourceTypeName: "test_source_resource",
				TargetResource: &testprovider.ResourceWithMoveState{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"request-TargetTypeName-unimplemented-interface": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceProviderAddress: "example.com/namespace/type",
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				SourceTypeName:       "test_source_resource",
				TargetResource:       &testprovider.Resource{},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Move Resource State",
						"The target resource implementation does not include move resource state support. "+
							"The resource implementation can be updated by the provider developers to include this support with the ResourceWithMoveState interface.\n\n"+
							"Source Provider Address: example.com/namespace/type\n"+
							"Source Resource Type: test_source_resource\n"+
							"Source Resource Schema Version: 0\n"+
							"Target Resource Type: test_resource",
					),
				},
			},
		},
		"request-TargetTypeName-unimplemented-no-responses": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceProviderAddress: "example.com/namespace/type",
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				SourceTypeName: "test_source_resource",
				TargetResource: &testprovider.ResourceWithMoveState{
					MoveStateMethod: func(ctx context.Context) []resource.StateMover {
						return []resource.StateMover{}
					},
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Move Resource State",
						"The target resource implementation does not include support for the given source resource. "+
							"The resource implementation can be updated by the provider developers to include this support by returning the moved state when the request matches this source.\n\n"+
							"Source Provider Address: example.com/namespace/type\n"+
							"Source Resource Type: test_source_resource\n"+
							"Source Resource Schema Version: 0\n"+
							"Target Resource Type: test_resource",
					),
				},
				TargetPrivate: privatestate.EmptyData(ctx),
			},
		},
		"response-Diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
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
				TargetPrivate: privatestate.EmptyData(ctx),
			},
		},
		"response-Diagnostics-first-error-always-responds": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
					MoveStateMethod: func(ctx context.Context) []resource.StateMover {
						return []resource.StateMover{
							{
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									// This intentionally should not be included in the response.
									// The error in the next StateMover should always be the response.
									resp.Diagnostics.AddWarning("warning summary 1", "warning detail 1")
								},
							},
							{
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									resp.Diagnostics.AddError("error summary 2", "error detail 2")
								},
							},
							{
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									// These intentionally should not be included in the response.
									// The error in the prior StateMover should always be the response.
									resp.Diagnostics.AddWarning("warning summary 3", "warning detail 3")
									resp.Diagnostics.AddError("error summary 3", "error detail 3")
								},
							},
						}
					},
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"error summary 2",
						"error detail 2",
					),
				},
				TargetPrivate: privatestate.EmptyData(ctx),
			},
		},
		"response-TargetPrivate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: &privatestate.Data{
					Provider: privatestate.MustProviderData(ctx, privatestate.MustMarshalToJson(map[string][]byte{
						"providerKey": []byte(`{"key": "value"}`),
					})),
				},
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"response-TargetState": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"response-TargetState-first-state-responds": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
					MoveStateMethod: func(ctx context.Context) []resource.StateMover {
						return []resource.StateMover{
							{
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									// Intentionally empty TargetState as below StateMover should respond.
								},
							},
							{
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value-2")...)
									resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
								},
							},
							{
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									// Intentionally different TargetState as above StateMover should respond.
									resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value-3")...)
									resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "false")...)
								},
							},
						}
					},
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value-2"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"response-TargetState-write-only-nullification": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                   "test-id-value",
					"write_only_attribute": nil,
					"required_attribute":   true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
					MoveStateMethod: func(ctx context.Context) []resource.StateMover {
						return []resource.StateMover{
							{
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
									resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("write_only_attribute"), "movestate-val")...)
									resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
								},
							},
						}
					},
				},
				TargetResourceSchema: testSchemaWriteOnly,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaTypeWriteOnly, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "test-id-value"),
						"write_only_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute":   tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchemaWriteOnly,
				},
			},
		},
		"request-SourceIdentitySchemaVersion": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				SourceIdentitySchemaVersion: 123,
				// SourceRawState required to prevent error
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
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
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"request-SourceIdentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.MoveResourceStateRequest{
				// SourceRawState required to prevent error
				SourceIdentity: testNewRawState(t, map[string]interface{}{
					"optionalforimport_attribute": false,
					"requiredforimport_attribute": true,
				}),
				SourceRawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TargetResource: &testprovider.ResourceWithMoveState{
					MoveStateMethod: func(ctx context.Context) []resource.StateMover {
						return []resource.StateMover{
							{
								StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
									expectedSourceIdentity := testNewRawState(t, map[string]interface{}{
										"optionalforimport_attribute": false,
										"requiredforimport_attribute": true,
									})

									if diff := cmp.Diff(req.SourceIdentity, expectedSourceIdentity); diff != "" {
										resp.Diagnostics.AddError("Unexpected req.SourceIdentity difference", diff)
									}

									resp.Diagnostics.Append(resp.TargetIdentity.SetAttribute(ctx, path.Root("optionalforimport_attribute"), "false")...)
									resp.Diagnostics.Append(resp.TargetIdentity.SetAttribute(ctx, path.Root("requiredforimport_attribute"), "true")...)

									// Prevent missing implementation error, the values do not matter except for response assertion
									resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("id"), "test-id-value")...)
									resp.Diagnostics.Append(resp.TargetState.SetAttribute(ctx, path.Root("required_attribute"), "true")...)
								},
							},
						}
					},
				},
				TargetResourceSchema: testSchema,
				TargetTypeName:       "test_resource",
			},
			expectedResponse: &fwserver.MoveResourceStateResponse{
				TargetPrivate: privatestate.EmptyData(ctx),
				TargetState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
				TargetIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(schemaIdentityType, map[string]tftypes.Value{
						"optionalforimport_attribute": tftypes.NewValue(tftypes.String, "false"),
						"requiredforimport_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testIdentitySchema,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.MoveResourceStateResponse{}

			testCase.server.MoveResourceState(context.Background(), testCase.request, response)

			if diff := cmp.Diff(testCase.expectedResponse, response, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
