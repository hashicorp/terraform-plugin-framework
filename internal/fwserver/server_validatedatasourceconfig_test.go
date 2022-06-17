package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	testSchema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testConfig := tfsdk.Config{
		Raw:    testValue,
		Schema: testSchema,
	}

	testSchemaAttributeValidator := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test": {
				Required: true,
				Type:     types.StringType,
				Validators: []tfsdk.AttributeValidator{
					&testprovider.AttributeValidator{
						ValidateMethod: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
							var got types.String

							resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.AttributeConfig, &got)...)

							if resp.Diagnostics.HasError() {
								return
							}

							if got.Value != "test-value" {
								resp.Diagnostics.AddError("Incorrect req.AttributeConfig", "expected test-value, got "+got.Value)
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

	testSchemaAttributeValidatorError := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test": {
				Required: true,
				Type:     types.StringType,
				Validators: []tfsdk.AttributeValidator{
					&testprovider.AttributeValidator{
						ValidateMethod: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
							resp.Diagnostics.AddAttributeError(req.AttributePath, "error summary", "error detail")
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
				DataSourceType: &testprovider.DataSourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
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
				DataSourceType: &testprovider.DataSourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchemaAttributeValidator, nil
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
				DataSourceType: &testprovider.DataSourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchemaAttributeValidatorError, nil
					},
				},
			},
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.RootPath("test"),
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
				DataSourceType: &testprovider.DataSourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewDataSourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
						return &testprovider.DataSourceWithConfigValidators{
							DataSource: &testprovider.DataSource{},
							ConfigValidatorsMethod: func(ctx context.Context) []tfsdk.DataSourceConfigValidator {
								return []tfsdk.DataSourceConfigValidator{
									&testprovider.DataSourceConfigValidator{
										ValidateMethod: func(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
											var got types.String

											resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.RootPath("test"), &got)...)

											if resp.Diagnostics.HasError() {
												return
											}

											if got.Value != "test-value" {
												resp.Diagnostics.AddError("Incorrect req.Config", "expected test-value, got "+got.Value)
											}
										},
									},
								}
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{},
		},
		"request-config-DataSourceWithConfigValidators-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateDataSourceConfigRequest{
				Config: &testConfig,
				DataSourceType: &testprovider.DataSourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewDataSourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
						return &testprovider.DataSourceWithConfigValidators{
							DataSource: &testprovider.DataSource{},
							ConfigValidatorsMethod: func(ctx context.Context) []tfsdk.DataSourceConfigValidator {
								return []tfsdk.DataSourceConfigValidator{
									&testprovider.DataSourceConfigValidator{
										ValidateMethod: func(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
											resp.Diagnostics.AddError("error summary", "error detail")
										},
									},
								}
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ValidateDataSourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"error summary",
						"error detail",
					),
				}},
		},
		"request-config-DataSourceWithValidateConfig": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateDataSourceConfigRequest{
				Config: &testConfig,
				DataSourceType: &testprovider.DataSourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewDataSourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
						return &testprovider.DataSourceWithValidateConfig{
							DataSource: &testprovider.DataSource{},
							ValidateConfigMethod: func(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
								var got types.String

								resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.RootPath("test"), &got)...)

								if resp.Diagnostics.HasError() {
									return
								}

								if got.Value != "test-value" {
									resp.Diagnostics.AddError("Incorrect req.Config", "expected test-value, got "+got.Value)
								}
							},
						}, nil
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
				DataSourceType: &testprovider.DataSourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewDataSourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
						return &testprovider.DataSourceWithValidateConfig{
							DataSource: &testprovider.DataSource{},
							ValidateConfigMethod: func(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
								resp.Diagnostics.AddWarning("warning summary", "warning detail")
								resp.Diagnostics.AddError("error summary", "error detail")
							},
						}, nil
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
