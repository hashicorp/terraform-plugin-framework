package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	testEmptyPlan := &tfsdk.Plan{
		Raw:    tftypes.NewValue(testSchemaType, nil),
		Schema: testSchema,
	}

	testEmptyState := &tfsdk.State{
		Raw:    tftypes.NewValue(testSchemaType, nil),
		Schema: testSchema,
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

	testProviderMetaValue := tftypes.NewValue(testProviderMetaType, map[string]tftypes.Value{
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

	testProviderMetaConfig := &tfsdk.Config{
		Raw:    testProviderMetaValue,
		Schema: testProviderMetaSchema,
	}

	type testProviderMetaData struct {
		TestProviderMetaAttribute types.String `tfsdk:"test_provider_meta_attribute"`
	}

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.ApplyResourceChangeRequest
		expectedResponse *fwserver.ApplyResourceChangeResponse
	}{
		"create-request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
								var data testSchemaData

								resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

								if data.TestRequired.Value != "test-config-value" {
									resp.Diagnostics.AddError("unexpected req.Config value: %s", data.TestRequired.Value)
								}

								// Prevent missing resource state error diagnostic
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"create-request-plannedstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
								var data testSchemaData

								resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

								if data.TestRequired.Value != "test-plannedstate-value" {
									resp.Diagnostics.AddError("unexpected req.Plan value: %s", data.TestRequired.Value)
								}

								// Prevent missing resource state error diagnostic
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"create-request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
								var metadata testProviderMetaData

								resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &metadata)...)

								if metadata.TestProviderMetaAttribute.Value != "test-provider-meta-value" {
									resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+metadata.TestProviderMetaAttribute.Value)
								}

								// Prevent missing resource state error diagnostic
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
				ProviderMeta: testProviderMetaConfig,
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"create-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
				// Intentionally empty, Create implementation does not call resp.State.Set()
				NewState: testEmptyState,
			},
		},
		"create-response-newstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"create-response-newstate-null": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
								// Intentionally missing resp.State.Set()
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Missing Resource State After Create",
						"The Terraform Provider unexpectedly returned no resource state after having no errors in the resource creation. "+
							"This is always an issue in the Terraform Provider and should be reported to the provider developers.\n\n"+
							"The resource may have been successfully created, but Terraform is not tracking it. "+
							"Applying the configuration again with no other action may result in duplicate resource errors.",
					),
				},
				NewState: testEmptyState,
			},
		},
		"delete-request-priorstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
									resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestRequired.Value)
								}
							},
							UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
								resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: testEmptyState,
			},
		},
		"delete-request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
									resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", data.TestProviderMetaAttribute.Value)
								}
							},
							UpdateMethod: func(_ context.Context, _ tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
								resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
							},
						}, nil
					},
				},
				ProviderMeta: testProviderMetaConfig,
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: testEmptyState,
			},
		},
		"delete-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"delete-response-newstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: testEmptyState,
			},
		},
		"update-request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"update-request-plannedstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"update-request-priorstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"update-request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ProviderMeta:   testProviderMetaConfig,
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"update-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"update-response-newstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"update-response-newstate-null": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
								resp.State.RemoveResource(ctx)
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Missing Resource State After Update",
						"The Terraform Provider unexpectedly returned no resource state after having no errors in the resource update. "+
							"This is always an issue in the Terraform Provider and should be reported to the provider developers.",
					),
				},
				NewState: testEmptyState,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ApplyResourceChangeResponse{}
			testCase.server.ApplyResourceChange(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
