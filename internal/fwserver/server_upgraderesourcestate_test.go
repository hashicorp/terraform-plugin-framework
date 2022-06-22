package fwserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerUpgradeResourceState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"optional_attribute": {
				Type:     types.StringType,
				Optional: true,
			},
			"required_attribute": {
				Type:     types.StringType,
				Required: true,
			},
		},
		Version: 1, // Must be above 0
	}
	schemaType := schema.TerraformType(ctx)

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.UpgradeResourceStateRequest
		expectedResponse *fwserver.UpgradeResourceStateResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{},
		},
		"RawState-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				ResourceSchema: schema,
				ResourceType:   &testprovider.ResourceType{},
				Version:        0,
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{},
		},
		"RawState-Unmarshal-and-DynamicValue": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithUpgradeState{
							Resource: &testprovider.Resource{},
							UpgradeStateMethod: func(ctx context.Context) map[int64]tfsdk.ResourceStateUpgrader {
								return map[int64]tfsdk.ResourceStateUpgrader{
									0: {
										StateUpgrader: func(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
											rawStateValue, err := req.RawState.Unmarshal(tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"id":                 tftypes.String,
													"optional_attribute": tftypes.Bool,
													"required_attribute": tftypes.Bool,
												},
											})

											if err != nil {
												resp.Diagnostics.AddError(
													"Unable to Unmarshal Prior State",
													err.Error(),
												)
												return
											}

											var rawState map[string]tftypes.Value

											if err := rawStateValue.As(&rawState); err != nil {
												resp.Diagnostics.AddError(
													"Unable to Convert Prior State",
													err.Error(),
												)
												return
											}

											var optionalAttributeString *string

											if !rawState["optional_attribute"].IsNull() {
												var optionalAttribute bool

												if err := rawState["optional_attribute"].As(&optionalAttribute); err != nil {
													resp.Diagnostics.AddAttributeError(
														path.Root("optional_attribute"),
														"Unable to Convert Prior State",
														err.Error(),
													)
													return
												}

												v := fmt.Sprintf("%t", optionalAttribute)
												optionalAttributeString = &v
											}

											var requiredAttribute bool

											if err := rawState["required_attribute"].As(&requiredAttribute); err != nil {
												resp.Diagnostics.AddAttributeError(
													path.Root("required_attribute"),
													"Unable to Convert Prior State",
													err.Error(),
												)
												return
											}

											dynamicValue, err := tfprotov6.NewDynamicValue(
												schemaType,
												tftypes.NewValue(schemaType, map[string]tftypes.Value{
													"id":                 rawState["id"],
													"optional_attribute": tftypes.NewValue(tftypes.String, optionalAttributeString),
													"required_attribute": tftypes.NewValue(tftypes.String, fmt.Sprintf("%t", requiredAttribute)),
												}),
											)

											if err != nil {
												resp.Diagnostics.AddError(
													"Unable to Convert Upgraded State",
													err.Error(),
												)
												return
											}

											resp.DynamicValue = &dynamicValue
										},
									},
								}
							},
						}, nil
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				UpgradedState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: schema,
				},
			},
		},
		"RawState-JSON-and-DynamicValue": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithUpgradeState{
							Resource: &testprovider.Resource{},
							UpgradeStateMethod: func(ctx context.Context) map[int64]tfsdk.ResourceStateUpgrader {
								return map[int64]tfsdk.ResourceStateUpgrader{
									0: {
										StateUpgrader: func(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
											var rawState struct {
												Id                string `json:"id"`
												OptionalAttribute *bool  `json:"optional_attribute,omitempty"`
												RequiredAttribute bool   `json:"required_attribute"`
											}

											if err := json.Unmarshal(req.RawState.JSON, &rawState); err != nil {
												resp.Diagnostics.AddError(
													"Unable to Unmarshal Prior State",
													err.Error(),
												)
												return
											}

											var optionalAttribute *string

											if rawState.OptionalAttribute != nil {
												v := fmt.Sprintf("%t", *rawState.OptionalAttribute)
												optionalAttribute = &v
											}

											dynamicValue, err := tfprotov6.NewDynamicValue(
												schemaType,
												tftypes.NewValue(schemaType, map[string]tftypes.Value{
													"id":                 tftypes.NewValue(tftypes.String, rawState.Id),
													"optional_attribute": tftypes.NewValue(tftypes.String, optionalAttribute),
													"required_attribute": tftypes.NewValue(tftypes.String, fmt.Sprintf("%t", rawState.RequiredAttribute)),
												}),
											)

											if err != nil {
												resp.Diagnostics.AddError(
													"Unable to Create Upgraded State",
													err.Error(),
												)
												return
											}

											resp.DynamicValue = &dynamicValue
										},
									},
								}
							},
						}, nil
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				UpgradedState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: schema,
				},
			},
		},
		"ResourceType-UpgradeState-not-implemented": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.Resource{}, nil
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Upgrade Resource State",
						"This resource was implemented without an UpgradeState() method, "+
							"however Terraform was expecting an implementation for version 0 upgrade.\n\n"+
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"ResourceType-UpgradeState-empty": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithUpgradeState{
							Resource: &testprovider.Resource{},
							UpgradeStateMethod: func(ctx context.Context) map[int64]tfsdk.ResourceStateUpgrader {
								return nil
							},
						}, nil
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Upgrade Resource State",
						"This resource was implemented with an UpgradeState() method, "+
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
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithUpgradeState{
							Resource: &testprovider.Resource{},
							UpgradeStateMethod: func(ctx context.Context) map[int64]tfsdk.ResourceStateUpgrader {
								return map[int64]tfsdk.ResourceStateUpgrader{
									0: {
										PriorSchema: &tfsdk.Schema{
											Attributes: map[string]tfsdk.Attribute{
												"id": {
													Type:     types.StringType,
													Computed: true,
												},
												"optional_attribute": {
													Type:     types.Int64Type, // Purposefully incorrect
													Optional: true,
												},
												"required_attribute": {
													Type:     types.Int64Type, // Purposefully incorrect
													Required: true,
												},
											},
										},
										StateUpgrader: func(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
											// Expect error before reaching this logic.
										},
									},
								}
							},
						}, nil
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Read Previously Saved State for UpgradeResourceState",
						"There was an error reading the saved resource state using the prior resource schema defined for version 0 upgrade.\n\n"+
							"Please report this to the provider developer:\n\n"+
							"AttributeName(\"required_attribute\"): unsupported type bool sent as tftypes.Number",
					),
				},
			},
		},
		"PriorSchema-and-State": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithUpgradeState{
							Resource: &testprovider.Resource{},
							UpgradeStateMethod: func(ctx context.Context) map[int64]tfsdk.ResourceStateUpgrader {
								return map[int64]tfsdk.ResourceStateUpgrader{
									0: {
										PriorSchema: &tfsdk.Schema{
											Attributes: map[string]tfsdk.Attribute{
												"id": {
													Type:     types.StringType,
													Computed: true,
												},
												"optional_attribute": {
													Type:     types.BoolType,
													Optional: true,
												},
												"required_attribute": {
													Type:     types.BoolType,
													Required: true,
												},
											},
										},
										StateUpgrader: func(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
											var priorStateData struct {
												Id                string `tfsdk:"id"`
												OptionalAttribute *bool  `tfsdk:"optional_attribute"`
												RequiredAttribute bool   `tfsdk:"required_attribute"`
											}

											resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)

											if resp.Diagnostics.HasError() {
												return
											}

											upgradedStateData := struct {
												Id                string  `tfsdk:"id"`
												OptionalAttribute *string `tfsdk:"optional_attribute"`
												RequiredAttribute string  `tfsdk:"required_attribute"`
											}{
												Id:                priorStateData.Id,
												RequiredAttribute: fmt.Sprintf("%t", priorStateData.RequiredAttribute),
											}

											if priorStateData.OptionalAttribute != nil {
												v := fmt.Sprintf("%t", *priorStateData.OptionalAttribute)
												upgradedStateData.OptionalAttribute = &v
											}

											resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
										},
									},
								}
							},
						}, nil
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				UpgradedState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: schema,
				},
			},
		},
		"UpgradedState-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithUpgradeState{
							Resource: &testprovider.Resource{},
							UpgradeStateMethod: func(ctx context.Context) map[int64]tfsdk.ResourceStateUpgrader {
								return map[int64]tfsdk.ResourceStateUpgrader{
									0: {
										StateUpgrader: func(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
											// Purposfully not setting resp.DynamicValue or resp.State
										},
									},
								}
							},
						}, nil
					},
				},
				Version: 0,
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Missing Upgraded Resource State",
						"After attempting a resource state upgrade to version 0, the provider did not return any state data. "+
							"Preventing the unexpected loss of resource state data. "+
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"Version-current-flatmap": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
						return map[string]tfsdk.ResourceType{
							"test_resource": &testprovider.ResourceType{
								GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
									return schema, nil
								},
							},
						}, nil
					},
				},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: &tfprotov6.RawState{
					Flatmap: map[string]string{
						"flatmap": "is not supported",
					},
				},
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						// Framework should allow non-ResourceWithUpgradeState
						return &testprovider.Resource{}, nil
					},
				},
				Version: 1, // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Read Previously Saved State for UpgradeResourceState",
						"There was an error reading the saved resource state using the current resource schema.\n\n"+
							"If this resource state was last refreshed with Terraform CLI 0.11 and earlier, it must be refreshed or applied with an older provider version first. "+
							"If you manually modified the resource state, you will need to manually modify it to match the current resource schema. "+
							"Otherwise, please report this to the provider developer:\n\n"+
							"flatmap states cannot be unmarshaled, only states written by Terraform 0.12 and higher can be unmarshaled",
					),
				},
			},
		},
		"Version-current-json-match": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
						return map[string]tfsdk.ResourceType{
							"test_resource": &testprovider.ResourceType{
								GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
									return schema, nil
								},
							},
						}, nil
					},
				},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": "true",
				}),
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						// Framework should allow non-ResourceWithUpgradeState
						return &testprovider.Resource{}, nil
					},
				},
				Version: 1, // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				UpgradedState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: schema,
				},
			},
		},
		"Version-current-json-mismatch": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: &tfprotov6.RawState{
					JSON: []byte(`{"nonexistent_attribute":"value"}`),
				},
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						// Framework should allow non-ResourceWithUpgradeState
						return &testprovider.Resource{}, nil
					},
				},
				Version: 1, // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Read Previously Saved State for UpgradeResourceState",
						"There was an error reading the saved resource state using the current resource schema.\n\n"+
							"If this resource state was last refreshed with Terraform CLI 0.11 and earlier, it must be refreshed or applied with an older provider version first. "+
							"If you manually modified the resource state, you will need to manually modify it to match the current resource schema. "+
							"Otherwise, please report this to the provider developer:\n\n"+
							"ElementKeyValue(tftypes.String<unknown>): unsupported attribute \"nonexistent_attribute\"",
					),
				},
			},
		},
		"Version-not-implemented": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				ResourceSchema: schema,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return schema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithUpgradeState{
							Resource: &testprovider.Resource{},
							UpgradeStateMethod: func(ctx context.Context) map[int64]tfsdk.ResourceStateUpgrader {
								return nil
							},
						}, nil
					},
				},
				Version: 999,
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to Upgrade Resource State",
						"This resource was implemented with an UpgradeState() method, "+
							"however Terraform was expecting an implementation for version 999 upgrade.\n\n"+
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.UpgradeResourceStateResponse{}
			testCase.server.UpgradeResourceState(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
