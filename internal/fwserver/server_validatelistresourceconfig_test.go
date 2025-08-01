// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testvalidator"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerValidateListResourceConfig(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test": tftypes.String,
		},
	}

	testValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testConfig := tfsdk.Config{
		Raw:    testValue,
		Schema: testSchema,
	}

	testSchemaAttributeValidator := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					testvalidator.String{
						ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
							if req.ConfigValue.ValueString() != "test-value" {
								resp.Diagnostics.AddError("Incorrect req.AttributeConfig", "expected test-value, got "+req.ConfigValue.ValueString())
							}
						},
					},
				},
			},
		},
	}

	testConfigAttributeValidator := tfsdk.Config{
		Raw:    testValue,
		Schema: testSchemaAttributeValidator,
	}

	testSchemaAttributeValidatorError := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					testvalidator.String{
						ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
							resp.Diagnostics.AddAttributeError(req.Path, "error summary", "error detail")
						},
					},
				},
			},
		},
	}

	testConfigAttributeValidatorError := tfsdk.Config{
		Raw:    testValue,
		Schema: testSchemaAttributeValidatorError,
	}

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.ValidateListResourceConfigRequest
		expectedResponse *fwserver.ValidateListResourceConfigResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ValidateListResourceConfigResponse{},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateListResourceConfigRequest{
				Config: &testConfig,
				ListResource: &testprovider.ListResource{
					ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
						resp.Schema = testSchema
					},
				},
			},
			expectedResponse: &fwserver.ValidateListResourceConfigResponse{},
		},
		"request-config-AttributeValidator": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateListResourceConfigRequest{
				Config: &testConfigAttributeValidator,
				ListResource: &testprovider.ListResource{
					ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
						resp.Schema = testSchemaAttributeValidator
					},
				},
			},
			expectedResponse: &fwserver.ValidateListResourceConfigResponse{},
		},
		"request-config-AttributeValidator-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateListResourceConfigRequest{
				Config: &testConfigAttributeValidatorError,
				ListResource: &testprovider.ListResource{
					ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
						resp.Schema = testSchemaAttributeValidatorError
					},
				},
			},
			expectedResponse: &fwserver.ValidateListResourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"error summary",
						"error detail",
					),
				},
			},
		},
		"request-config-ListResourceWithConfigValidators": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateListResourceConfigRequest{
				Config: &testConfig,
				ListResource: &testprovider.ListResourceWithConfigValidators{
					ListResource: &testprovider.ListResource{
						ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
							resp.Schema = testSchema
						},
					},
					ConfigValidatorsMethod: func(ctx context.Context) []list.ConfigValidator {
						return []list.ConfigValidator{
							&testprovider.ListResourceConfigValidator{
								ValidateListResourceMethod: func(ctx context.Context, req list.ValidateConfigRequest, resp *list.ValidateConfigResponse) {
									var got types.String

									resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("test"), &got)...)

									if resp.Diagnostics.HasError() {
										return
									}

									if got.ValueString() != "test-value" {
										resp.Diagnostics.AddError("Incorrect req.Config", "expected test-value, got "+got.ValueString())
									}
								},
							},
						}
					},
				},
			},
			expectedResponse: &fwserver.ValidateListResourceConfigResponse{},
		},
		"request-config-ListResourceWithConfigValidators-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateListResourceConfigRequest{
				Config: &testConfig,
				ListResource: &testprovider.ListResourceWithConfigValidators{
					ListResource: &testprovider.ListResource{
						ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
							resp.Schema = testSchema
						},
					},
					ConfigValidatorsMethod: func(ctx context.Context) []list.ConfigValidator {
						return []list.ConfigValidator{
							&testprovider.ListResourceConfigValidator{
								ValidateListResourceMethod: func(ctx context.Context, req list.ValidateConfigRequest, resp *list.ValidateConfigResponse) {
									resp.Diagnostics.AddError("error summary 1", "error detail 1")
								},
							},
							&testprovider.ListResourceConfigValidator{
								ValidateListResourceMethod: func(ctx context.Context, req list.ValidateConfigRequest, resp *list.ValidateConfigResponse) {
									// Intentionally set diagnostics instead of add/append.
									// The framework should not overwrite existing diagnostics.
									// Reference: https://github.com/hashicorp/terraform-plugin-framework-validators/pull/94
									resp.Diagnostics = diag.Diagnostics{
										diag.NewErrorDiagnostic("error summary 2", "error detail 2"),
									}
								},
							},
						}
					},
				},
			},
			expectedResponse: &fwserver.ValidateListResourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"error summary 1",
						"error detail 1",
					),
					diag.NewErrorDiagnostic(
						"error summary 2",
						"error detail 2",
					),
				}},
		},
		"request-config-ListResourceWithValidateConfig": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateListResourceConfigRequest{
				Config: &testConfig,
				ListResource: &testprovider.ListResourceWithValidateConfig{
					ListResource: &testprovider.ListResource{
						ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
							resp.Schema = testSchema
						},
					},
					ValidateConfigMethod: func(ctx context.Context, req list.ValidateConfigRequest, resp *list.ValidateConfigResponse) {
						var got types.String

						resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("test"), &got)...)

						if resp.Diagnostics.HasError() {
							return
						}

						if got.ValueString() != "test-value" {
							resp.Diagnostics.AddError("Incorrect req.Config", "expected test-value, got "+got.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.ValidateListResourceConfigResponse{},
		},
		"request-config-ListResourceWithValidateConfig-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateListResourceConfigRequest{
				Config: &testConfig,
				ListResource: &testprovider.ListResourceWithValidateConfig{
					ListResource: &testprovider.ListResource{
						ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
							resp.Schema = testSchema
						},
					},
					ValidateConfigMethod: func(ctx context.Context, req list.ValidateConfigRequest, resp *list.ValidateConfigResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.ValidateListResourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic(
						"warning summary",
						"warning detail",
					),
					diag.NewErrorDiagnostic(
						"error summary",
						"error detail",
					),
				}},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ValidateListResourceConfigResponse{}
			testCase.server.ValidateListResourceConfig(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
