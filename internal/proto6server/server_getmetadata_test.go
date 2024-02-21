// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestServerGetMetadata(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.GetMetadataRequest
		expectedError    error
		expectedResponse *tfprotov6.GetMetadataResponse
	}{
		"datasources": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source1"
										},
									}
								},
								func() datasource.DataSource {
									return &testprovider.DataSource{
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{
					{
						TypeName: "test_data_source1",
					},
					{
						TypeName: "test_data_source2",
					},
				},
				Functions: []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
		"datasources-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source"
										},
									}
								},
								func() datasource.DataSource {
									return &testprovider.DataSource{
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
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Duplicate Data Source Type Defined",
						Detail: "The test_data_source data source type name was returned for multiple data sources. " +
							"Data source type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
		"datasources-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = ""
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Data Source Type Name Missing",
						Detail: "The *testprovider.DataSource DataSource returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
		"functions": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(_ context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "function1"
										},
									}
								},
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "function2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Functions: []tfprotov6.FunctionMetadata{
					{
						Name: "function1",
					},
					{
						Name: "function2",
					},
				},
				Resources: []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
		"functions-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(_ context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction" // intentionally duplicate
										},
									}
								},
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction" // intentionally duplicate
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Duplicate Function Name Defined",
						Detail: "The testfunction function name was returned for multiple functions. " +
							"Function names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
		"functions-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(_ context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "" // intentionally empty
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Function Name Missing",
						Detail: "The *testprovider.Function Function returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
		"resources": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource1"
										},
									}
								},
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Functions:   []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{
					{
						TypeName: "test_resource1",
					},
					{
						TypeName: "test_resource2",
					},
				},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
		"resources-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Duplicate Resource Type Defined",
						Detail: "The test_resource resource type name was returned for multiple resources. " +
							"Resource type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
		"resources-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = ""
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Resource Type Name Missing",
						Detail: "The *testprovider.Resource Resource returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.GetMetadata(context.Background(), new(tfprotov6.GetMetadataRequest))

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			// Prevent false positives with random map access in testing
			sort.Slice(got.DataSources, func(i int, j int) bool {
				return got.DataSources[i].TypeName < got.DataSources[j].TypeName
			})

			sort.Slice(got.Functions, func(i int, j int) bool {
				return got.Functions[i].Name < got.Functions[j].Name
			})

			sort.Slice(got.Resources, func(i int, j int) bool {
				return got.Resources[i].TypeName < got.Resources[j].TypeName
			})

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
