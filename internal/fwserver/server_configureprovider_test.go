package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerConfigureProvider(t *testing.T) {
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

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *provider.ConfigureRequest
		expectedResponse *provider.ConfigureResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &provider.ConfigureResponse{},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
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
			request: &provider.ConfigureRequest{
				Config: testConfig,
			},
			expectedResponse: &provider.ConfigureResponse{},
		},
		"request-terraformversion": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return tfsdk.Schema{}, nil
					},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
						if req.TerraformVersion != "1.0.0" {
							resp.Diagnostics.AddError("Incorrect req.TerraformVersion", "expected 1.0.0, got "+req.TerraformVersion)
						}
					},
				},
			},
			request: &provider.ConfigureRequest{
				TerraformVersion: "1.0.0",
			},
			expectedResponse: &provider.ConfigureResponse{},
		},
		"response-datasourcedata": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return tfsdk.Schema{}, nil
					},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
						resp.DataSourceData = "test-provider-configure-value"
					},
				},
			},
			request: &provider.ConfigureRequest{},
			expectedResponse: &provider.ConfigureResponse{
				DataSourceData: "test-provider-configure-value",
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return tfsdk.Schema{}, nil
					},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			request: &provider.ConfigureRequest{},
			expectedResponse: &provider.ConfigureResponse{
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
			},
		},
		"response-resourcedata": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return tfsdk.Schema{}, nil
					},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
						resp.ResourceData = "test-provider-configure-value"
					},
				},
			},
			request: &provider.ConfigureRequest{},
			expectedResponse: &provider.ConfigureResponse{
				ResourceData: "test-provider-configure-value",
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &provider.ConfigureResponse{}
			testCase.server.ConfigureProvider(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.server.DataSourceConfigureData, testCase.expectedResponse.DataSourceData); diff != "" {
				t.Errorf("unexpected server.DataSourceConfigureData difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.server.ResourceConfigureData, testCase.expectedResponse.ResourceData); diff != "" {
				t.Errorf("unexpected server.ResourceConfigureData difference: %s", diff)
			}
		})
	}
}
