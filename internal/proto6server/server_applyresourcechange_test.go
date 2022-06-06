package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerApplyResourceChange(t *testing.T) {
	t.Parallel()

	testSchemaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testEmptyDynamicValue, _ := tfprotov6.NewDynamicValue(testSchemaType, tftypes.NewValue(testSchemaType, nil))

	testSchema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test_computed": {
				Computed: true,
				Type:     types.StringType,
			},
			"test_required": {
				Required: true,
				Type:     types.StringType,
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

	testProviderMetaSchema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test_provider_meta_attribute": {
				Optional: true,
				Type:     types.StringType,
			},
		},
	}

	type testProviderMetaData struct {
		TestProviderMetaAttribute types.String `tfsdk:"test_provider_meta_attribute"`
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.ApplyResourceChangeRequest
		expectedError    error
		expectedResponse *tfprotov6.ApplyResourceChangeResponse
	}{
		"create-request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												var data testSchemaData

												resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

												if data.TestRequired.Value != "test-config-value" {
													resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.Value)
												}
											},
											DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
											},
											UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
				// Intentionally empty, Create implementation does not call resp.State.Set()
				NewState: &testEmptyDynamicValue,
			},
		},
		"create-request-plannedstate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												var data testSchemaData

												resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

												if data.TestComputed.Value != "test-plannedstate-value" {
													resp.Diagnostics.AddError("Unexpected req.Plan Value", "Got: "+data.TestComputed.Value)
												}
											},
											DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
											},
											UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
				// Intentionally empty, Create implementation does not call resp.State.Set()
				NewState: &testEmptyDynamicValue,
			},
		},
		"create-request-providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithProviderMeta{
						Provider: &testprovider.Provider{
							GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
								return map[string]tfsdk.ResourceType{
									"test_resource": &testprovider.ResourceType{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return testSchema, nil
										},
										NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
											return &testprovider.Resource{
												CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
													var data testProviderMetaData

													resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

													if data.TestProviderMetaAttribute.Value != "test-provider-meta-value" {
														resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+data.TestProviderMetaAttribute.Value)
													}
												},
												DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
													resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
												},
												UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
													resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
												},
											}, nil
										},
									},
								}, nil
							},
						},
						GetMetaSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return testProviderMetaSchema, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
				// Intentionally empty, Create implementation does not call resp.State.Set()
				NewState: &testEmptyDynamicValue,
			},
		},
		"create-response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												resp.Diagnostics.AddWarning("warning summary", "warning detail")
												resp.Diagnostics.AddError("error summary", "error detail")
											},
											DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
											},
											UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
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
				NewState: &testEmptyDynamicValue,
			},
		},
		"create-response-newstate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												var data testSchemaData

												resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
												resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
											},
											DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
											},
											UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
				}),
			},
		},
		"delete-request-priorstate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
											},
											DeleteMethod: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												var data testSchemaData

												resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

												if data.TestRequired.Value != "test-priorstate-value" {
													resp.Diagnostics.AddError("Unexpected req.State Value", "Got: "+data.TestRequired.Value)
												}
											},
											UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
				PlannedState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
				NewState: &testEmptyDynamicValue,
			},
		},
		"delete-request-providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithProviderMeta{
						Provider: &testprovider.Provider{
							GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
								return map[string]tfsdk.ResourceType{
									"test_resource": &testprovider.ResourceType{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return testSchema, nil
										},
										NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
											return &testprovider.Resource{
												CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
													resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
												},
												DeleteMethod: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
													var data testProviderMetaData

													resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

													if data.TestProviderMetaAttribute.Value != "test-provider-meta-value" {
														resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+data.TestProviderMetaAttribute.Value)
													}
												},
												UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
													resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
												},
											}, nil
										},
									},
								}, nil
							},
						},
						GetMetaSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return testProviderMetaSchema, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
				PlannedState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				ProviderMeta: testProviderMetaValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
				NewState: &testEmptyDynamicValue,
			},
		},
		"delete-response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
											},
											DeleteMethod: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddWarning("warning summary", "warning detail")
												resp.Diagnostics.AddError("error summary", "error detail")
											},
											UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
				PlannedState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
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
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
											},
											DeleteMethod: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												// Intentionally empty, should call resp.State.RemoveResource() automatically.
											},
											UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
				PlannedState: &testEmptyDynamicValue,
				PriorState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
				}),
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
				NewState: &testEmptyDynamicValue,
			},
		},
		"update-request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")

											},
											DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
											},
											UpdateMethod: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												var data testSchemaData

												resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

												if data.TestRequired.Value != "test-new-value" {
													resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.Value)
												}
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
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
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")

											},
											DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
											},
											UpdateMethod: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												var data testSchemaData

												resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

												if data.TestComputed.Value != "test-plannedstate-value" {
													resp.Diagnostics.AddError("Unexpected req.Plan Value", "Got: "+data.TestComputed.Value)
												}
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
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
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
											},
											DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
											},
											UpdateMethod: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												var data testSchemaData

												resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

												if data.TestRequired.Value != "test-old-value" {
													resp.Diagnostics.AddError("Unexpected req.State Value", "Got: "+data.TestRequired.Value)
												}
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
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
					Provider: &testprovider.ProviderWithProviderMeta{
						Provider: &testprovider.Provider{
							GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
								return map[string]tfsdk.ResourceType{
									"test_resource": &testprovider.ResourceType{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return testSchema, nil
										},
										NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
											return &testprovider.Resource{
												CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
													resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
												},
												DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
													resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
												},
												UpdateMethod: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
													var data testProviderMetaData

													resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

													if data.TestProviderMetaAttribute.Value != "test-provider-meta-value" {
														resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+data.TestProviderMetaAttribute.Value)
													}
												},
											}, nil
										},
									},
								}, nil
							},
						},
						GetMetaSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return testProviderMetaSchema, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, nil),
					"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
				}),
			},
		},
		"update-response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
											},
											DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
											},
											UpdateMethod: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												resp.Diagnostics.AddWarning("warning summary", "warning detail")
												resp.Diagnostics.AddError("error summary", "error detail")
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
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
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											CreateMethod: func(_ context.Context, _ tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
											},
											DeleteMethod: func(_ context.Context, _ tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
												resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
											},
											UpdateMethod: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
												var data testSchemaData

												resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
												resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ApplyResourceChangeRequest{
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
			expectedResponse: &tfprotov6.ApplyResourceChangeResponse{
				NewState: testNewDynamicValue(t, testSchemaType, map[string]tftypes.Value{
					"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
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
