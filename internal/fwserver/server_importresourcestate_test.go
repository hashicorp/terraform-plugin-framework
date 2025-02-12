// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

	testTypeWriteOnly := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":         tftypes.String,
			"write-only": tftypes.String,
			"required":   tftypes.String,
		},
	}

	testEmptyStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, nil),
		"optional": tftypes.NewValue(tftypes.String, nil),
		"required": tftypes.NewValue(tftypes.String, nil),
	})

	testEmptyStateValueWriteOnly := tftypes.NewValue(testTypeWriteOnly, map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, nil),
		"write-only": tftypes.NewValue(tftypes.String, nil),
		"required":   tftypes.NewValue(tftypes.String, nil),
	})

	testUnknownStateValue := tftypes.NewValue(testType, tftypes.UnknownValue)

	testStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, "test-id"),
		"optional": tftypes.NewValue(tftypes.String, nil),
		"required": tftypes.NewValue(tftypes.String, nil),
	})

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"optional": schema.StringAttribute{
				Optional: true,
			},
			"required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testSchemaWriteOnly := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"write-only": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testEmptyState := &tfsdk.State{
		Raw:    testEmptyStateValue,
		Schema: testSchema,
	}

	testEmptyStateWriteOnly := &tfsdk.State{
		Raw:    testEmptyStateValueWriteOnly,
		Schema: testSchemaWriteOnly,
	}

	testUnknownState := &tfsdk.State{
		Raw:    testUnknownStateValue,
		Schema: testSchema,
	}

	testState := &tfsdk.State{
		Raw:    testStateValue,
		Schema: testSchema,
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testPrivate := &privatestate.Data{
		Provider: testProviderData,
	}

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testEmptyPrivate := &privatestate.Data{
		Provider: testEmptyProviderData,
	}

	testDeferral := resource.ImportStateClientCapabilities{
		DeferralAllowed: true,
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.ImportResourceStateRequest
		expectedResponse     *fwserver.ImportResourceStateResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{},
		},
		"request-client-capabilities-deferral-allowed": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				ClientCapabilities: testDeferral,
				EmptyState:         *testEmptyState,
				ID:                 "test-id",
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						if req.ClientCapabilities.DeferralAllowed != true {
							resp.Diagnostics.AddError("Unexpected req.ClientCapabilities.DeferralAllowed value",
								"expected: true but got: false")
						}

						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testState,
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
					},
				},
			},
		},
		"request-id": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						if req.ID != "test-id" {
							resp.Diagnostics.AddError("unexpected req.ID value: %s", req.ID)
						}

						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testState,
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
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
				Resource:   &testprovider.Resource{},
				TypeName:   "test_resource",
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
		"resource-configure-data": {
			server: &fwserver.Server{
				Provider:              &testprovider.Provider{},
				ResourceConfigureData: "test-provider-configure-value",
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				TypeName:   "test_resource",
				Resource: &testprovider.ResourceWithConfigureAndImportState{
					ConfigureMethod: func(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
						providerData, ok := req.ProviderData.(string)

						if !ok {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.ProviderData",
								fmt.Sprintf("Expected string, got: %T", req.ProviderData),
							)
							return
						}

						if providerData != "test-provider-configure-value" {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.ProviderData",
								fmt.Sprintf("Expected test-provider-configure-value, got: %q", providerData),
							)
						}
					},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						// In practice, the Configure method would save the
						// provider data to the Resource implementation and
						// use it here. The fact that Configure is able to
						// read the data proves this can work.

						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
					Resource: &testprovider.Resource{},
				},
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testState,
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
					},
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
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
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
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testState,
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
					},
				},
			},
		},
		"response-importedresources-deferral-automatic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
						resp.Deferred = &provider.Deferred{Reason: provider.DeferredReasonProviderConfigUnknown}
					},
				},
			},
			configureProviderReq: &provider.ConfigureRequest{
				ClientCapabilities: provider.ConfigureProviderClientCapabilities{
					DeferralAllowed: true,
				},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resp.Diagnostics.AddError("Test assertion failed: ", "import shouldn't be called")
					},
				},
				TypeName:           "test_resource",
				ClientCapabilities: testDeferral,
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testUnknownState,
						TypeName: "test_resource",
						Private:  &privatestate.Data{},
					},
				},
				Deferred: &resource.Deferred{Reason: resource.DeferredReasonProviderConfigUnknown},
			},
		},
		"response-importedresources-deferral-manual": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						if req.ID != "test-id" {
							resp.Diagnostics.AddError("unexpected req.ID value: %s", req.ID)
						}

						resp.Deferred = &resource.Deferred{
							Reason: resource.DeferredReasonAbsentPrereq,
						}

						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

					},
				},
				TypeName:           "test_resource",
				ClientCapabilities: testDeferral,
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testState,
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
					},
				},
				Deferred: &resource.Deferred{Reason: resource.DeferredReasonAbsentPrereq},
			},
		},
		"response-importedresources-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)

						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testState,
						TypeName: "test_resource",
						Private:  testPrivate,
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
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						// Intentionally empty
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
		"response-importedresources-write-only-nullification": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyStateWriteOnly,
				ID:         "test-id",
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("write-only"), "write-only-val")...)
						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State: tfsdk.State{
							Raw: tftypes.NewValue(testTypeWriteOnly, map[string]tftypes.Value{
								"id":         tftypes.NewValue(tftypes.String, "test-id"),
								"write-only": tftypes.NewValue(tftypes.String, nil),
								"required":   tftypes.NewValue(tftypes.String, nil),
							}),
							Schema: testSchemaWriteOnly,
						},
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.configureProviderReq != nil {
				configureProviderResp := &provider.ConfigureResponse{}
				testCase.server.ConfigureProvider(context.Background(), testCase.configureProviderReq, configureProviderResp)
			}

			response := &fwserver.ImportResourceStateResponse{}
			testCase.server.ImportResourceState(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
