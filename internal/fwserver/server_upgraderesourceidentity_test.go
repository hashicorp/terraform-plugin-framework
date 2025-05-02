// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestServerUpgradeResourceIdentity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
		Version: 1,
	}

	schemaIdentityType := testIdentitySchema.Type().TerraformType(ctx)

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.UpgradeResourceIdentityRequest
		expectedResponse *fwserver.UpgradeResourceIdentityResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{},
		},
		"resource-configure-data": {
			server: &fwserver.Server{
				Provider:              &testprovider.Provider{},
				ResourceConfigureData: "test-provider-configure-value",
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id": "test-id-value",
				}),
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithConfigureAndUpgradeIdentity{
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
					Resource: &testprovider.Resource{},
					UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
						return map[int64]resource.IdentityUpgrader{
							0: {
								PriorSchema: &identityschema.Schema{
									Attributes: map[string]identityschema.Attribute{
										"id": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								IdentityUpgrader: func(ctx context.Context, req resource.UpgradeResourceIdentityRequest, resp *resource.UpgradeResourceIdentityResponse) {
									// In practice, the Configure method would save the
									// provider data to the Resource implementation and
									// use it here. The fact that Configure is able to
									// read the data proves this can work.

									rawStateValue, err := req.RawState.Unmarshal(tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"id": tftypes.String,
										},
									})

									if err != nil {
										resp.Diagnostics.AddError(
											"Unable to Read Previously Saved Identity for UpgradeResourceIdentity",
											fmt.Sprintf("There was an error reading the saved resource Identity using the prior resource schema defined for version %d upgrade.\n\n", req.Version)+
												"Please report this to the provider developer:\n\n"+err.Error(),
										)
										return
									}
									rawValues := make(map[string]tftypes.Value)
									err = rawStateValue.As(&rawValues)
									if err != nil {
										resp.Diagnostics.AddError(
											"Unable to convert raw state value into prior identity struct",
											fmt.Sprintf("There was an error converting the raw state value into the prior resource identity struct for version %d upgrade.\n\n", req.Version)+
												"Please report this to the provider developer:\n\n"+err.Error(),
										)
										return
									}

									upgradedIdentityData := struct {
										Id string `tfsdk:"id"`
									}{
										Id: id,
									}

									resp.Diagnostics.Append(resp.UpgradedIdentity.Set(ctx, upgradedIdentityData)...)
								},
							},
						}
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(schemaIdentityType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.String, "test-id-value"),
					}),
					Schema: testIdentitySchema,
				},
			},
		},
		"RawState-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				IdentitySchema: testIdentitySchema,
				Resource:       &testprovider.Resource{},
				Version:        0,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{},
		},
		"RawState-Unmarshal-and-ResourceIdentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id": "test-id-value",
				}),
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithUpgradeIdentity{
					Resource: &testprovider.Resource{},
					UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
						return map[int64]resource.IdentityUpgrader{
							0: {
								IdentityUpgrader: func(ctx context.Context, req resource.UpgradeResourceIdentityRequest, resp *resource.UpgradeResourceIdentityResponse) {
									RawStateValue, err := req.RawState.Unmarshal(tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"id": tftypes.String,
										},
									})

									if err != nil {
										resp.Diagnostics.AddError(
											"Unable to Unmarshal Prior Identity",
											err.Error(),
										)
										return
									}

									var RawState map[string]tftypes.Value

									if err := RawStateValue.As(&RawState); err != nil {
										resp.Diagnostics.AddError(
											"Unable to Convert Prior Identity",
											err.Error(),
										)
										return
									}

									ResourceIdentity := &tfsdk.ResourceIdentity{
										Raw: tftypes.NewValue(schemaIdentityType, map[string]tftypes.Value{
											"id": RawState["id"],
										}),
										Schema: testIdentitySchema,
									}

									if err != nil {
										resp.Diagnostics.AddError(
											"Unable to Convert Upgraded Identity",
											err.Error(),
										)
										return
									}

									resp.UpgradedIdentity = ResourceIdentity
								},
							},
						}
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(schemaIdentityType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.String, "test-id-value"),
					}),
					Schema: testIdentitySchema,
				},
			},
		},
		"RawState-JSON-and-ResourceIdentity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithUpgradeIdentity{
					Resource: &testprovider.Resource{},
					UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
						return map[int64]resource.IdentityUpgrader{
							0: {
								IdentityUpgrader: func(ctx context.Context, req resource.UpgradeResourceIdentityRequest, resp *resource.UpgradeResourceIdentityResponse) {
									var RawState struct {
										Id string `json:"id"`
									}

									if err := json.Unmarshal(req.RawState.JSON, &RawState); err != nil {
										resp.Diagnostics.AddError(
											"Unable to Unmarshal Prior Identity",
											err.Error(),
										)
										return
									}

									ResourceIdentity := tfsdk.ResourceIdentity{
										Raw: tftypes.NewValue(schemaIdentityType, map[string]tftypes.Value{
											"id": tftypes.NewValue(tftypes.String, RawState.Id),
										}),
										Schema: testIdentitySchema,
									}

									resp.UpgradedIdentity = &ResourceIdentity
								},
							},
						}
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(schemaIdentityType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.String, "test-id-value"),
					}),
					Schema: testIdentitySchema,
				},
			},
		},
		"ResourceType-UpgradeIdentity-not-implemented": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				IdentitySchema: testIdentitySchema,
				Resource:       &testprovider.Resource{},
				Version:        0,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Upgrade Resource Identity",
						"This resource was implemented without an UpgradeIdentity() method, "+
							"however Terraform was expecting an implementation for version 0 upgrade.\n\n"+
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"ResourceType-UpgradeIdentity-empty": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithUpgradeIdentity{
					Resource: &testprovider.Resource{},
					UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
						return nil
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Upgrade Resource Identity",
						"This resource was implemented with an UpgradeIdentity() method, "+
							"however Terraform was expecting an implementation for version 0 upgrade.\n\n"+
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"PriorSchema-incorrect": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                            "test-id-value",
					"optional_for_import_attribute": true,
				}),
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithUpgradeIdentity{
					Resource: &testprovider.Resource{},
					UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
						return map[int64]resource.IdentityUpgrader{
							0: {
								PriorSchema: &identityschema.Schema{
									Attributes: map[string]identityschema.Attribute{
										"id": schema.StringAttribute{
											Computed: true,
										},
										"optional_for_import_attribute": identityschema.Int64Attribute{ // Purposefully incorrect
											OptionalForImport: true,
										},
									},
								},
								IdentityUpgrader: func(ctx context.Context, req resource.UpgradeResourceIdentityRequest, resp *resource.UpgradeResourceIdentityResponse) {
									// Expect error before reaching this logic.
								},
							},
						}
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Read Previously Saved Identity for UpgradeResourceIdentity",
						"There was an error reading the saved resource Identity using the prior resource schema defined for version 0 upgrade.\n\n"+
							"Please report this to the provider developer:\n\n"+
							"AttributeName(\"optional_for_import_attribute\"): unsupported type bool sent as tftypes.Number",
					),
				},
			},
		},
		"PriorSchema-and-Identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id": "test-id-value",
				}),
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithUpgradeIdentity{
					Resource: &testprovider.Resource{},
					UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
						return map[int64]resource.IdentityUpgrader{
							0: {
								PriorSchema: &testIdentitySchema,
								IdentityUpgrader: func(ctx context.Context, req resource.UpgradeResourceIdentityRequest, resp *resource.UpgradeResourceIdentityResponse) {
									rawStateValue, err := req.RawState.Unmarshal(tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"id": tftypes.String,
										},
									})

									if err != nil {
										resp.Diagnostics.AddError(
											"Unable to Read Previously Saved Identity for UpgradeResourceIdentity",
											fmt.Sprintf("There was an error reading the saved resource Identity using the prior resource schema defined for version %d upgrade.\n\n", req.Version)+
												"Please report this to the provider developer:\n\n"+err.Error(),
										)
										return
									}
									rawValues := make(map[string]tftypes.Value)
									err = rawStateValue.As(&rawValues)
									if err != nil {
										resp.Diagnostics.AddError(
											"Unable to convert raw state value into prior identity struct",
											fmt.Sprintf("There was an error converting the raw state value into the prior resource identity struct for version %d upgrade.\n\n", req.Version)+
												"Please report this to the provider developer:\n\n"+err.Error(),
										)
										return
									}

									priorIdentityId := rawValues["id"]
									var id string
									if priorIdentityId.Type().Is(tftypes.String) {
										err := priorIdentityId.As(&id)
										if err != nil {
											resp.Diagnostics.AddError(
												"Unable to convert raw state id value into string",
												fmt.Sprintf("There was an error converting the raw state id value into string for version %d upgrade.\n\n", req.Version)+
													"Please report this to the provider developer:\n\n"+err.Error(),
											)
											return
										}
									}

									upgradedIdentityData := struct {
										Id string `tfsdk:"id"`
									}{
										Id: id,
									}

									resp.Diagnostics.Append(resp.UpgradedIdentity.Set(ctx, upgradedIdentityData)...)
								},
							},
						}
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(schemaIdentityType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.String, "test-id-value"),
					}),
					Schema: testIdentitySchema,
				},
			},
		},
		"UpgradedIdentity-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id": "test-id-value",
				}),
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithUpgradeIdentity{
					Resource: &testprovider.Resource{},
					UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
						return map[int64]resource.IdentityUpgrader{
							0: {
								IdentityUpgrader: func(ctx context.Context, req resource.UpgradeResourceIdentityRequest, resp *resource.UpgradeResourceIdentityResponse) {
									// Purposfully not setting resp.ResourceIdentity or resp.UpgradedIdentity
								},
							},
						}
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Missing Upgraded Resource Identity",
						"After attempting a resource Identity upgrade to version 0, the provider did not return any Identity data. "+
							"Preventing the unexpected loss of resource Identity data. "+
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					),
				},
			},
		},
		/*"Version-current-flatmap": { // TODO: See if we need to add these tests back
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.ResourceWithUpgradeIdentity{

									Resource: &testprovider.Resource{},
									UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
										return map[int64]resource.IdentityUpgrader{
											0: {
												IdentityUpgrader: func(ctx context.Context, req resource.UpgradeResourceIdentityRequest, resp *resource.UpgradeResourceIdentityResponse) {
													// Purposfully not setting resp.ResourceIdentity or resp.UpgradedIdentity
												},
											},
										}
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: &tfprotov6.RawState{
					Flatmap: map[string]string{
						"flatmap": "is not supported",
					},
				},
				IdentitySchema: testIdentitySchema,
				Resource:       &testprovider.ResourceWithUpgradeIdentity{},
				Version:        1, // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Read Previously Saved Identity for UpgradeResourceIdentity",
						"There was an error reading the saved resource Identity using the current resource schema.\n\n"+
							"If this resource Identity was last refreshed with Terraform CLI 0.11 and earlier, it must be refreshed or applied with an older provider version first. "+
							"If you manually modified the resource Identity, you will need to manually modify it to match the current resource schema. "+
							"Otherwise, please report this to the provider developer:\n\n"+
							"flatmap Identitys cannot be unmarshaled, only Identitys written by Terraform 0.12 and higher can be unmarshaled",
					),
				},
			},
		},
		"Version-current-json-match": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.ResourceWithUpgradeIdentity{
									Resource: &testprovider.Resource{},
									UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
										return map[int64]resource.IdentityUpgrader{
											0: {
												IdentityUpgrader: func(ctx context.Context, req resource.UpgradeResourceIdentityRequest, resp *resource.UpgradeResourceIdentityResponse) {
													// Purposfully not setting resp.ResourceIdentity or resp.UpgradedIdentity
												},
											},
										}
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id": "test-id-value",
				}),
				IdentitySchema: testIdentitySchema,
				Resource:       &testprovider.ResourceWithUpgradeIdentity{},
				Version:        1, // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(schemaIdentityType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.String, "test-id-value"),
					}),
					Schema: testIdentitySchema,
				},
			},
		},
		"Version-current-json-mismatch": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                    "test-id-value",
					"nonexistent_attribute": "value",
				}),
				IdentitySchema: testIdentitySchema,
				Resource:       &testprovider.ResourceWithUpgradeIdentity{},
				Version:        1, // Must match current tfsdk.IdentitySchema version to trigger framework implementation
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfsdk.ResourceIdentity{
					Raw: tftypes.NewValue(schemaIdentityType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.String, "test-id-value"),
					}),
					Schema: testIdentitySchema,
				},
			},
		},*/
		"Version-not-implemented": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id": "test-id-value",
				}),
				IdentitySchema: testIdentitySchema,
				Resource: &testprovider.ResourceWithUpgradeIdentity{
					Resource: &testprovider.Resource{},
					UpgradeResourceIdentityMethod: func(ctx context.Context) map[int64]resource.IdentityUpgrader {
						return nil
					},
				},
				Version: 999,
			},
			expectedResponse: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Upgrade Resource Identity",
						"This resource was implemented with an UpgradeIdentity() method, "+
							"however Terraform was expecting an implementation for version 999 upgrade.\n\n"+
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.UpgradeResourceIdentityResponse{}
			testCase.server.UpgradeResourceIdentity(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
