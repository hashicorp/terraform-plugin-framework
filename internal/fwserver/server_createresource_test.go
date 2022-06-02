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

func TestServerCreateResource(t *testing.T) {
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
		request          *fwserver.CreateResourceRequest
		expectedResponse *fwserver.CreateResourceResponse
	}{
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
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
							CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
								var data testSchemaData

								resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

								if data.TestRequired.Value != "test-config-value" {
									resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.Value)
								}
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				// Intentionally empty, Create implementation does not call resp.State.Set()
				NewState: testEmptyState,
			},
		},
		"request-plannedstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
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
							CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
								var data testSchemaData

								resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

								if data.TestRequired.Value != "test-plannedstate-value" {
									resp.Diagnostics.AddError("Unexpected req.Plan Value", "Got: "+data.TestRequired.Value)
								}
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				// Intentionally empty, Create implementation does not call resp.State.Set()
				NewState: testEmptyState,
			},
		},
		"request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
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
						}, nil
					},
				},
				ProviderMeta: testProviderMetaConfig,
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				// Intentionally empty, Create implementation does not call resp.State.Set()
				NewState: testEmptyState,
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
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
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
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
		"response-newstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
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
							CreateMethod: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
								var data testSchemaData

								resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
								resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.CreateResourceResponse{}
			testCase.server.CreateResource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
