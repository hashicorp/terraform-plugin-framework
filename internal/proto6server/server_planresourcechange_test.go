// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerPlanResourceChange(t *testing.T) {
	t.Parallel()

	testSchemaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testEmptyDynamicValue, _ := tfprotov6.NewDynamicValue(testSchemaType, tftypes.NewValue(testSchemaType, nil))

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
		request          *tfprotov6.PlanResourceChangeRequest
		expectedError    error
		expectedResponse *tfprotov6.PlanResourceChangeResponse
	}{
		"create-request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

											if data.TestRequired.ValueString() != "test-config-value" {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

											if !data.TestComputed.IsUnknown() {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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
										return &testprovider.ResourceWithModifyPlan{
											Resource: &testprovider.Resource{
												SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
													resp.Schema = testSchema
												},
												MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
													resp.TypeName = "test_resource"
												},
											},
											ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState:   &testEmptyDynamicValue,
				ProviderMeta: testProviderMetaValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
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
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
			},
		},
		"create-response-plannedstate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

											data.TestComputed = types.StringValue("test-plannedstate-value")

											resp.Diagnostics.Append(resp.Plan.Set(ctx, &data)...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
			},
		},
		"create-response-requiresreplace": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											// This is a strange thing to signal on creation,
											// but the framework does not prevent you from
											// doing it and it might be overly burdensome on
											// provider developers to have the framework raise
											// an error if it is technically valid in the
											// protocol.
											resp.RequiresReplace = path.Paths{
												path.Root("test_required"),
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				PriorState: &testEmptyDynamicValue,
				TypeName:   "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("test_required"),
				},
			},
		},
		"delete-request-priorstate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

											if data.TestRequired.ValueString() != "test-priorstate-value" {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				ProposedNewState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: &testEmptyDynamicValue,
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
										return &testprovider.ResourceWithModifyPlan{
											Resource: &testprovider.Resource{
												SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
													resp.Schema = testSchema
												},
												MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
													resp.TypeName = "test_resource"
												},
											},
											ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				ProposedNewState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				ProviderMeta: testProviderMetaValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: &testEmptyDynamicValue,
			},
		},
		"delete-response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				ProposedNewState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
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
				PlannedState: &testEmptyDynamicValue,
			},
		},
		"delete-response-plannedstate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											// This is invalid logic to run during deletion.
											resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("test_computed"), types.StringValue("test-plannedstate-value"))...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.PlanResourceChangeRequest{
				ProposedNewState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Unexpected Planned Resource State on Destroy",
						Detail: "The Terraform Provider unexpectedly returned resource state data when the resource was planned for destruction. " +
							"This is always an issue in the Terraform Provider and should be reported to the provider developers.\n\n" +
							"Ensure all resource plan modifiers do not attempt to change resource plan data from being a null value if the request plan is a null value.",
					},
				},
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		},
		"delete-response-requiresreplace": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											// This is a strange thing to signal on destroy,
											// but the framework does not prevent you from
											// doing it and it might be overly burdensome on
											// provider developers to have the framework raise
											// an error if it is technically valid in the
											// protocol.
											resp.RequiresReplace = path.Paths{
												path.Root("test_required"),
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.PlanResourceChangeRequest{
				ProposedNewState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: &testEmptyDynamicValue,
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("test_required"),
				},
			},
		},
		"update-request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
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
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

											if !data.TestComputed.IsUnknown() {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
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
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
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
										return &testprovider.ResourceWithModifyPlan{
											Resource: &testprovider.Resource{
												SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
													resp.Schema = testSchema
												},
												MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
													resp.TypeName = "test_resource"
												},
											},
											ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				ProviderMeta: testProviderMetaValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
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
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
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
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
			},
		},
		"update-response-plannedstate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											var data testSchemaData

											resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

											data.TestComputed = types.StringValue("test-plannedstate-value")

											resp.Diagnostics.Append(resp.Plan.Set(ctx, &data)...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
			},
		},
		"update-response-requiresreplace": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithModifyPlan{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
											resp.RequiresReplace = path.Paths{
												path.Root("test_required"),
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.PlanResourceChangeRequest{
				Config: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				ProposedNewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.PlanResourceChangeResponse{
				PlannedState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
				}),
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("test_required"),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.PlanResourceChange(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
