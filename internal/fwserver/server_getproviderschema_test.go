package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerGetProviderSchema(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.GetProviderSchemaRequest
		expectedResponse *fwserver.GetProviderSchemaResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{},
				Provider:          &tfsdk.Schema{},
				ResourceSchemas:   map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
						return []func() datasource.DataSource{
							func() datasource.DataSource {
								return &testprovider.DataSourceWithGetSchemaAndMetadata{
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
									MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = "test_data_source1"
									},
								}
							},
							func() datasource.DataSource {
								return &testprovider.DataSourceWithGetSchemaAndMetadata{
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
									MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = "test_data_source2"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source1": tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test1": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
					"test_data_source2": tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test2": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources-duplicate-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
						return []func() datasource.DataSource{
							func() datasource.DataSource {
								return &testprovider.DataSourceWithGetSchemaAndMetadata{
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
									MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = "test_data_source"
									},
								}
							},
							func() datasource.DataSource {
								return &testprovider.DataSourceWithGetSchemaAndMetadata{
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
									MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = "test_data_source"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Data Source Type Defined",
						"The test_data_source data source type name was returned for multiple data sources. "+
							"Data source type names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources-empty-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
						return []func() datasource.DataSource{
							func() datasource.DataSource {
								return &testprovider.DataSourceWithMetadata{
									MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = ""
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Data Source Type Name Missing",
						"The *testprovider.DataSourceWithMetadata DataSource returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources-missing-schema": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
						return []func() datasource.DataSource{
							func() datasource.DataSource {
								return &testprovider.DataSourceWithMetadata{
									MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = "test_data_source"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Data Source Schema Missing",
						"The *testprovider.DataSourceWithMetadata DataSource in the provider is missing the GetSchema method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources-missing-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Data Source Type Name Missing",
						"The *testprovider.DataSourceWithGetSchema DataSource in the provider DataSources method results is missing the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-DataSources-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithMetadata{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSourceWithGetSchemaAndMetadata{
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
										MetadataMethod: func(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = req.ProviderTypeName + "_data_source"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"testprovidertype_data_source": tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"datasourceschemas-GetDataSources": {
			server: &fwserver.Server{
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source1": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test1": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
					"test_data_source2": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test2": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"provider": {
			server: &fwserver.Server{
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{},
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Required: true,
							Type:     types.StringType,
						},
					},
				},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"providermeta": {
			server: &fwserver.Server{
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{},
				Provider:          &tfsdk.Schema{},
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Required: true,
							Type:     types.StringType,
						},
					},
				},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-GetResources": {
			server: &fwserver.Server{
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{},
				Provider:          &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource1": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test1": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
					"test_resource2": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test2": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.ResourceWithGetSchemaAndMetadata{
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
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource1"
									},
								}
							},
							func() resource.Resource {
								return &testprovider.ResourceWithGetSchemaAndMetadata{
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
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource2"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{},
				Provider:          &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource1": tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test1": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
					"test_resource2": tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test2": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources-duplicate-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.ResourceWithGetSchemaAndMetadata{
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
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource"
									},
								}
							},
							func() resource.Resource {
								return &testprovider.ResourceWithGetSchemaAndMetadata{
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
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Resource Type Defined",
						"The test_resource resource type name was returned for multiple resources. "+
							"Resource type names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: nil,
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources-empty-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.ResourceWithMetadata{
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = ""
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Resource Type Name Missing",
						"The *testprovider.ResourceWithMetadata Resource returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: nil,
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources-missing-schema": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.ResourceWithMetadata{
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_data_source"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Resource Schema Missing",
						"The *testprovider.ResourceWithMetadata Resource in the provider is missing the GetSchema method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: nil,
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources-missing-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Resource Type Name Missing",
						"The *testprovider.ResourceWithGetSchema Resource in the provider Resources method results is missing the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: nil,
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
		"resourceschemas-Resources-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithMetadata{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithGetSchemaAndMetadata{
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
										MetadataMethod: func(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = req.ProviderTypeName + "_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{},
				Provider:          &tfsdk.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{
					"testprovidertype_resource": tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
					PlanDestroy: true,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.GetProviderSchemaResponse{}
			testCase.server.GetProviderSchema(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
