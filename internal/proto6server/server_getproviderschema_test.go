package proto6server

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tfsdklogtest"
)

func TestServerGetProviderSchema(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.GetProviderSchemaRequest
		expectedError    error
		expectedResponse *tfprotov6.GetProviderSchemaResponse
	}{
		"datasourceschemas-DataSources": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test1": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
										TypeNameMethod: func(_ context.Context, _ datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
											resp.TypeName = "test_data_source1"
										},
									}
								},
								func() datasource.DataSource {
									return &testprovider.DataSourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test2": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
										TypeNameMethod: func(_ context.Context, _ datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
											resp.TypeName = "test_data_source2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source1": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test1",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_data_source2": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test2",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test1": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
										TypeNameMethod: func(_ context.Context, _ datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
											resp.TypeName = "test_data_source"
										},
									}
								},
								func() datasource.DataSource {
									return &testprovider.DataSourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test2": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
										TypeNameMethod: func(_ context.Context, _ datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
											resp.TypeName = "test_data_source"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Duplicate Data Source Type Defined",
						Detail: "The test_data_source data source type name was returned for multiple data sources. " +
							"Data source type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSourceWithTypeName{
										TypeNameMethod: func(_ context.Context, _ datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
											resp.TypeName = ""
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Data Source Type Name Missing",
						Detail: "The *testprovider.DataSourceWithTypeName DataSource returned an empty string from the TypeName method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources-missing-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSourceWithTypeName{
										TypeNameMethod: func(_ context.Context, _ datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
											resp.TypeName = "test_data_source"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Data Source Schema Missing",
						Detail: "The *testprovider.DataSourceWithTypeName DataSource in the provider is missing the GetSchema method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources-missing-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSourceWithGetSchema{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test1": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Data Source Type Name Missing",
						Detail: "The *testprovider.DataSourceWithGetSchema DataSource in the provider DataSources method results is missing the TypeName method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-GetDataSources": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithGetDataSources{
						//nolint:staticcheck // Internal implementation
						GetDataSourcesMethod: func(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
							return map[string]provider.DataSourceType{
								"test_data_source1": &testprovider.DataSourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return tfsdk.Schema{
											Attributes: map[string]tfsdk.Attribute{
												"test1": {
													Required: true,
													Type:     types.StringType,
												},
											},
										}, nil
									},
								},
								"test_data_source2": &testprovider.DataSourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return tfsdk.Schema{
											Attributes: map[string]tfsdk.Attribute{
												"test2": {
													Required: true,
													Type:     types.StringType,
												},
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source1": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test1",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_data_source2": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test2",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"provider": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return tfsdk.Schema{
								Attributes: map[string]tfsdk.Attribute{
									"test": {
										Required: true,
										Type:     types.StringType,
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test",
								Required: true,
								Type:     tftypes.String,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithMetaSchema{
						Provider: &testprovider.Provider{},
						GetMetaSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
							return tfsdk.Schema{
								Attributes: map[string]tfsdk.Attribute{
									"test": {
										Required: true,
										Type:     types.StringType,
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test",
								Required: true,
								Type:     tftypes.String,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test1": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
										TypeNameMethod: func(_ context.Context, _ resource.TypeNameRequest, resp *resource.TypeNameResponse) {
											resp.TypeName = "test_resource1"
										},
									}
								},
								func() resource.Resource {
									return &testprovider.ResourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test2": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
										TypeNameMethod: func(_ context.Context, _ resource.TypeNameRequest, resp *resource.TypeNameResponse) {
											resp.TypeName = "test_resource2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource1": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test1",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_resource2": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test2",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test1": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
										TypeNameMethod: func(_ context.Context, _ resource.TypeNameRequest, resp *resource.TypeNameResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
								func() resource.Resource {
									return &testprovider.ResourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test2": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
										TypeNameMethod: func(_ context.Context, _ resource.TypeNameRequest, resp *resource.TypeNameResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Duplicate Resource Type Defined",
						Detail: "The test_resource resource type name was returned for multiple resources. " +
							"Resource type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithTypeName{
										TypeNameMethod: func(_ context.Context, _ resource.TypeNameRequest, resp *resource.TypeNameResponse) {
											resp.TypeName = ""
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Resource Type Name Missing",
						Detail: "The *testprovider.ResourceWithTypeName Resource returned an empty string from the TypeName method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources-missing-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithTypeName{
										TypeNameMethod: func(_ context.Context, _ resource.TypeNameRequest, resp *resource.TypeNameResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Resource Schema Missing",
						Detail: "The *testprovider.ResourceWithTypeName Resource in the provider is missing the GetSchema method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources-missing-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithGetSchema{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{
												Attributes: map[string]tfsdk.Attribute{
													"test1": {
														Required: true,
														Type:     types.StringType,
													},
												},
											}, nil
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Resource Type Name Missing",
						Detail: "The *testprovider.ResourceWithGetSchema Resource in the provider Resources method results is missing the TypeName method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-GetResources": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithGetResources{
						//nolint:staticcheck // Internal implementation
						GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
							return map[string]provider.ResourceType{
								"test_resource1": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return tfsdk.Schema{
											Attributes: map[string]tfsdk.Attribute{
												"test1": {
													Required: true,
													Type:     types.StringType,
												},
											},
										}, nil
									},
								},
								"test_resource2": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return tfsdk.Schema{
											Attributes: map[string]tfsdk.Attribute{
												"test2": {
													Required: true,
													Type:     types.StringType,
												},
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource1": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test1",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_resource2": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test2",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.GetProviderSchema(context.Background(), new(tfprotov6.GetProviderSchemaRequest))

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}

func TestServerGetProviderSchema_logging(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tfsdklogtest.RootLogger(context.Background(), &output)
	ctx = logging.InitContext(ctx)

	testServer := &Server{
		FrameworkServer: fwserver.Server{
			Provider: &testprovider.Provider{},
		},
	}

	_, err := testServer.GetProviderSchema(ctx, new(tfprotov6.GetProviderSchemaRequest))

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	entries, err := tfsdklogtest.MultilineJSONDecode(&output)

	if err != nil {
		t.Fatalf("unable to read multiple line JSON: %s", err)
	}

	expectedEntries := []map[string]interface{}{
		{
			"@level":   "trace",
			"@message": "Checking ProviderSchema lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Calling provider defined Provider GetSchema",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Called provider defined Provider GetSchema",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ResourceSchemas lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ResourceTypes lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Calling provider defined Provider Resources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Called provider defined Provider Resources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ResourceTypes lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking DataSourceSchemas lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking DataSourceTypes lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Calling provider defined Provider DataSources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Called provider defined Provider DataSources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking DataSourceTypes lock",
			"@module":  "sdk.framework",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}
