package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	testConfigValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, nil),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
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

	testConfig := &tfsdk.Config{
		Raw:    testConfigValue,
		Schema: testSchema,
	}

	testStateUnchanged := &tfsdk.State{
		Raw:    testConfigValue,
		Schema: testSchema,
	}

	testState := &tfsdk.State{
		Raw:    testStateValue,
		Schema: testSchema,
	}

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.ReadDataSourceRequest
		expectedResponse *fwserver.ReadDataSourceResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSourceType: &testprovider.DataSourceType{
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
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State: testStateUnchanged,
			},
		},
		"request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSourceType: &testprovider.DataSourceType{
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

								resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &config)...)

								if config.TestRequired.Value != "test-config-value" {
									resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", config.TestRequired.Value)
								}
							},
						}, nil
					},
				},
				ProviderMeta: testConfig,
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State: testStateUnchanged,
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSourceType: &testprovider.DataSourceType{
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
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
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
				State: testStateUnchanged,
			},
		},
		"response-state": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadDataSourceRequest{
				Config:           testConfig,
				DataSourceSchema: testSchema,
				DataSourceType: &testprovider.DataSourceType{
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

								resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ReadDataSourceResponse{
				State: testState,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ReadDataSourceResponse{}
			testCase.server.ReadDataSource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
