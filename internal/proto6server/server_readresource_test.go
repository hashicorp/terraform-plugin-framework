package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerReadResource(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testCurrentStateValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, nil),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testEmptyDynamicValue := testNewDynamicValue(t, tftypes.Object{}, nil)

	testNewStateDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-newstate-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testNewStateRemovedDynamicValue, _ := tfprotov6.NewDynamicValue(testType, tftypes.NewValue(testType, nil))

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

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.ReadResourceRequest
		expectedError    error
		expectedResponse *tfprotov6.ReadResourceResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
							return map[string]provider.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return tfsdk.Schema{}, nil
									},
									NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
										return &testprovider.Resource{}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testEmptyDynamicValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testEmptyDynamicValue,
			},
		},
		"request-currentstate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
							return map[string]provider.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
												var data struct {
													TestComputed types.String `tfsdk:"test_computed"`
													TestRequired types.String `tfsdk:"test_required"`
												}

												resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

												if data.TestRequired.Value != "test-currentstate-value" {
													resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestRequired.Value)
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
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testCurrentStateValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testCurrentStateValue,
			},
		},
		"request-providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithMetaSchema{
						Provider: &testprovider.Provider{
							GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
								return map[string]provider.ResourceType{
									"test_resource": &testprovider.ResourceType{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{}, nil
										},
										NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
											return &testprovider.Resource{
												ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
													var data struct {
														TestComputed types.String `tfsdk:"test_computed"`
														TestRequired types.String `tfsdk:"test_required"`
													}

													resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

													if data.TestRequired.Value != "test-currentstate-value" {
														resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", data.TestRequired.Value)
													}
												},
											}, nil
										},
									},
								}, nil
							},
						},
						GetMetaSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return testSchema, nil
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testEmptyDynamicValue,
				ProviderMeta: testCurrentStateValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testEmptyDynamicValue,
			},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
							return map[string]provider.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testCurrentStateValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
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
				NewState: testCurrentStateValue,
			},
		},
		"response-state": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
							return map[string]provider.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
												var data struct {
													TestComputed types.String `tfsdk:"test_computed"`
													TestRequired types.String `tfsdk:"test_required"`
												}

												resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

												data.TestComputed = types.String{Value: "test-newstate-value"}

												resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testCurrentStateValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testNewStateDynamicValue,
			},
		},
		"response-state-removeresource": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
							return map[string]provider.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
										return &testprovider.Resource{
											ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
												resp.State.RemoveResource(ctx)
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testCurrentStateValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: &testNewStateRemovedDynamicValue,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ReadResource(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
