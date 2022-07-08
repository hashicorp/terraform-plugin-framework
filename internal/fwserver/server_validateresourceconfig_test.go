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

func TestServerValidateResourceConfig(t *testing.T) {
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
		request          *fwserver.ValidateResourceConfigRequest
		expectedResponse *fwserver.ValidateResourceConfigResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ValidateResourceConfigResponse{},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateResourceConfigRequest{
				Config: &testConfig,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
				},
			},
			expectedResponse: &fwserver.ValidateResourceConfigResponse{},
		},
		"request-config-AttributeValidator": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateResourceConfigRequest{
				Config: &testConfigAttributeValidator,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchemaAttributeValidator, nil
					},
				},
			},
			expectedResponse: &fwserver.ValidateResourceConfigResponse{},
		},
		"request-config-AttributeValidator-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateResourceConfigRequest{
				Config: &testConfigAttributeValidatorError,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchemaAttributeValidatorError, nil
					},
				},
			},
			expectedResponse: &fwserver.ValidateResourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"error summary",
						"error detail",
					),
				},
			},
		},
		"request-config-ResourceWithConfigValidators": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateResourceConfigRequest{
				Config: &testConfig,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithConfigValidators{
							Resource: &testprovider.Resource{},
							ConfigValidatorsMethod: func(ctx context.Context) []tfsdk.ResourceConfigValidator {
								return []tfsdk.ResourceConfigValidator{
									&testprovider.ResourceConfigValidator{
										ValidateResourceMethod: func(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
											var got types.String

											resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("test"), &got)...)

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
			expectedResponse: &fwserver.ValidateResourceConfigResponse{},
		},
		"request-config-ResourceWithConfigValidators-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateResourceConfigRequest{
				Config: &testConfig,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithConfigValidators{
							Resource: &testprovider.Resource{},
							ConfigValidatorsMethod: func(ctx context.Context) []tfsdk.ResourceConfigValidator {
								return []tfsdk.ResourceConfigValidator{
									&testprovider.ResourceConfigValidator{
										ValidateResourceMethod: func(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
											resp.Diagnostics.AddError("error summary", "error detail")
										},
									},
								}
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ValidateResourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"error summary",
						"error detail",
					),
				}},
		},
		"request-config-ResourceWithValidateConfig": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateResourceConfigRequest{
				Config: &testConfig,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithValidateConfig{
							Resource: &testprovider.Resource{},
							ValidateConfigMethod: func(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
								var got types.String

								resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("test"), &got)...)

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
			expectedResponse: &fwserver.ValidateResourceConfigResponse{},
		},
		"request-config-ResourceWithValidateConfig-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ValidateResourceConfigRequest{
				Config: &testConfig,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
						return &testprovider.ResourceWithValidateConfig{
							Resource: &testprovider.Resource{},
							ValidateConfigMethod: func(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
								resp.Diagnostics.AddWarning("warning summary", "warning detail")
								resp.Diagnostics.AddError("error summary", "error detail")
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ValidateResourceConfigResponse{
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

			response := &fwserver.ValidateResourceConfigResponse{}
			testCase.server.ValidateResourceConfig(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
