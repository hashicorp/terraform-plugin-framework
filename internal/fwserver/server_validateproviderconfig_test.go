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

func TestServerValidateProviderConfig(t *testing.T) {
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
		request          *fwserver.ValidateProviderConfigRequest
		expectedResponse *fwserver.ValidateProviderConfigResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ValidateProviderConfigResponse{},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
				},
			},
			request: &fwserver.ValidateProviderConfigRequest{
				Config: &testConfig,
			},
			expectedResponse: &fwserver.ValidateProviderConfigResponse{
				PreparedConfig: &testConfig,
			},
		},
		"request-config-AttributeValidator": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchemaAttributeValidator, nil
					},
				},
			},
			request: &fwserver.ValidateProviderConfigRequest{
				Config: &testConfigAttributeValidator,
			},
			expectedResponse: &fwserver.ValidateProviderConfigResponse{
				PreparedConfig: &testConfigAttributeValidator,
			},
		},
		"request-config-AttributeValidator-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchemaAttributeValidatorError, nil
					},
				},
			},
			request: &fwserver.ValidateProviderConfigRequest{
				Config: &testConfigAttributeValidatorError,
			},
			expectedResponse: &fwserver.ValidateProviderConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.RootPath("test"),
						"error summary",
						"error detail",
					),
				},
				PreparedConfig: &testConfigAttributeValidatorError,
			},
		},
		"request-config-ProviderWithConfigValidators": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithConfigValidators{
					Provider: &testprovider.Provider{
						GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return testSchema, nil
						},
					},
					ConfigValidatorsMethod: func(ctx context.Context) []tfsdk.ProviderConfigValidator {
						return []tfsdk.ProviderConfigValidator{
							&testprovider.ProviderConfigValidator{
								ValidateMethod: func(ctx context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
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
				},
			},
			request: &fwserver.ValidateProviderConfigRequest{
				Config: &testConfig,
			},
			expectedResponse: &fwserver.ValidateProviderConfigResponse{
				PreparedConfig: &testConfig,
			},
		},
		"request-config-ProviderWithConfigValidators-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithConfigValidators{
					Provider: &testprovider.Provider{
						GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return testSchema, nil
						},
					},
					ConfigValidatorsMethod: func(ctx context.Context) []tfsdk.ProviderConfigValidator {
						return []tfsdk.ProviderConfigValidator{
							&testprovider.ProviderConfigValidator{
								ValidateMethod: func(ctx context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
									resp.Diagnostics.AddError("error summary", "error detail")
								},
							},
						}
					},
				},
			},
			request: &fwserver.ValidateProviderConfigRequest{
				Config: &testConfig,
			},
			expectedResponse: &fwserver.ValidateProviderConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"error summary",
						"error detail",
					),
				},
				PreparedConfig: &testConfig,
			},
		},
		"request-config-ProviderWithValidateConfig": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithValidateConfig{
					Provider: &testprovider.Provider{
						GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return testSchema, nil
						},
					},
					ValidateConfigMethod: func(ctx context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
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
			},
			request: &fwserver.ValidateProviderConfigRequest{
				Config: &testConfig,
			},
			expectedResponse: &fwserver.ValidateProviderConfigResponse{
				PreparedConfig: &testConfig,
			},
		},
		"request-config-ProviderWithValidateConfig-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithValidateConfig{
					Provider: &testprovider.Provider{
						GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return testSchema, nil
						},
					},
					ValidateConfigMethod: func(ctx context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			request: &fwserver.ValidateProviderConfigRequest{
				Config: &testConfig,
			},
			expectedResponse: &fwserver.ValidateProviderConfigResponse{
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
				PreparedConfig: &testConfig,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ValidateProviderConfigResponse{}
			testCase.server.ValidateProviderConfig(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
