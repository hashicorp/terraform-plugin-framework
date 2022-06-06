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

func TestServerReadDataSource(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testConfigDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, nil),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testEmptyDynamicValue := testNewDynamicValue(t, tftypes.Object{}, nil)

	testStateDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testSchema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test_computed": {
				Computed: true,
				Type:     types.StringType,
			},
			"test_required": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.ReadDataSourceRequest
		expectedError    error
		expectedResponse *tfprotov6.ReadDataSourceResponse
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
			request: &tfprotov6.ReadDataSourceRequest{
				Config:   testEmptyDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ReadDataSourceResponse{
				State: testEmptyDynamicValue,
			},
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
										return &testprovider.DataSource{
											ReadMethod: func(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
												var config struct {
													TestComputed types.String `tfsdk:"test_computed"`
													TestRequired types.String `tfsdk:"test_required"`
												}

												resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

												if config.TestRequired.Value != "test-config-value" {
													resp.Diagnostics.AddError("unexpected req.Config value: %s", config.TestRequired.Value)
												}
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ReadDataSourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ReadDataSourceResponse{
				State: testConfigDynamicValue,
			},
		},
		"request-providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithProviderMeta{
						Provider: &testprovider.Provider{
							GetDataSourcesMethod: func(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
								return map[string]tfsdk.DataSourceType{
									"test_data_source": &testprovider.DataSourceType{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{}, nil
										},
										NewDataSourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
											return &testprovider.DataSource{
												ReadMethod: func(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
													var config struct {
														TestComputed types.String `tfsdk:"test_computed"`
														TestRequired types.String `tfsdk:"test_required"`
													}

													resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &config)...)

													if config.TestRequired.Value != "test-config-value" {
														resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", config.TestRequired.Value)
													}
												},
											}, nil
										},
									},
								}, nil
							},
						},
						GetMetaSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return testSchema, nil
						},
					},
				},
			},
			request: &tfprotov6.ReadDataSourceRequest{
				Config:       testEmptyDynamicValue,
				ProviderMeta: testConfigDynamicValue,
				TypeName:     "test_data_source",
			},
			expectedResponse: &tfprotov6.ReadDataSourceResponse{
				State: testEmptyDynamicValue,
			},
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
										return &testprovider.DataSource{
											ReadMethod: func(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
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
			request: &tfprotov6.ReadDataSourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ReadDataSourceResponse{
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
				State: testConfigDynamicValue,
			},
		},
		"response-state": {
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
										return &testprovider.DataSource{
											ReadMethod: func(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
												var data struct {
													TestComputed types.String `tfsdk:"test_computed"`
													TestRequired types.String `tfsdk:"test_required"`
												}

												resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

												data.TestComputed = types.String{Value: "test-state-value"}

												resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ReadDataSourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ReadDataSourceResponse{
				State: testStateDynamicValue,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ReadDataSource(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
