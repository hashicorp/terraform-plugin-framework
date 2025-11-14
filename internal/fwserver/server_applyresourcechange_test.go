// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerApplyResourceChange(t *testing.T) {
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

	type testMultiIdentitySchemaData struct {
		TestAttrA types.String `tfsdk:"test_attr_a"`
		TestAttrB types.Int64  `tfsdk:"test_attr_b"`
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

	testEmptyPlan := &tfsdk.Plan{
		Raw:    tftypes.NewValue(testSchemaType, nil),
		Schema: testSchema,
	}

	testEmptyState := &tfsdk.State{
		Raw:    tftypes.NewValue(testSchemaType, nil),
		Schema: testSchema,
	}

	testEmptyIdentity := &tfsdk.ResourceIdentity{
		Raw:    tftypes.NewValue(testIdentityType, nil),
		Schema: testIdentitySchema,
	}

	type testSchemaData struct {
		TestComputed types.String `tfsdk:"test_computed"`
		TestRequired types.String `tfsdk:"test_required"`
	}

	testSchemaTypeWriteOnly := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_optional_write_only": tftypes.String,
			"test_required_write_only": tftypes.String,
		},
	}

	testSchemaWriteOnly := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_optional_write_only": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"test_required_write_only": schema.StringAttribute{
				Required:  true,
				WriteOnly: true,
			},
		},
	}

	type testSchemaDataWriteOnly struct {
		TestOptionalWriteOnly types.String `tfsdk:"test_optional_write_only"`
		TestRequiredWriteOnly types.String `tfsdk:"test_required_write_only"`
	}

	type testIdentitySchemaData struct {
		TestID types.String `tfsdk:"test_id"`
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

	testPrivateFrameworkMap := map[string][]byte{
		".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
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

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.ApplyResourceChangeRequest
		expectedResponse *fwserver.ApplyResourceChangeResponse
	}{
		"create-request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
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
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("unexpected req.Config value: %s", data.TestRequired.ValueString())
						}

						// Prevent missing resource state error diagnostic
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
		"create-request-plannedstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-plannedstate-value" {
							resp.Diagnostics.AddError("unexpected req.Plan value: %s", data.TestRequired.ValueString())
						}

						// Prevent missing resource state error diagnostic
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate},
		},
		"create-request-plannedidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
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
				PriorState:     testEmptyState,
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
						DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
						},
						UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
				Private: testEmptyPrivate},
		},
		"create-request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
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
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
					},
				},
				ProviderMeta: testProviderMetaConfig,
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
		"create-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
		"create-response-newstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"create-response-newidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
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
						"test_required": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
					}),
					Schema: testSchema,
				},
				IdentitySchema: testIdentitySchema,
				PriorState:     testEmptyState,
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
						DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
						},
						UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
		"create-response-newidentity-changes": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
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
				PriorState:     testEmptyState,
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
						DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
						},
						UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
		"create-response-newstate-null": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						// Intentionally missing resp.State.Set()
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
		"create-response-newstate-write-only": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaTypeWriteOnly, map[string]tftypes.Value{
						"test_optional_write_only": tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_required_write_only": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaWriteOnly,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaTypeWriteOnly, map[string]tftypes.Value{
						"test_optional_write_only": tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_required_write_only": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaWriteOnly,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchemaWriteOnly,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaDataWriteOnly

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeWriteOnly, map[string]tftypes.Value{
						"test_optional_write_only": tftypes.NewValue(tftypes.String, nil),
						"test_required_write_only": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchemaWriteOnly,
				},
				Private: testEmptyPrivate,
			},
		},
		"create-response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState:   testEmptyPlan,
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
						var data testSchemaData

						// Prevent missing resource state error diagnostic
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Delete")
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Create, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
		"delete-request-priorstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
					},
					DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-priorstate-value" {
							resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestRequired.ValueString())
						}
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: testEmptyState,
			},
		},
		"delete-request-prioridentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
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
						CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
						},
						DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
							var identityData testIdentitySchemaData

							resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)

							if identityData.TestID.ValueString() != "id-123" {
								resp.Diagnostics.AddError("Unexpected req.Identity Value", "Got: "+identityData.TestID.ValueString())
							}
						},
						UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewIdentity: testEmptyIdentity,
				NewState:    testEmptyState,
			},
		},
		"delete-request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
					},
					DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
						var data testProviderMetaData

						resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

						if data.TestProviderMetaAttribute.ValueString() != "test-provider-meta-value" {
							resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", data.TestProviderMetaAttribute.ValueString())
						}
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
					},
				},
				ProviderMeta: testProviderMetaConfig,
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: testEmptyState,
			},
		},
		"delete-request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
					},
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
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
					},
				},
				PlannedPrivate: &privatestate.Data{
					Provider: testProviderData,
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: testEmptyState,
			},
		},
		"delete-request-private-planned-private-nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
					},
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
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: testEmptyState,
			},
		},
		"delete-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
					},
					DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
				Private: testEmptyPrivate,
			},
		},
		"delete-response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)

						// Must return error to prevent automatic private state clearing
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
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
				Private: testPrivateProvider,
			},
		},
		"delete-response-private-updated": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedPrivate: testPrivateFramework,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)

						// Must return error to prevent automatic private state clearing
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
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
				Private: testPrivate,
			},
		},
		"delete-response-newstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
					},
					DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
						// Intentionally empty, should call resp.State.RemoveResource() automatically.
					},
					UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: testEmptyState,
			},
		},
		"delete-response-newidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-priorstate-value"),
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
						CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Create")
						},
						DeleteMethod: func(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
							// The identity is automatically set to null in the response after the Delete method is called
							if resp.Identity == nil || resp.Identity.Raw.IsNull() {
								resp.Diagnostics.AddError(
									"Unexpected resp.Identity",
									"expected resp.Identity to be a known non-null object of the schema type.",
								)
							}
						},
						UpdateMethod: func(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Delete, Got: Update")
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewIdentity: testEmptyIdentity,
				NewState:    testEmptyState,
			},
		},
		"update-request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-new-value" {
							resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-request-plannedstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						if data.TestComputed.ValueString() != "test-plannedstate-value" {
							resp.Diagnostics.AddError("Unexpected req.Plan Value", "Got: "+data.TestComputed.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-request-plannedidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "id-123"),
					}),
					Schema: testIdentitySchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				IdentitySchema: testIdentitySchema,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
						},
						DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
						},
						UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
							var identityData testIdentitySchemaData

							resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)

							if identityData.TestID.ValueString() != "id-123" {
								resp.Diagnostics.AddError("Unexpected req.Identity Value", "Got: "+identityData.TestID.ValueString())
							}
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "id-123"),
					}),
					Schema: testIdentitySchema,
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-request-priorstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-old-value" {
							resp.Diagnostics.AddError("Unexpected req.State Value", "Got: "+data.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ProviderMeta:   testProviderMetaConfig,
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						var data testProviderMetaData

						resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

						if data.TestProviderMetaAttribute.ValueString() != "test-provider-meta-value" {
							resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+data.TestProviderMetaAttribute.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
				},
				PlannedPrivate: testPrivateProvider,
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
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
		"update-request-private-nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				// Intentionally old, Update implementation does not call resp.State.Set()
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
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
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-response-newstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-response-newidentity-null-plannedidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				IdentitySchema: testIdentitySchema,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
						},
						DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
						},
						UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.Append(resp.Identity.Set(ctx, testIdentitySchemaData{
								TestID: types.StringValue("new-id-123"),
							})...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "new-id-123"),
					}),
					Schema: testIdentitySchema,
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-response-newidentity-empty-plannedidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				PlannedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testMultiAttrIdentityType, map[string]tftypes.Value{
						"test_attr_a": tftypes.NewValue(tftypes.String, nil),
						"test_attr_b": tftypes.NewValue(tftypes.Number, nil),
					}),
					Schema: testMultiAttrIdentitySchema,
				},
				IdentitySchema: testMultiAttrIdentitySchema,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
						},
						DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
						},
						UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.Append(resp.Identity.Set(ctx, testMultiIdentitySchemaData{
								TestAttrA: types.StringValue("new value"),
								TestAttrB: types.Int64Value(20),
							})...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testMultiAttrIdentityType, map[string]tftypes.Value{
						"test_attr_a": tftypes.NewValue(tftypes.String, "new value"),
						"test_attr_b": tftypes.NewValue(tftypes.Number, 20),
					}),
					Schema: testMultiAttrIdentitySchema,
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-response-newidentity-with-plannedidentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "id-123"),
					}),
					Schema: testIdentitySchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				IdentitySchema: testIdentitySchema,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
						},
						DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
						},
						UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.Append(resp.Identity.Set(ctx, testIdentitySchemaData{
								TestID: types.StringValue("id-123"),
							})...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "id-123"),
					}),
					Schema: testIdentitySchema,
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-invalid-response-newidentity-changes": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "id-123"),
					}),
					Schema: testIdentitySchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				IdentitySchema: testIdentitySchema,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
						},
						DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
						},
						UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.Append(resp.Identity.Set(ctx, testIdentitySchemaData{
								TestID: types.StringValue("new-id-123"),
							})...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Identity Change",
						"During the update operation, the Terraform Provider unexpectedly returned a different identity than the previously stored one.\n\n"+
							"This is always a problem with the provider and should be reported to the provider developer.\n\n"+
							"Planned Identity: tftypes.Object[\"test_id\":tftypes.String]<\"test_id\":tftypes.String<\"id-123\">>\n\n"+
							"New Identity: tftypes.Object[\"test_id\":tftypes.String]<\"test_id\":tftypes.String<\"new-id-123\">>",
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
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-valid-response-mutable-identity-newidentity-changes": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "id-123"),
					}),
					Schema: testIdentitySchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				IdentitySchema: testIdentitySchema,
				ResourceSchema: testSchema,
				ResourceBehavior: resource.ResourceBehavior{
					MutableIdentity: true,
				},
				Resource: &testprovider.ResourceWithIdentity{
					Resource: &testprovider.Resource{
						CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
						},
						DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
							resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
						},
						UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
							resp.Diagnostics.Append(resp.Identity.Set(ctx, testIdentitySchemaData{
								TestID: types.StringValue("new-id-123"),
							})...)
						},
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
						"test_id": tftypes.NewValue(tftypes.String, "new-id-123"),
					}),
					Schema: testIdentitySchema,
				},
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-response-newstate-null": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						resp.State.RemoveResource(ctx)
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Missing Resource State After Update",
						"The Terraform Provider unexpectedly returned no resource state after having no errors in the resource update. "+
							"This is always an issue in the Terraform Provider and should be reported to the provider developers.",
					),
				},
				NewState: testEmptyState,
				Private:  testEmptyPrivate,
			},
		},
		"update-response-newstate-write-only-nullification": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaTypeWriteOnly, map[string]tftypes.Value{
						"test_optional_write_only": tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_required_write_only": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaWriteOnly,
				},
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaTypeWriteOnly, map[string]tftypes.Value{
						"test_optional_write_only": tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_required_write_only": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaWriteOnly,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeWriteOnly, map[string]tftypes.Value{
						"test_optional_write_only": tftypes.NewValue(tftypes.String, "old-optional-value"),
						"test_required_write_only": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaWriteOnly,
				},
				ResourceSchema: testSchemaWriteOnly,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						var data testSchemaDataWriteOnly

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeWriteOnly, map[string]tftypes.Value{
						"test_optional_write_only": tftypes.NewValue(tftypes.String, nil),
						"test_required_write_only": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchemaWriteOnly,
				},
				Private: testEmptyPrivate,
			},
		},
		"update-response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				Private: testPrivateProvider,
			},
		},
		"update-response-private-updated": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ApplyResourceChangeRequest{
				PlannedState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.Resource{
					CreateMethod: func(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Create")
					},
					DeleteMethod: func(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
						resp.Diagnostics.AddError("Unexpected Method Call", "Expected: Update, Got: Delete")
					},
					UpdateMethod: func(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
				PlannedPrivate: testPrivateFramework,
			},
			expectedResponse: &fwserver.ApplyResourceChangeResponse{
				NewState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				Private: testPrivate,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ApplyResourceChangeResponse{}
			testCase.server.ApplyResourceChange(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
