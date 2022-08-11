package fwserver_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerDeleteResource(t *testing.T) {
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

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.DeleteResourceRequest
		expectedResponse *fwserver.DeleteResourceResponse
	}{
		"request-priorstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.DeleteResourceRequest{
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
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
								var data testSchemaData

								resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

								if data.TestRequired.Value != "test-priorstate-value" {
									resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestRequired.Value)
								}
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.DeleteResourceResponse{
				NewState: testEmptyState,
			},
		},
		"request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.DeleteResourceRequest{
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
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
								var data testProviderMetaData

								resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

								if data.TestProviderMetaAttribute.Value != "test-provider-meta-value" {
									resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", data.TestProviderMetaAttribute.Value)
								}
							},
						}, nil
					},
				},
				ProviderMeta: testProviderMetaConfig,
			},
			expectedResponse: &fwserver.DeleteResourceResponse{
				NewState: testEmptyState,
			},
		},
		"request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.DeleteResourceRequest{
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
								expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`
								got, diags := req.Private.GetKey(ctx, "providerKeyOne")

								resp.Diagnostics.Append(diags...)

								if string(got) != expected {
									resp.Diagnostics.AddError(
										"Unexpected req.Private Value",
										fmt.Sprintf("expected %q, got %q", expected, got),
									)
								}
							},
						}, nil
					},
				},
				PlannedPrivate: &privatestate.Data{
					Provider: testProviderData,
				},
			},
			expectedResponse: &fwserver.DeleteResourceResponse{
				NewState: testEmptyState,
			},
		},
		"request-private-planned-private-nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.DeleteResourceRequest{
				ResourceSchema: testSchema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
								var expected []byte

								got, diags := req.Private.GetKey(ctx, "providerKeyOne")

								resp.Diagnostics.Append(diags...)

								if !bytes.Equal(got, expected) {
									resp.Diagnostics.AddError(
										"Unexpected req.Private Value",
										fmt.Sprintf("expected %q, got %q", expected, got),
									)
								}
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.DeleteResourceResponse{
				NewState: testEmptyState,
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.DeleteResourceRequest{
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
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
								resp.Diagnostics.AddWarning("warning summary", "warning detail")
								resp.Diagnostics.AddError("error summary", "error detail")
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.DeleteResourceResponse{
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
		"response-newstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.DeleteResourceRequest{
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
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
								// Intentionally empty, should call resp.State.RemoveResource() automatically.
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.DeleteResourceResponse{
				NewState: testEmptyState,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.DeleteResourceResponse{}
			testCase.server.DeleteResource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
