// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerReadResource(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testIdentityType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id": tftypes.String,
		},
	}

	testTypeWriteOnly := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_write_only": tftypes.String,
			"test_required":   tftypes.String,
		},
	}

	testCurrentStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, nil),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testCurrentIdentityValue := tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
		"test_id": tftypes.NewValue(tftypes.String, "id-123"),
	})

	testCurrentStateValueWriteOnly := tftypes.NewValue(testTypeWriteOnly, map[string]tftypes.Value{
		"test_write_only": tftypes.NewValue(tftypes.String, nil),
		"test_required":   tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testNewStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-newstate-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testNewIdentityValue := tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
		"test_id": tftypes.NewValue(tftypes.String, "new-id-123"),
	})

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}

	testMultiAttrIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_attr_a": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"test_attr_b": identityschema.Int64Attribute{
				OptionalForImport: true,
			},
		},
	}

	testMultiAttrIdentityType := testMultiAttrIdentitySchema.Type().TerraformType(context.Background())

	testSchemaWriteOnly := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_write_only": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testSchemaWithSemanticEquals := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				CustomType: testtypes.StringTypeWithSemanticEquals{
					SemanticEquals: true,
				},
				Required: true,
			},
		},
	}

	testSchemaWithSemanticEqualsDiagnostics := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				CustomType: testtypes.StringTypeWithSemanticEquals{
					SemanticEquals: true,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Required: true,
			},
		},
	}

	testConfig := &tfsdk.Config{
		Raw:    testCurrentStateValue,
		Schema: testSchema,
	}

	testCurrentState := &tfsdk.State{
		Raw:    testCurrentStateValue,
		Schema: testSchema,
	}

	testCurrentIdentity := &tfsdk.ResourceIdentity{
		Raw:    testCurrentIdentityValue,
		Schema: testIdentitySchema,
	}

	testCurrentStateWriteOnly := &tfsdk.State{
		Raw:    testCurrentStateValueWriteOnly,
		Schema: testSchemaWriteOnly,
	}

	testNewState := &tfsdk.State{
		Raw:    testNewStateValue,
		Schema: testSchema,
	}

	testNewIdentity := &tfsdk.ResourceIdentity{
		Raw:    testNewIdentityValue,
		Schema: testIdentitySchema,
	}

	testEmptyIdentity := &tfsdk.ResourceIdentity{
		Schema: testIdentitySchema,
		Raw:    tftypes.NewValue(testIdentitySchema.Type().TerraformType(context.Background()), nil),
	}

	testNewStateRemoved := &tfsdk.State{
		Raw:    tftypes.NewValue(testType, nil),
		Schema: testSchema,
	}

	testPrivateFrameworkMap := map[string][]byte{
		".frameworkKey": []byte(`{"fk": "framework value"}`),
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testPrivate := &privatestate.Data{
		Framework: testPrivateFrameworkMap,
		Provider:  testProviderData,
	}

	testPrivateFramework := &privatestate.Data{
		Framework: testPrivateFrameworkMap,
	}

	testPrivateProvider := &privatestate.Data{
		Provider: testProviderData,
	}

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testEmptyPrivate := &privatestate.Data{
		Provider: testEmptyProviderData,
	}

	testPrivateAfterImport := &privatestate.Data{
		Framework: map[string][]byte{
			privatestate.ImportBeforeReadKey: []byte(`true`),
		},
		Provider: testEmptyProviderData,
	}

	testDeferralAllowed := resource.ReadClientCapabilities{
		DeferralAllowed: true,
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.ReadResourceRequest
		expectedResponse     *fwserver.ReadResourceResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ReadResourceResponse{},
		},
		"request-client-capabilities-deferral-allowed": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				ClientCapabilities: testDeferralAllowed,
				CurrentState:       testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						if req.ClientCapabilities.DeferralAllowed != true {
							resp.Diagnostics.AddError("Unexpected req.ClientCapabilities.DeferralAllowed value",
								"expected: true but got: false")
						}
						var data struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"request-currentstate-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{},
			expectedResponse: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Read Request",
						"An unexpected error was encountered when reading the resource. The current state was missing.\n\n"+
							"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
					),
				},
			},
		},
		"request-currentstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var data struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-currentstate-value" {
							resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"request-currentidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState:    testCurrentState,
				IdentitySchema:  testIdentitySchema,
				CurrentIdentity: testCurrentIdentity,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							var identityData struct {
								TestID types.String `tfsdk:"test_id"`
							}

							resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)

							if identityData.TestID.ValueString() != "id-123" {
								resp.Diagnostics.AddError("unexpected req.Identity value: %s", identityData.TestID.ValueString())
							}
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState:    testCurrentState,
				NewIdentity: testCurrentIdentity,
				Private:     testEmptyPrivate,
			},
		},
		"request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var config struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &config)...)

						if config.TestRequired.ValueString() != "test-currentstate-value" {
							resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", config.TestRequired.ValueString())
						}
					},
				},
				ProviderMeta: testConfig,
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

						key := "providerKeyOne"
						got, diags := req.Private.GetKey(ctx, key)

						resp.Diagnostics.Append(diags...)

						if string(got) != expected {
							resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
						}
					},
				},
				Private: testPrivate,
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testPrivate,
			},
		},
		"request-private-nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var expected []byte

						key := "providerKeyOne"
						got, diags := req.Private.GetKey(ctx, key)

						resp.Diagnostics.Append(diags...)

						if !bytes.Equal(got, expected) {
							resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
						}
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"resource-configure-data": {
			server: &fwserver.Server{
				Provider:              &testprovider.Provider{},
				ResourceConfigureData: "test-provider-configure-value",
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.ResourceWithConfigure{
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
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							// In practice, the Configure method would save the
							// provider data to the Resource implementation and
							// use it here. The fact that Configure is able to
							// read the data proves this can work.
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"response-deferral-automatic": {
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
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						resp.Diagnostics.AddError("Test assertion failed: ", "read shouldn't be called")
					},
				},
				ClientCapabilities: testDeferralAllowed,
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Deferred: &resource.Deferred{Reason: resource.DeferredReasonProviderConfigUnknown},
			},
		},
		"response-deferral-manual": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var data struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						resp.Deferred = &resource.Deferred{Reason: resource.DeferredReasonAbsentPrereq}

						if data.TestRequired.ValueString() != "test-currentstate-value" {
							resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestRequired.ValueString())
						}
					},
				},
				ClientCapabilities: testDeferralAllowed,
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
				Deferred: &resource.Deferred{Reason: resource.DeferredReasonAbsentPrereq},
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
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
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"response-diagnostics-semantic-equality": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: &tfsdk.State{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
					}),
					Schema: testSchemaWithSemanticEqualsDiagnostics,
				},
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var data struct {
							TestComputed types.String                            `tfsdk:"test_computed"`
							TestRequired testtypes.StringValueWithSemanticEquals `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						data.TestRequired = testtypes.StringValueWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
							StringValue: types.StringValue("test-semantic-equal-value"),
						}

						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						// The response state is intentionally not updated when there are diagnostics
						"test_required": tftypes.NewValue(tftypes.String, "test-semantic-equal-value"),
					}),
					Schema: testSchemaWithSemanticEqualsDiagnostics,
				},
				Private: testEmptyPrivate,
			},
		},
		"response-state": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var data struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						data.TestComputed = types.StringValue("test-newstate-value")

						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testNewState,
				Private:  testEmptyPrivate,
			},
		},
		"response-identity-new": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				// Resource supports identity but there isn't one in state yet
				CurrentIdentity: nil,
				IdentitySchema:  testIdentitySchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							if !req.Identity.Raw.IsNull() {
								resp.Diagnostics.AddError("Unexpected request", "expected req.Identity to be null")
								return
							}

							identityData := struct {
								TestID types.String `tfsdk:"test_id"`
							}{
								TestID: types.StringValue("new-id-123"),
							}

							resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityData)...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState:    testCurrentState,
				NewIdentity: testNewIdentity,
				Private:     testEmptyPrivate,
			},
		},
		"response-invalid-nil-identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState:    testCurrentState,
				CurrentIdentity: nil,
				IdentitySchema:  testIdentitySchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							resp.Identity = req.Identity
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Missing Resource Identity After Read",
						"The Terraform Provider unexpectedly returned no resource identity data after having no errors in the resource read. "+
							"This is always an issue in the Terraform Provider and should be reported to the provider developers.",
					),
				},
				NewState:    testCurrentState,
				NewIdentity: testEmptyIdentity,
				Private:     testEmptyPrivate,
			},
		},
		"response-identity-valid-update-null-currentidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState:   testCurrentState,
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							identityData := struct {
								TestID types.String `tfsdk:"test_id"`
							}{
								TestID: types.StringValue("new-id-123"),
							}

							resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityData)...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState:    testCurrentState,
				NewIdentity: testNewIdentity,
				Private:     testEmptyPrivate,
			},
		},
		"response-identity-valid-update-empty-currentidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				CurrentIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testMultiAttrIdentityType, map[string]tftypes.Value{
						"test_attr_a": tftypes.NewValue(tftypes.String, nil),
						"test_attr_b": tftypes.NewValue(tftypes.Number, nil),
					}),
					Schema: testMultiAttrIdentitySchema,
				},
				IdentitySchema: testMultiAttrIdentitySchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							identityData := struct {
								TestAttrA types.String `tfsdk:"test_attr_a"`
								TestAttrB types.Int64  `tfsdk:"test_attr_b"`
							}{
								TestAttrA: types.StringValue("new value"),
								TestAttrB: types.Int64Value(20),
							}

							resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityData)...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				NewIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testMultiAttrIdentityType, map[string]tftypes.Value{
						"test_attr_a": tftypes.NewValue(tftypes.String, "new value"),
						"test_attr_b": tftypes.NewValue(tftypes.Number, 20),
					}),
					Schema: testMultiAttrIdentitySchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"response-identity-valid-update-after-import": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState:    testCurrentState,
				CurrentIdentity: testCurrentIdentity,
				Private:         testPrivateAfterImport,
				IdentitySchema:  testIdentitySchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							var identityData struct {
								TestID types.String `tfsdk:"test_id"`
							}

							resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)

							identityData.TestID = types.StringValue("new-id-123")

							resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityData)...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState:    testCurrentState,
				NewIdentity: testNewIdentity,
				Private: &privatestate.Data{
					Framework: make(map[string][]byte, 0), // Private import key should be cleared
					Provider:  testEmptyProviderData,
				},
			},
		},
		"response-identity-invalid-update": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState:    testCurrentState,
				CurrentIdentity: testCurrentIdentity,
				IdentitySchema:  testIdentitySchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							var identityData struct {
								TestID types.String `tfsdk:"test_id"`
							}

							resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)

							identityData.TestID = types.StringValue("new-id-123")

							resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityData)...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Identity Change",
						"During the read operation, the Terraform Provider unexpectedly returned a different identity than the previously stored one.\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.\n\n"+
							"Current Identity: tftypes.Object[\"test_id\":tftypes.String]<\"test_id\":tftypes.String<\"id-123\">>\n\n"+
							"New Identity: tftypes.Object[\"test_id\":tftypes.String]<\"test_id\":tftypes.String<\"new-id-123\">>",
					),
				},
				NewState:    testCurrentState,
				NewIdentity: testNewIdentity,
				Private:     testEmptyPrivate,
			},
		},
		"response-identity-valid-update-mutable-identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState:    testCurrentState,
				CurrentIdentity: testCurrentIdentity,
				ResourceBehavior: resource.ResourceBehavior{
					MutableIdentity: true,
				},
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							var identityData struct {
								TestID types.String `tfsdk:"test_id"`
							}

							resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)

							identityData.TestID = types.StringValue("new-id-123")

							resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityData)...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState:    testCurrentState,
				NewIdentity: testNewIdentity,
				Private:     testEmptyPrivate,
			},
		},
		"response-invalid-identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						// This resource doesn't indicate identity support (via a schema), so this should raise a diagnostic.
						resp.Identity = &tfsdk.ResourceIdentity{
							Raw:    testNewIdentityValue,
							Schema: testIdentitySchema,
						}
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Read Response",
						"An unexpected error was encountered when creating the read response. New identity data was returned by the provider read operation, but the resource does not indicate identity support.\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
				NewState:    testCurrentState,
				NewIdentity: testNewIdentity,
				Private:     testEmptyPrivate,
			},
		},
		"response-state-removeresource": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						resp.State.RemoveResource(ctx)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testNewStateRemoved,
				Private:  testEmptyPrivate,
			},
		},
		"response-state-semantic-equality": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: &tfsdk.State{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
					}),
					Schema: testSchemaWithSemanticEquals,
				},
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var data struct {
							TestComputed types.String                            `tfsdk:"test_computed"`
							TestRequired testtypes.StringValueWithSemanticEquals `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						// This value should be overwritten back to the config value.
						data.TestRequired = testtypes.StringValueWithSemanticEquals{
							SemanticEquals: true,
							StringValue:    types.StringValue("test-semantic-equal-value"),
						}

						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
					}),
					Schema: testSchemaWithSemanticEquals,
				},
				Private: testEmptyPrivate,
			},
		},
		"response-state-write-only-nullification": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentStateWriteOnly,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var data struct {
							TestWriteOnly types.String `tfsdk:"test_write_only"`
							TestRequired  types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						data.TestWriteOnly = types.StringValue("test-write-only-value")

						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testTypeWriteOnly, map[string]tftypes.Value{
						"test_write_only": tftypes.NewValue(tftypes.String, nil),
						"test_required":   tftypes.NewValue(tftypes.String, "test-currentstate-value"),
					}),
					Schema: testSchemaWriteOnly,
				},
				Private: testEmptyPrivate,
			},
		},
		"response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testPrivateProvider,
			},
		},
		"response-private-updated": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
				Private: testPrivateFramework,
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testPrivate,
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

			response := &fwserver.ReadResourceResponse{}
			testCase.server.ReadResource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
