// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerValidateDataSourceConfig(t *testing.T) {
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
		request          *fwserver.ValidateDataSourceConfigRequest
		expectedResponse *fwserver.ValidateDataSourceConfigResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateDataSourceConfigRequest{
				Config: &testConfig,
				DataSource: &testprovider.DataSource{
					SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
						resp.Schema = testSchema
					},
				},
			},
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{},
		},
		"request-config-AttributeValidator": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateDataSourceConfigRequest{
				Config: &testConfigAttributeValidator,
				DataSource: &testprovider.DataSource{
					SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
						resp.Schema = testSchemaAttributeValidator
					},
				},
			},
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{},
		},
		"request-config-AttributeValidator-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateDataSourceConfigRequest{
				Config: &testConfigAttributeValidatorError,
				DataSource: &testprovider.DataSource{
					SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
						resp.Schema = testSchemaAttributeValidatorError
					},
				},
			},
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"error summary",
						"error detail",
					),
				},
			},
		},
		"request-config-DataSourceWithConfigValidators": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateDataSourceConfigRequest{
				Config: &testConfig,
				DataSource: &testprovider.DataSourceWithConfigValidators{
					DataSource: &testprovider.DataSource{
						SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
							resp.Schema = testSchema
						},
					},
					ConfigValidatorsMethod: func(ctx context.Context) []datasource.ConfigValidator {
						return []datasource.ConfigValidator{
							&testprovider.DataSourceConfigValidator{
								ValidateDataSourceMethod: func(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
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
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{},
		},
		"request-config-DataSourceWithConfigValidators-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateDataSourceConfigRequest{
				Config: &testConfig,
				DataSource: &testprovider.DataSourceWithConfigValidators{
					DataSource: &testprovider.DataSource{
						SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
							resp.Schema = testSchema
						},
					},
					ConfigValidatorsMethod: func(ctx context.Context) []datasource.ConfigValidator {
						return []datasource.ConfigValidator{
							&testprovider.DataSourceConfigValidator{
								ValidateDataSourceMethod: func(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
									resp.Diagnostics.AddError("error summary 1", "error detail 1")
								},
							},
							&testprovider.DataSourceConfigValidator{
								ValidateDataSourceMethod: func(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
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
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{
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
		"request-config-DataSourceWithValidateConfig": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateDataSourceConfigRequest{
				Config: &testConfig,
				DataSource: &testprovider.DataSourceWithValidateConfig{
					DataSource: &testprovider.DataSource{
						SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
							resp.Schema = testSchema
						},
					},
					ValidateConfigMethod: func(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
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
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{},
		},
		"request-config-DataSourceWithValidateConfig-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateDataSourceConfigRequest{
				Config: &testConfig,
				DataSource: &testprovider.DataSourceWithValidateConfig{
					DataSource: &testprovider.DataSource{
						SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
							resp.Schema = testSchema
						},
					},
					ValidateConfigMethod: func(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{
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
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ValidateDataSourceConfigResponse{}
			testCase.server.ValidateDataSourceConfig(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
