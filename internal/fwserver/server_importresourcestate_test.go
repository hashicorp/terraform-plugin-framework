package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	testEmptyStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, nil),
		"optional": tftypes.NewValue(tftypes.String, nil),
		"required": tftypes.NewValue(tftypes.String, nil),
	})

	testStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
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

	testEmptyState := &tfsdk.State{
		Raw:    testEmptyStateValue,
		Schema: testSchema,
	}

	testState := &tfsdk.State{
		Raw:    testStateValue,
		Schema: testSchema,
	}

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.ImportResourceStateRequest
		expectedResponse *fwserver.ImportResourceStateResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{},
		},
		"request-id": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				ResourceType: &testprovider.ResourceType{
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
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testState,
						TypeName: "test_resource",
					},
				},
			},
		},
		"request-resourcetype-importstate-not-implemented": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{}, nil
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Resource Import Not Implemented",
						"This resource does not support import. Please contact the provider developer for additional information.",
					),
				},
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				ResourceType: &testprovider.ResourceType{
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
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
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
			},
		},
		"response-importedresources": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				ResourceType: &testprovider.ResourceType{
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
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testState,
						TypeName: "test_resource",
					},
				},
			},
		},
		"response-importedresources-empty-state": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithImportState{
							Resource: &testprovider.Resource{},
							ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
								// Intentionally empty
							},
						}, nil
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Missing Resource Import State",
						"An unexpected error was encountered when importing the resource. This is always a problem with the provider. Please give the following information to the provider developer:\n\n"+
							"Resource ImportState method returned no State in response. If import is intentionally not supported, remove the Resource type ImportState method or return an error.",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ImportResourceStateResponse{}
			testCase.server.ImportResourceState(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
