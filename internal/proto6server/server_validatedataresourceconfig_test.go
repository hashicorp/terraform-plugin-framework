package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerValidateDataResourceConfig(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test": tftypes.String,
		},
	}

	testValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testDynamicValue, err := tfprotov6.NewDynamicValue(testType, testValue)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testSchema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.ValidateDataResourceConfigRequest
		expectedError    error
		expectedResponse *tfprotov6.ValidateDataResourceConfigResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetDataSourcesMethod: func(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
							return map[string]tfsdk.DataSourceType{
								"test_data_source": &testprovider.DataSourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return tfsdk.Schema{}, nil
									},
									NewDataSourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
										return &testprovider.DataSource{}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ValidateDataResourceConfigRequest{
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ValidateDataResourceConfigResponse{},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetDataSourcesMethod: func(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
							return map[string]tfsdk.DataSourceType{
								"test_data_source": &testprovider.DataSourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewDataSourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
										return &testprovider.DataSource{}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ValidateDataResourceConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ValidateDataResourceConfigResponse{},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetDataSourcesMethod: func(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
							return map[string]tfsdk.DataSourceType{
								"test_data_source": &testprovider.DataSourceType{
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
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ValidateDataResourceConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ValidateDataResourceConfigResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ValidateDataResourceConfig(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
