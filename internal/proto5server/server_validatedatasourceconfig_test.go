package proto5server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
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

	testDynamicValue, err := tfprotov5.NewDynamicValue(testType, testValue)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov5.NewDynamicValue(): %s", err)
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
		request          *tfprotov5.ValidateDataSourceConfigRequest
		expectedError    error
		expectedResponse *tfprotov5.ValidateDataSourceConfigResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetDataSourcesMethod: func(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
							return map[string]provider.DataSourceType{
								"test_data_source": &testprovider.DataSourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return tfsdk.Schema{}, nil
									},
									NewDataSourceMethod: func(_ context.Context, _ provider.Provider) (datasource.DataSource, diag.Diagnostics) {
										return &testprovider.DataSource{}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov5.ValidateDataSourceConfigRequest{
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov5.ValidateDataSourceConfigResponse{},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetDataSourcesMethod: func(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
							return map[string]provider.DataSourceType{
								"test_data_source": &testprovider.DataSourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewDataSourceMethod: func(_ context.Context, _ provider.Provider) (datasource.DataSource, diag.Diagnostics) {
										return &testprovider.DataSource{}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov5.ValidateDataSourceConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov5.ValidateDataSourceConfigResponse{},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetDataSourcesMethod: func(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
							return map[string]provider.DataSourceType{
								"test_data_source": &testprovider.DataSourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewDataSourceMethod: func(_ context.Context, _ provider.Provider) (datasource.DataSource, diag.Diagnostics) {
										return &testprovider.DataSourceWithValidateConfig{
											DataSource: &testprovider.DataSource{},
											ValidateConfigMethod: func(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
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
			request: &tfprotov5.ValidateDataSourceConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov5.ValidateDataSourceConfigResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov5.DiagnosticSeverityError,
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

			got, err := testCase.server.ValidateDataSourceConfig(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
