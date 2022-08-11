package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

	testStateDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, "test-id"),
		"optional": tftypes.NewValue(tftypes.String, nil),
		"required": tftypes.NewValue(tftypes.String, nil),
	})

	testSchema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"optional": {
				Optional: true,
				Type:     types.StringType,
			},
			"required": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.ImportResourceStateRequest
		expectedError    error
		expectedResponse *tfprotov6.ImportResourceStateResponse
	}{
		"request-id": {
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
										return &testprovider.ResourceWithImportState{
											Resource: &testprovider.Resource{},
											ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
												if req.ID != "test-id" {
													resp.Diagnostics.AddError("unexpected req.ID value: %s", req.ID)
												}

												resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.ImportResourceStateResponse{
				ImportedResources: []*tfprotov6.ImportedResource{
					{
						State:    testStateDynamicValue,
						TypeName: "test_resource",
					},
				},
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
										return &testprovider.ResourceWithImportState{
											Resource: &testprovider.Resource{},
											ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
			request: &tfprotov6.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.ImportResourceStateResponse{
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
			},
		},
		"response-importedresources": {
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
										return &testprovider.ResourceWithImportState{
											Resource: &testprovider.Resource{},
											ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
												resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.ImportResourceStateResponse{
				ImportedResources: []*tfprotov6.ImportedResource{
					{
						State:    testStateDynamicValue,
						TypeName: "test_resource",
					},
				},
			},
		},
		"response-importedresources-private": {
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
										return &testprovider.ResourceWithImportState{
											Resource: &testprovider.Resource{},
											ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
												diags := resp.Private.SetKey(ctx, "providerKey", []byte(`{"key": "value"}`))

												resp.Diagnostics.Append(diags...)

												resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov6.ImportResourceStateResponse{
				ImportedResources: []*tfprotov6.ImportedResource{
					{
						State:    testStateDynamicValue,
						TypeName: "test_resource",
						Private: privatestate.MustMarshalToJson(map[string][]byte{
							"providerKey": []byte(`{"key": "value"}`),
						}),
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

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
