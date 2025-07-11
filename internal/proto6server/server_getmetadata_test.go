// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/list"
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
		"actions": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action1"
										},
									}
								},
								func() action.Action {
									return &testprovider.Action{
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action2"
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
				Actions: []tfprotov6.ActionMetadata{
					{
						TypeName: "test_action1",
					},
					{
						TypeName: "test_action2",
					},
				},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources:      []tfprotov6.ListResourceMetadata{},
				Resources:          []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
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
				Actions: []tfprotov6.ActionMetadata{},
				DataSources: []tfprotov6.DataSourceMetadata{
					{
						TypeName: "test_data_source1",
					},
					{
						TypeName: "test_data_source2",
					},
				},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources:      []tfprotov6.ListResourceMetadata{},
				Resources:          []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
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
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				Actions:     []tfprotov6.ActionMetadata{},
				DataSources: []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{
					{
						TypeName: "test_ephemeral_resource1",
					},
					{
						TypeName: "test_ephemeral_resource2",
					},
				},
				Functions:     []tfprotov6.FunctionMetadata{},
				ListResources: []tfprotov6.ListResourceMetadata{},
				Resources:     []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
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
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				Actions:            []tfprotov6.ActionMetadata{},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions: []tfprotov6.FunctionMetadata{
					{
						Name: "function1",
					},
					{
						Name: "function2",
					},
				},
				ListResources: []tfprotov6.ListResourceMetadata{},
				Resources:     []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
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
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				Actions:            []tfprotov6.ActionMetadata{},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources: []tfprotov6.ListResourceMetadata{
					{
						TypeName: "test_list_resource1",
					},
					{
						TypeName: "test_list_resource2",
					},
				},
				Resources: []tfprotov6.ResourceMetadata{
					{
						TypeName: "test_list_resource1",
					},
					{
						TypeName: "test_list_resource2",
					},
				},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
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
			request: &tfprotov6.GetMetadataRequest{},
			expectedResponse: &tfprotov6.GetMetadataResponse{
				Actions:            []tfprotov6.ActionMetadata{},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources:      []tfprotov6.ListResourceMetadata{},
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
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
	}

	for name, testCase := range testCases {
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

			sort.Slice(got.EphemeralResources, func(i int, j int) bool {
				return got.EphemeralResources[i].TypeName < got.EphemeralResources[j].TypeName
			})

			sort.Slice(got.Functions, func(i int, j int) bool {
				return got.Functions[i].Name < got.Functions[j].Name
			})

			sort.Slice(got.ListResources, func(i int, j int) bool {
				return got.ListResources[i].TypeName < got.ListResources[j].TypeName
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
