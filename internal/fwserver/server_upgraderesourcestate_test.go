// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestServerUpgradeResourceState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"optional_attribute": schema.StringAttribute{
				Optional: true,
			},
			"required_attribute": schema.StringAttribute{
				Required: true,
			},
		},
		Version: 1, // Must be above 0
	}
	schemaType := testSchema.Type().TerraformType(ctx)

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
		"resource-configure-data": {
			server: &fwserver.Server{
				Provider:              &testprovider.Provider{},
				ResourceConfigureData: "test-provider-configure-value",
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithConfigureAndUpgradeState{
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
					UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
						return map[int64]resource.StateUpgrader{
							0: {
								PriorSchema: &schema.Schema{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Computed: true,
										},
										"optional_attribute": schema.BoolAttribute{
											Optional: true,
										},
										"required_attribute": schema.BoolAttribute{
											Required: true,
										},
									},
								},
								StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
									// In practice, the Configure method would save the
									// provider data to the Resource implementation and
									// use it here. The fact that Configure is able to
									// read the data proves this can work.

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
					Schema: testSchema,
				},
			},
		},
		"RawState-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				ResourceSchema: testSchema,
				Resource:       &testprovider.Resource{},
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
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithUpgradeState{
					Resource: &testprovider.Resource{},
					UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
						return map[int64]resource.StateUpgrader{
							0: {
								StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
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
					Schema: testSchema,
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
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithUpgradeState{
					Resource: &testprovider.Resource{},
					UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
						return map[int64]resource.StateUpgrader{
							0: {
								StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
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
					Schema: testSchema,
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
				ResourceSchema: testSchema,
				Resource:       &testprovider.Resource{},
				Version:        0,
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
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithUpgradeState{
					Resource: &testprovider.Resource{},
					UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
						return nil
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
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithUpgradeState{
					Resource: &testprovider.Resource{},
					UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
						return map[int64]resource.StateUpgrader{
							0: {
								PriorSchema: &schema.Schema{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Computed: true,
										},
										"optional_attribute": schema.Int64Attribute{ // Purposefully incorrect
											Optional: true,
										},
										"required_attribute": schema.Int64Attribute{ // Purposefully incorrect
											Required: true,
										},
									},
								},
								StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
									// Expect error before reaching this logic.
								},
							},
						}
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
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithUpgradeState{
					Resource: &testprovider.Resource{},
					UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
						return map[int64]resource.StateUpgrader{
							0: {
								PriorSchema: &schema.Schema{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Computed: true,
										},
										"optional_attribute": schema.BoolAttribute{
											Optional: true,
										},
										"required_attribute": schema.BoolAttribute{
											Required: true,
										},
									},
								},
								StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
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
					Schema: testSchema,
				},
			},
		},
		"PriorSchema-and-State-json-mismatch": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                    "test-id-value",
					"required_attribute":    true,
					"nonexistent_attribute": "value",
				}),
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithUpgradeState{
					Resource: &testprovider.Resource{},
					UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
						return map[int64]resource.StateUpgrader{
							0: {
								PriorSchema: &schema.Schema{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Computed: true,
										},
										"optional_attribute": schema.BoolAttribute{
											Optional: true,
										},
										"required_attribute": schema.BoolAttribute{
											Required: true,
										},
									},
								},
								StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
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
					Schema: testSchema,
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
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithUpgradeState{
					Resource: &testprovider.Resource{},
					UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
						return map[int64]resource.StateUpgrader{
							0: {
								StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
									// Purposfully not setting resp.DynamicValue or resp.State
								},
							},
						}
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
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.Resource{
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = testSchema
									},
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: &tfprotov6.RawState{
					Flatmap: map[string]string{
						"flatmap": "is not supported",
					},
				},
				ResourceSchema: testSchema,
				Resource:       &testprovider.Resource{},
				Version:        1, // Must match current tfsdk.Schema version to trigger framework implementation
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
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.Resource{
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = testSchema
									},
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": "true",
				}),
				ResourceSchema: testSchema,
				Resource:       &testprovider.Resource{},
				Version:        1, // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				UpgradedState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
				},
			},
		},
		"Version-current-json-mismatch": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                    "test-id-value",
					"required_attribute":    "true",
					"nonexistent_attribute": "value",
				}),
				ResourceSchema: testSchema,
				Resource:       &testprovider.Resource{},
				Version:        1, // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &fwserver.UpgradeResourceStateResponse{
				UpgradedState: &tfsdk.State{
					Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
						"optional_attribute": tftypes.NewValue(tftypes.String, nil),
						"required_attribute": tftypes.NewValue(tftypes.String, "true"),
					}),
					Schema: testSchema,
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
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithUpgradeState{
					Resource: &testprovider.Resource{},
					UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
						return nil
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
