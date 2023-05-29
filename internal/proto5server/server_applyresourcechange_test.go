// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerApplyResourceChange(t *testing.T) {
	t.Parallel()

	testSchemaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testEmptyDynamicValue, _ := tfprotov5.NewDynamicValue(testSchemaType, tftypes.NewValue(testSchemaType, nil))

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

	type testSchemaData struct {
		TestComputed types.String `tfsdk:"test_computed"`
		TestRequired types.String `tfsdk:"test_required"`
	}

	testProviderMetaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_provider_meta_attribute": tftypes.String,
		},
	}

	testProviderMetaValue := testNewDynamicValue(t, testProviderMetaType, map[string]tftypes.Value{
		"test_provider_meta_attribute": tftypes.NewValue(tftypes.String, "test-provider-meta-value"),
	})

	testProviderMetaSchema := metaschema.Schema{
		Attributes: map[string]metaschema.Attribute{
			"test_provider_meta_attribute": metaschema.StringAttribute{
				Optional: true,
			},
		},
	}

	type testProviderMetaData struct {
		TestProviderMetaAttribute types.String `tfsdk:"test_provider_meta_attribute"`
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.ApplyResourceChangeRequest
		expectedError    error
		expectedResponse *tfprotov5.ApplyResourceChangeResponse
	}{
		"create-request-config": {
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
										CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

											if data.TestRequired.ValueString() != "test-config-value" {
												resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.ValueString())
											}

											// Prevent missing resource state error diagnostic
											resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
			},
		},
		"create-request-plannedstate": {
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
										CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

											if data.TestComputed.ValueString() != "test-plannedstate-value" {
												resp.Diagnostics.AddError("Unexpected req.Plan Value", "Got: "+data.TestComputed.ValueString())
											}

											// Prevent missing resource state error diagnostic
											resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
			},
		},
		"create-request-providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithMetaSchema{
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
											CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
												var metadata testProviderMetaData

												resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &metadata)...)

												if metadata.TestProviderMetaAttribute.ValueString() != "test-provider-meta-value" {
													resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+metadata.TestProviderMetaAttribute.ValueString())
												}

												// Prevent missing resource state error diagnostic
												var data testSchemaData

												resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
												resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
											},
											DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
											},
											UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
											},
										}
									},
								}
							},
						},
						MetaSchemaMethod: func(_ context.Context, _ provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
							resp.Schema = testProviderMetaSchema
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState:   &testEmptyDynamicValue,
				ProviderMeta: testProviderMetaValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
			},
		},
		"create-response-diagnostics": {
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
										CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddWarning("warning summary", "warning detail")
											resp.Diagnostics.AddError("error summary", "error detail")
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
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
				NewState: &testEmptyDynamicValue,
			},
		},
		"create-response-newstate": {
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
										CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
											resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
			},
		},
		"create-response-newstate-null": {
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
										CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
											// Intentionally missing resp.State.Set()
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Missing Resource State After Create",
						Detail: "The Terraform Provider unexpectedly returned no resource state after having no errors in the resource creation. " +
							"This is always an issue in the Terraform Provider and should be reported to the provider developers.\n\n" +
							"The resource may have been successfully created, but Terraform is not tracking it. " +
							"Applying the configuration again with no other action may result in duplicate resource errors.",
					},
				},
				NewState: &testEmptyDynamicValue,
			},
		},
		"create-response-private": {
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
										CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
											var data testSchemaData

											// Prevent missing resource state error diagnostic
											resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

											diags := resp.Private.SetKey(ctx, "providerKey", []byte(`{"key": "value"}`))

											resp.Diagnostics.Append(diags...)
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, nil),
				}),
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					"providerKey": []byte(`{"key": "value"}`),
				}),
			},
		},
		"delete-request-priorstate": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
										},
										DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

											if data.TestRequired.ValueString() != "test-priorstate-value" {
												resp.Diagnostics.AddError("Unexpected req.State Value", "Got: "+data.TestRequired.ValueString())
											}
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				PlannedState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: &testEmptyDynamicValue,
			},
		},
		"delete-request-providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithMetaSchema{
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
											CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
											},
											DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
												var data testProviderMetaData

												resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

												if data.TestProviderMetaAttribute.ValueString() != "test-provider-meta-value" {
													resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+data.TestProviderMetaAttribute.ValueString())
												}
											},
											UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
											},
										}
									},
								}
							},
						},
						MetaSchemaMethod: func(_ context.Context, _ provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
							resp.Schema = testProviderMetaSchema
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				PlannedState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				ProviderMeta: testProviderMetaValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: &testEmptyDynamicValue,
			},
		},
		"delete-request-private": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
										},
										DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
											expected := `{"key": "value"}`
											got, diags := req.Private.GetKey(ctx, "providerKey")

											resp.Diagnostics.Append(diags...)

											if string(got) != expected {
												resp.Diagnostics.AddError(
													"Unexpected req.Private Value",
													fmt.Sprintf("expected %q, got %q", expected, got),
												)
											}
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
				PlannedPrivate: privatestate.MustMarshalToJson(map[string][]byte{
					"providerKey": []byte(`{"key": "value"}`),
				}),
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: &testEmptyDynamicValue,
			},
		},
		"delete-response-diagnostics": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
										},
										DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddWarning("warning summary", "warning detail")
											resp.Diagnostics.AddError("error summary", "error detail")
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				PlannedState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
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
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
			},
		},
		"delete-response-newstate": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
										},
										DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
											// Intentionally empty, should call resp.State.RemoveResource() automatically.
										},
										UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				PlannedState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: &testEmptyDynamicValue,
			},
		},
		"update-request-config": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")

										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
										},
										UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

											if data.TestRequired.ValueString() != "test-new-value" {
												resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.ValueString())
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
			},
		},
		"update-request-plannedstate": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")

										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
										},
										UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

											if data.TestComputed.ValueString() != "test-plannedstate-value" {
												resp.Diagnostics.AddError("Unexpected req.Plan Value", "Got: "+data.TestComputed.ValueString())
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
			},
		},
		"update-request-priorstate": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
										},
										UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

											if data.TestRequired.ValueString() != "test-old-value" {
												resp.Diagnostics.AddError("Unexpected req.State Value", "Got: "+data.TestRequired.ValueString())
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
			},
		},
		"update-request-providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithMetaSchema{
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
											CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
											},
											DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
											},
											UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
												var data testProviderMetaData

												resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

												if data.TestProviderMetaAttribute.ValueString() != "test-provider-meta-value" {
													resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+data.TestProviderMetaAttribute.ValueString())
												}
											},
										}
									},
								}
							},
						},
						MetaSchemaMethod: func(_ context.Context, _ provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
							resp.Schema = testProviderMetaSchema
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				ProviderMeta: testProviderMetaValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
			},
		},
		"update-request-private": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithMetaSchema{
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
											CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
											},
											DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
											},
											UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
												expected := `{"providerKey": "provider value"}`
												got, diags := req.Private.GetKey(ctx, "providerKey")

												resp.Diagnostics.Append(diags...)

												if string(got) != expected {
													resp.Diagnostics.AddError(
														"Unexpected req.Private Value",
														fmt.Sprintf("expected %q, got %q", expected, got),
													)
												}
											},
										}
									},
								}
							},
						},
						MetaSchemaMethod: func(_ context.Context, _ provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
							resp.Schema = testProviderMetaSchema
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				ProviderMeta: testProviderMetaValue,
				TypeName:     "test_resource",
				PlannedPrivate: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey": []byte(`{"frameworkKey": "framework value"}`),
					"providerKey":   []byte(`{"providerKey": "provider value"}`),
				}),
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey": []byte(`{"frameworkKey": "framework value"}`),
					"providerKey":   []byte(`{"providerKey": "provider value"}`),
				}),
			},
		},
		"update-response-diagnostics": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
										},
										UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
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
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
			},
		},
		"update-response-newstate": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
										},
										UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
											resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
			},
		},
		"update-response-newstate-null": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
										},
										UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
											resp.State.RemoveResource(ctx)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Missing Resource State After Update",
						Detail: "The Terraform Provider unexpectedly returned no resource state after having no errors in the resource update. " +
							"This is always an issue in the Terraform Provider and should be reported to the provider developers.",
					},
				},
				NewState: &testEmptyDynamicValue,
			},
		},
		"update-response-private": {
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
										CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
										},
										DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
											resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
										},
										UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
											diags := resp.Private.SetKey(ctx, "providerKey", []byte(`{"providerKey": "provider value"}`))

											resp.Diagnostics.Append(diags...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ApplyResourceChangeRequest{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
				PlannedPrivate: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey": []byte(`{"frameworkKey": "framework value"}`),
				}),
			},
			expectedResponse: &tfprotov5.ApplyResourceChangeResponse{
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey": []byte(`{"frameworkKey": "framework value"}`),
					"providerKey":   []byte(`{"providerKey": "provider value"}`),
				}),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ApplyResourceChange(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
