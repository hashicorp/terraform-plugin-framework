// Copyright IBM Corp. 2021, 2026
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
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	testIdentityType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id":       tftypes.String,
			"other_test_id": tftypes.String,
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

	testRequestIdentityValue := tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
		"test_id":       tftypes.NewValue(tftypes.String, "id-123"),
		"other_test_id": tftypes.NewValue(tftypes.String, nil),
	})

	testImportedResourceIdentityValue := tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
		"test_id":       tftypes.NewValue(tftypes.String, "id-123"),
		"other_test_id": tftypes.NewValue(tftypes.String, "new-value-123"),
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

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"other_test_id": identityschema.StringAttribute{
				OptionalForImport: true,
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

	testRequestIdentity := &tfsdk.ResourceIdentity{
		Raw:    testRequestIdentityValue,
		Schema: testIdentitySchema,
	}

	testState := &tfsdk.State{
		Raw:    testStateValue,
		Schema: testSchema,
	}

	testStatePassThroughIdentity := &tfsdk.State{
		Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
			"id":       tftypes.NewValue(tftypes.String, "id-123"),
			"optional": tftypes.NewValue(tftypes.String, nil),
			"required": tftypes.NewValue(tftypes.String, nil),
		}),
		Schema: testSchema,
	}

	testImportedResourceIdentity := &tfsdk.ResourceIdentity{
		Raw:    testImportedResourceIdentityValue,
		Schema: testIdentitySchema,
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testPrivate := &privatestate.Data{
		Framework: map[string][]byte{
			privatestate.ImportBeforeReadKey: []byte(`true`),
		},
		Provider: testProviderData,
	}

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testEmptyPrivate := &privatestate.Data{
		Framework: map[string][]byte{
			privatestate.ImportBeforeReadKey: []byte(`true`),
		},
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
		"request-identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState:     *testEmptyState,
				Identity:       testRequestIdentity,
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithIdentityAndImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						var identityData struct {
							TestID      types.String `tfsdk:"test_id"`
							OtherTestID types.String `tfsdk:"other_test_id"`
						}

						resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)

						if identityData.TestID.ValueString() != "id-123" {
							resp.Diagnostics.AddError("unexpected req.Identity value: %s", identityData.TestID.ValueString())
						}

						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
					IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
						resp.IdentitySchema = testIdentitySchema
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testEmptyState,
						Identity: testRequestIdentity,
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
		"response-importedresources-identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState:     *testEmptyState,
				Identity:       testRequestIdentity,
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithIdentityAndImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("other_test_id"), types.StringValue("new-value-123"))...)

						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
					IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
						resp.IdentitySchema = testIdentitySchema
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testEmptyState,
						Identity: testImportedResourceIdentity,
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
					},
				},
			},
		},
		"response-importedresources-identity-supported": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState:     *testEmptyState,
				ID:             "test-id",
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("test_id"), types.StringValue("id-123"))...)
						resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("other_test_id"), types.StringValue("new-value-123"))...)
						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testState,
						Identity: testImportedResourceIdentity,
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
					},
				},
			},
		},
		"response-importedresources-invalid-identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState: *testEmptyState,
				ID:         "test-id",
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						// This resource doesn't indicate identity support (via a schema), so this should raise a diagnostic.
						resp.Identity = &tfsdk.ResourceIdentity{
							Raw:    testImportedResourceIdentityValue,
							Schema: testIdentitySchema,
						}
						resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected ImportState Response",
						"An unexpected error was encountered when creating the import response. New identity data was returned by the provider import operation, but the resource does not indicate identity support.\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.",
					),
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
		"response-importedresources-passthrough-identity-imported-by-id": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState:     *testEmptyState,
				ID:             "id-123",
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("test_id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State: *testStatePassThroughIdentity,
						Identity: &tfsdk.ResourceIdentity{
							Raw:    tftypes.NewValue(testIdentityType, nil),
							Schema: testIdentitySchema,
						},
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
					},
				},
			},
		},
		"response-importedresources-passthrough-identity-imported-by-identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState:     *testEmptyState,
				Identity:       testRequestIdentity,
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("other_test_id"), types.StringValue("new-value-123"))...)
						resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("test_id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State:    *testStatePassThroughIdentity,
						Identity: testImportedResourceIdentity,
						TypeName: "test_resource",
						Private:  testEmptyPrivate,
					},
				},
			},
		},
		"response-importedresources-passthrough-identity-invalid-state-path": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState:     *testEmptyState,
				ID:             "id-123",
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resource.ImportStatePassthroughWithIdentity(ctx, path.Root("not-valid"), path.Root("test_id"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("not-valid"),
						"State Write Error",
						"An unexpected error was encountered trying to retrieve type information at a given path. "+
							"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"Error: AttributeName(\"not-valid\") still remains in the path: could not find attribute or block "+
							"\"not-valid\" in schema",
					),
				},
			},
		},
		"response-importedresources-passthrough-identity-invalid-identity-path": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ImportResourceStateRequest{
				EmptyState:     *testEmptyState,
				Identity:       testRequestIdentity,
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithImportState{
					Resource: &testprovider.Resource{},
					ImportStateMethod: func(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
						resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("not-valid"), req, resp)
					},
				},
				TypeName: "test_resource",
			},
			expectedResponse: &fwserver.ImportResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("not-valid"),
						"Resource Identity Read Error",
						"An unexpected error was encountered trying to retrieve type information at a given path. "+
							"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"Error: AttributeName(\"not-valid\") still remains in the path: could not find attribute or block "+
							"\"not-valid\" in schema",
					),
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
