// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestServerGetMetadata(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.GetMetadataRequest
		expectedResponse *fwserver.GetMetadataResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Functions:          []fwserver.FunctionMetadata{},
				Resources:          []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"actions": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions: []fwserver.ActionMetadata{
					{
						TypeName: "test_action1",
					},
					{
						TypeName: "test_action2",
					},
				},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Functions:          []fwserver.FunctionMetadata{},
				Resources:          []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"actions-duplicate-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ActionsMethod: func(_ context.Context) []func() action.Action {
						return []func() action.Action{
							func() action.Action {
								return &testprovider.Action{
									MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = "test_action"
									},
								}
							},
							func() action.Action {
								return &testprovider.Action{
									MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = "test_action"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Action Defined",
						"The test_action action type was returned for multiple actions. "+
							"Action types must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"actions-empty-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ActionsMethod: func(_ context.Context) []func() action.Action {
						return []func() action.Action{
							func() action.Action {
								return &testprovider.Action{
									MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = ""
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Action Type Missing",
						"The *testprovider.Action Action returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"actions-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					ActionsMethod: func(_ context.Context) []func() action.Action {
						return []func() action.Action{
							func() action.Action {
								return &testprovider.Action{
									MetadataMethod: func(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = req.ProviderTypeName + "_action"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions: []fwserver.ActionMetadata{
					{
						TypeName: "testprovidertype_action",
					},
				},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Functions:          []fwserver.FunctionMetadata{},
				Resources:          []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasources": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions: []fwserver.ActionMetadata{},
				DataSources: []fwserver.DataSourceMetadata{
					{
						TypeName: "test_data_source1",
					},
					{
						TypeName: "test_data_source2",
					},
				},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Functions:          []fwserver.FunctionMetadata{},
				Resources:          []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasources-duplicate-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Data Source Type Defined",
						"The test_data_source data source type name was returned for multiple data sources. "+
							"Data source type names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasources-empty-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Data Source Type Name Missing",
						"The *testprovider.DataSource DataSource returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasources-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
						return []func() datasource.DataSource{
							func() datasource.DataSource {
								return &testprovider.DataSource{
									MetadataMethod: func(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = req.ProviderTypeName + "_data_source"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions: []fwserver.ActionMetadata{},
				DataSources: []fwserver.DataSourceMetadata{
					{
						TypeName: "testprovidertype_data_source",
					},
				},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Functions:          []fwserver.FunctionMetadata{},
				Resources:          []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralresources": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:     []fwserver.ActionMetadata{},
				DataSources: []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{
					{
						TypeName: "test_ephemeral_resource1",
					},
					{
						TypeName: "test_ephemeral_resource2",
					},
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralresources-duplicate-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Ephemeral Resource Type Defined",
						"The test_ephemeral_resource ephemeral resource type name was returned for multiple ephemeral resources. "+
							"Ephemeral resource type names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralresources-empty-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Ephemeral Resource Type Name Missing",
						"The *testprovider.EphemeralResource EphemeralResource returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralresources-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
						return []func() ephemeral.EphemeralResource{
							func() ephemeral.EphemeralResource {
								return &testprovider.EphemeralResource{
									MetadataMethod: func(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
										resp.TypeName = req.ProviderTypeName + "_ephemeral_resource"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:     []fwserver.ActionMetadata{},
				DataSources: []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{
					{
						TypeName: "testprovidertype_ephemeral_resource",
					},
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functions": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Functions: []fwserver.FunctionMetadata{
					{
						Name: "function1",
					},
					{
						Name: "function2",
					},
				},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functions-duplicate-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Function Name Defined",
						"The testfunction function name was returned for multiple functions. "+
							"Function names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functions-empty-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Function Name Missing",
						"The *testprovider.Function Function returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"listresources": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
						return []func() list.ListResource{
							func() list.ListResource {
								return &testprovider.ListResource{
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource_1"
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
										resp.TypeName = "test_resource_1"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics:        diag.Diagnostics{},
				Functions:          []fwserver.FunctionMetadata{},
				ListResources: []fwserver.ListResourceMetadata{
					{TypeName: "test_resource_1"},
				},
				Resources: []fwserver.ResourceMetadata{
					{TypeName: "test_resource_1"},
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"list-resources-empty-type-name": {
			server: &fwserver.Server{
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
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"ListResource Type Name Missing",
						"The *testprovider.ListResource ListResource returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"list-resources-duplicate-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
						return []func() list.ListResource{
							func() list.ListResource {
								return &testprovider.ListResource{
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource"
									},
								}
							},
							func() list.ListResource {
								return &testprovider.ListResource{
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource"
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
										resp.TypeName = "test_resource"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate ListResource Type Defined",
						"The test_resource ListResource type name was returned for multiple list resources. "+
							"ListResource type names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"list-resources-no-matching-managed-resource-type": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
						return []func() list.ListResource{
							func() list.ListResource {
								return &testprovider.ListResource{
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource_1"
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
										resp.TypeName = "test_resource_2"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"ListResource Type Defined without a Matching Managed Resource Type",
						"The test_resource_1 ListResource type name was returned, but no matching managed Resource type was defined. "+
							"If the matching managed Resource type not a framework resource either ProtoV5Schema and ProtoV5IdentitySchema must be specified in the RawV5Schemas method, "+
							"or ProtoV6Schema and ProtoV6IdentitySchema must be specified in the RawV6Schemas method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resources": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Functions:          []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{
					{
						TypeName: "test_resource1",
					},
					{
						TypeName: "test_resource2",
					},
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resources-duplicate-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Resource Type Defined",
						"The test_resource resource type name was returned for multiple resources. "+
							"Resource type names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resources-empty-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Resource Type Name Missing",
						"The *testprovider.Resource Resource returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Functions: []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resources-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.Resource{
									MetadataMethod: func(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = req.ProviderTypeName + "_resource"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetMetadataRequest{},
			expectedResponse: &fwserver.GetMetadataResponse{
				Actions:            []fwserver.ActionMetadata{},
				DataSources:        []fwserver.DataSourceMetadata{},
				EphemeralResources: []fwserver.EphemeralResourceMetadata{},
				Functions:          []fwserver.FunctionMetadata{},
				Resources: []fwserver.ResourceMetadata{
					{
						TypeName: "testprovidertype_resource",
					},
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
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

			response := &fwserver.GetMetadataResponse{}
			testCase.server.GetMetadata(context.Background(), testCase.request, response)

			// Prevent false positives with random map access in testing
			sort.Slice(response.Actions, func(i int, j int) bool {
				return response.Actions[i].TypeName < response.Actions[j].TypeName
			})

			sort.Slice(response.DataSources, func(i int, j int) bool {
				return response.DataSources[i].TypeName < response.DataSources[j].TypeName
			})

			sort.Slice(response.EphemeralResources, func(i int, j int) bool {
				return response.EphemeralResources[i].TypeName < response.EphemeralResources[j].TypeName
			})

			sort.Slice(response.Functions, func(i int, j int) bool {
				return response.Functions[i].Name < response.Functions[j].Name
			})

			sort.Slice(response.ListResources, func(i int, j int) bool {
				return response.ListResources[i].TypeName < response.ListResources[j].TypeName
			})

			sort.Slice(response.Resources, func(i int, j int) bool {
				return response.Resources[i].TypeName < response.Resources[j].TypeName
			})

			opts := cmp.Options{
				cmpopts.EquateEmpty(),
			}

			if diff := cmp.Diff(response, testCase.expectedResponse, opts...); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
