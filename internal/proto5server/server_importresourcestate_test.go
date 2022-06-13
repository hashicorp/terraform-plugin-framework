package proto5server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
		request          *tfprotov5.ImportResourceStateRequest
		expectedError    error
		expectedResponse *tfprotov5.ImportResourceStateResponse
	}{
		"request-id": {
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
										return &testprovider.ResourceWithImportState{
											Resource: &testprovider.Resource{},
											ImportStateMethod: func(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
												if req.ID != "test-id" {
													resp.Diagnostics.AddError("unexpected req.ID value: %s", req.ID)
												}

												tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov5.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ImportResourceStateResponse{
				ImportedResources: []*tfprotov5.ImportedResource{
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
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.ResourceWithImportState{
											Resource: &testprovider.Resource{},
											ImportStateMethod: func(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
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
			request: &tfprotov5.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ImportResourceStateResponse{
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
		"response-importedresources": {
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
										return &testprovider.ResourceWithImportState{
											Resource: &testprovider.Resource{},
											ImportStateMethod: func(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
												tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov5.ImportResourceStateRequest{
				ID:       "test-id",
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ImportResourceStateResponse{
				ImportedResources: []*tfprotov5.ImportedResource{
					{
						State:    testStateDynamicValue,
						TypeName: "test_resource",
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
