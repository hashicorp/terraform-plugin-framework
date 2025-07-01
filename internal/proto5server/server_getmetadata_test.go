// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func TestServerGetMetadata(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.GetMetadataRequest
		expectedError    error
		expectedResponse *tfprotov5.GetMetadataResponse
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
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{
					{
						TypeName: "test_data_source1",
					},
					{
						TypeName: "test_data_source2",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
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
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Data Source Type Defined",
						Detail: "The test_data_source data source type name was returned for multiple data sources. " +
							"Data source type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
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
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Data Source Type Name Missing",
						Detail: "The *testprovider.DataSource DataSource returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralresources": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource1"
										},
									}
								},
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Functions:   []tfprotov5.FunctionMetadata{},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{
					{
						TypeName: "test_ephemeral_resource1",
					},
					{
						TypeName: "test_ephemeral_resource2",
					},
				},
				ListResources: []tfprotov5.ListResourceMetadata{},
				Resources:     []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralresources-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource"
										},
									}
								},
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Ephemeral Resource Type Defined",
						Detail: "The test_ephemeral_resource ephemeral resource type name was returned for multiple ephemeral resources. " +
							"Ephemeral resource type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralresources-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = ""
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Ephemeral Resource Type Name Missing",
						Detail: "The *testprovider.EphemeralResource EphemeralResource returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
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
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Functions: []tfprotov5.FunctionMetadata{
					{
						Name: "function1",
					},
					{
						Name: "function2",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
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
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Function Name Defined",
						Detail: "The testfunction function name was returned for multiple functions. " +
							"Function names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
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
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Function Name Missing",
						Detail: "The *testprovider.Function Function returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"listresources": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
							return []func() list.ListResource{
								func() list.ListResource {
									return &testprovider.ListResource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_list_resource1"
										},
									}
								},
								func() list.ListResource {
									return &testprovider.ListResource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_list_resource2"
										},
									}
								},
							}
						},
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_list_resource1"
										},
									}
								},
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_list_resource2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources:        []tfprotov5.DataSourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				ListResources: []tfprotov5.ListResourceMetadata{
					{
						TypeName: "test_list_resource1",
					},
					{
						TypeName: "test_list_resource2",
					},
				},
				Resources: []tfprotov5.ResourceMetadata{
					{
						TypeName: "test_list_resource1",
					},
					{
						TypeName: "test_list_resource2",
					},
				},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"listresources-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
							return []func() list.ListResource{
								func() list.ListResource {
									return &testprovider.ListResource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_list_resource"
										},
									}
								},
								func() list.ListResource {
									return &testprovider.ListResource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_list_resource"
										},
									}
								},
							}
						},
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_list_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate ListResource Type Defined",
						Detail: "The test_list_resource ListResource type name was returned for multiple list resources. " +
							"ListResource type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"listresources-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
							return []func() list.ListResource{
								func() list.ListResource {
									return &testprovider.ListResource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = ""
										},
									}
								},
							}
						},
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_list_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "ListResource Type Name Missing",
						Detail: "The *testprovider.ListResource ListResource returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"listresources-missing-resource-definition": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
							return []func() list.ListResource{
								func() list.ListResource {
									return &testprovider.ListResource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_list_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "ListResource Type Defined without a Matching Managed Resource Type",
						Detail: "The test_list_resource ListResource type name was returned, but no matching managed Resource type was defined. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
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
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources:        []tfprotov5.DataSourceMetadata{},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources: []tfprotov5.ResourceMetadata{
					{
						TypeName: "test_resource1",
					},
					{
						TypeName: "test_resource2",
					},
				},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
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
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Resource Type Defined",
						Detail: "The test_resource resource type name was returned for multiple resources. " +
							"Resource type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
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
			request: &tfprotov5.GetMetadataRequest{},
			expectedResponse: &tfprotov5.GetMetadataResponse{
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Resource Type Name Missing",
						Detail: "The *testprovider.Resource Resource returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.GetMetadata(context.Background(), new(tfprotov5.GetMetadataRequest))

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			// Prevent false positives with random map access in testing
			sort.Slice(got.DataSources, func(i int, j int) bool {
				return got.DataSources[i].TypeName < got.DataSources[j].TypeName
			})

			sort.Slice(got.EphemeralResources, func(i int, j int) bool {
				return got.EphemeralResources[i].TypeName < got.EphemeralResources[j].TypeName
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
