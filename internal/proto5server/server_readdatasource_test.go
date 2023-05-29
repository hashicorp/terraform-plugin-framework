// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
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

	testProviderMetaDynamicValue := testNewDynamicValue(t,
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"test_optional": tftypes.String,
				"test_required": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"test_optional": tftypes.NewValue(tftypes.String, nil),
			"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
		},
	)

	testStateDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.ReadDataSourceRequest
		expectedError    error
		expectedResponse *tfprotov5.ReadDataSourceResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
											resp.Schema = schema.Schema{}
										},
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ReadDataSourceRequest{
				Config:   testEmptyDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov5.ReadDataSourceResponse{
				State: testEmptyDynamicValue,
			},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source"
										},
										ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
											var config struct {
												TestComputed types.String `tfsdk:"test_computed"`
												TestRequired types.String `tfsdk:"test_required"`
											}

											resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

											if config.TestRequired.ValueString() != "test-config-value" {
												resp.Diagnostics.AddError("unexpected req.Config value: %s", config.TestRequired.ValueString())
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ReadDataSourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov5.ReadDataSourceResponse{
				State: testConfigDynamicValue,
			},
		},
		"request-providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithMetaSchema{
						Provider: &testprovider.Provider{
							DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
								return []func() datasource.DataSource{
									func() datasource.DataSource {
										return &testprovider.DataSource{
											SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
												resp.Schema = schema.Schema{}
											},
											MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
												resp.TypeName = "test_data_source"
											},
											ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
												var config struct {
													TestOptional types.String `tfsdk:"test_optional"`
													TestRequired types.String `tfsdk:"test_required"`
												}

												resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &config)...)

												if config.TestRequired.ValueString() != "test-config-value" {
													resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", config.TestRequired.ValueString())
												}
											},
										}
									},
								}
							},
						},
						MetaSchemaMethod: func(_ context.Context, _ provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
							resp.Schema = metaschema.Schema{
								Attributes: map[string]metaschema.Attribute{
									"test_optional": metaschema.StringAttribute{
										Optional: true,
									},
									"test_required": metaschema.StringAttribute{
										Required: true,
									},
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ReadDataSourceRequest{
				Config:       testEmptyDynamicValue,
				ProviderMeta: testProviderMetaDynamicValue,
				TypeName:     "test_data_source",
			},
			expectedResponse: &tfprotov5.ReadDataSourceResponse{
				State: testEmptyDynamicValue,
			},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source"
										},
										ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
											resp.Diagnostics.AddWarning("warning summary", "warning detail")
											resp.Diagnostics.AddError("error summary", "error detail")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ReadDataSourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov5.ReadDataSourceResponse{
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
				State: testConfigDynamicValue,
			},
		},
		"response-state": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source"
										},
										ReadMethod: func(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
											var data struct {
												TestComputed types.String `tfsdk:"test_computed"`
												TestRequired types.String `tfsdk:"test_required"`
											}

											resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

											data.TestComputed = types.StringValue("test-state-value")

											resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ReadDataSourceRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov5.ReadDataSourceResponse{
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
