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
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerCreateResource(t *testing.T) {
	t.Parallel()

	testSchemaType := tftypes.Object{
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

	testSchemaTypeWriteOnly := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_required":   tftypes.String,
			"test_write_only": tftypes.String,
		},
	}

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

	testSchemaWithWriteOnly := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_required": schema.StringAttribute{
				Required: true,
			},
			"test_write_only": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
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

	type testIdentitySchemaData struct {
		TestID types.String `tfsdk:"test_id"`
	}

	type testSchemaDataWithSemanticEquals struct {
		TestComputed types.String                            `tfsdk:"test_computed"`
		TestRequired testtypes.StringValueWithSemanticEquals `tfsdk:"test_required"`
	}

	type testSchemaDataWriteOnly struct {
		TestRequired  types.String `tfsdk:"test_required"`
		TestWriteOnly types.String `tfsdk:"test_write_only"`
	}

	testProviderMetaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_provider_meta_attribute": tftypes.String,
		},
	}

	testProviderMetaValue := tftypes.NewValue(testProviderMetaType, map[string]tftypes.Value{
		"test_provider_meta_attribute": tftypes.NewValue(tftypes.String, "test-provider-meta-value"),
	})

	testProviderMetaSchema := metaschema.Schema{
		Attributes: map[string]metaschema.Attribute{
			"test_provider_meta_attribute": metaschema.StringAttribute{
				Optional: true,
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

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testEmptyPrivate := &privatestate.Data{
		Provider: testEmptyProviderData,
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
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.ValueString())
						}

						// Prevent missing resource state error diagnostic
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
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
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-plannedstate-value" {
							resp.Diagnostics.AddError("Unexpected req.Plan Value", "Got: "+data.TestRequired.ValueString())
						}

						// Prevent missing resource state error diagnostic
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
				Private: testEmptyPrivate,
			},
		},
		"request-plannedidentity": {
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
				PlannedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "id-123"),
					}),
					Schema: testIdentitySchema,
				},
				IdentitySchema: testIdentitySchema,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
							var identityData testIdentitySchemaData

							resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)

							if identityData.TestID.ValueString() != "id-123" {
								resp.Diagnostics.AddError("Unexpected req.Identity Value", "Got: "+identityData.TestID.ValueString())
							}

							// Prevent missing resource state error diagnostic
							var data testSchemaData

							resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
							resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
						},
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				NewIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "id-123"),
					}),
					Schema: testIdentitySchema,
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"request-providermeta": {
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
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var metadata testProviderMetaData

						resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &metadata)...)

						if metadata.TestProviderMetaAttribute.ValueString() != "test-provider-meta-value" {
							resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+metadata.TestProviderMetaAttribute.ValueString())
						}

						// Prevent missing resource state error diagnostic
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
				ProviderMeta: testProviderMetaConfig,
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"resource-configure-data": {
			server: &fwserver.Server{
				Provider:              &testprovider.Provider{},
				ResourceConfigureData: "test-provider-configure-value",
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
						CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
							// In practice, the Configure method would save the
							// provider data to the Resource implementation and
							// use it here. The fact that Configure is able to
							// read the data proves this can work.

							// Prevent missing resource state error diagnostic
							var data testSchemaData

							resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
							resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
						},
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
				Private: testEmptyPrivate,
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
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
				Private:  testEmptyPrivate,
			},
		},
		"response-diagnostics-semantic-equality": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchemaWithSemanticEqualsDiagnostics,
				},
				ResourceSchema: testSchemaWithSemanticEqualsDiagnostics,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaDataWithSemanticEquals

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

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
			expectedResponse: &fwserver.CreateResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						// The response state is intentionally not updated when there are diagnostics
						"test_required": tftypes.NewValue(tftypes.String, "test-semantic-equal-value"),
					}),
					Schema: testSchemaWithSemanticEqualsDiagnostics,
				},
				Private: testEmptyPrivate,
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
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
				Private: testEmptyPrivate,
			},
		},
		"response-newidentity": {
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
				IdentitySchema: testIdentitySchema,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
							resp.Diagnostics.Append(resp.Identity.Set(ctx, testIdentitySchemaData{
								TestID: types.StringValue("new-id-123"),
							})...)

							// Prevent missing resource state error diagnostic
							var data testSchemaData

							resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
							resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
						},
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				NewIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "new-id-123"),
					}),
					Schema: testIdentitySchema,
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"response-invalid-newidentity": {
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
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
							// This resource doesn't indicate identity support (via a schema), so this should raise a diagnostic.
							resp.Identity = &tfsdk.ResourceIdentity{
								Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
									"test_id": tftypes.NewValue(tftypes.String, "new-id-123"),
								}),
								Schema: testIdentitySchema,
							}

							// Prevent missing resource state error diagnostic
							var data testSchemaData

							resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
							resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
						},
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Create Response",
						"An unexpected error was encountered when creating the apply response. New identity data was returned by the provider create operation, but the resource does not indicate identity support.\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
				NewIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "new-id-123"),
					}),
					Schema: testIdentitySchema,
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"response-newstate-null": {
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
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						// Intentionally missing resp.State.Set()
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
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
				Private:  testEmptyPrivate,
			},
		},
		"response-newstate-semantic-equality": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchemaWithSemanticEquals,
				},
				ResourceSchema: testSchemaWithSemanticEquals,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaDataWithSemanticEquals

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						// This value should be overwritten back to the plan value.
						data.TestRequired = testtypes.StringValueWithSemanticEquals{
							SemanticEquals: true,
							StringValue:    types.StringValue("test-semantic-equal-value"),
						}

						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchemaWithSemanticEquals,
				},
				Private: testEmptyPrivate,
			},
		},
		"response-newstate-write-only-nullification": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaTypeWriteOnly, map[string]tftypes.Value{
						"test_required":   tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_write_only": tftypes.NewValue(tftypes.String, "test-write-only-value"),
					}),
					Schema: testSchemaWithWriteOnly,
				},
				ResourceSchema: testSchemaWithWriteOnly,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaDataWriteOnly

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeWriteOnly, map[string]tftypes.Value{
						"test_required":   tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_write_only": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchemaWithWriteOnly,
				},
				Private: testEmptyPrivate,
			},
		},
		"response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CreateResourceRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaData

						// Prevent missing resource state error diagnostic
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.CreateResourceResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				Private: &privatestate.Data{
					Provider: testProviderData,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.CreateResourceResponse{}
			testCase.server.CreateResource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
