// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
				ActionSchemas:            map[string]actionschema.SchemaType{},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
				Provider:                 providerschema.Schema{},
				ResourceSchemas:          map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"actionschemas": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ActionsMethod: func(_ context.Context) []func() action.Action {
						return []func() action.Action{
							func() action.Action {
								return &testprovider.Action{
									SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
										resp.Schema = actionschema.UnlinkedSchema{
											Attributes: map[string]actionschema.Attribute{
												"test1": actionschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = "test_action1"
									},
								}
							},
							func() action.Action {
								return &testprovider.Action{
									SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
										resp.Schema = actionschema.UnlinkedSchema{
											Attributes: map[string]actionschema.Attribute{
												"test2": actionschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = "test_action2"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas: map[string]actionschema.SchemaType{
					"test_action1": actionschema.UnlinkedSchema{
						Attributes: map[string]actionschema.Attribute{
							"test1": actionschema.StringAttribute{
								Required: true,
							},
						},
					},
					"test_action2": actionschema.UnlinkedSchema{
						Attributes: map[string]actionschema.Attribute{
							"test2": actionschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
				Provider:                 providerschema.Schema{},
				ResourceSchemas:          map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				ProviderMeta:      nil,
				DataSourceSchemas: map[string]fwschema.Schema{},
				Diagnostics:       diag.Diagnostics{},
			},
		},
		"actionschemas-invalid-attribute-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ActionsMethod: func(_ context.Context) []func() action.Action {
						return []func() action.Action{
							func() action.Action {
								return &testprovider.Action{
									SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
										resp.Schema = actionschema.UnlinkedSchema{
											Attributes: map[string]actionschema.Attribute{
												"$": actionschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = "test_action1"
									},
								}
							},
							func() action.Action {
								return &testprovider.Action{
									SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
										resp.Schema = actionschema.UnlinkedSchema{
											Attributes: map[string]actionschema.Attribute{
												"test2": actionschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = "test_action2"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				Provider: providerschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Attribute/Block Name",
						"When validating the schema, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"\"$\" at schema path \"$\" is an invalid attribute/block name. "+
							"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
					),
				},
				ProviderMeta:             nil,
				ResourceSchemas:          map[string]fwschema.Schema{},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
			},
		},
		"actionschemas-duplicate-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ActionsMethod: func(_ context.Context) []func() action.Action {
						return []func() action.Action{
							func() action.Action {
								return &testprovider.Action{
									SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
										resp.Schema = actionschema.UnlinkedSchema{
											Attributes: map[string]actionschema.Attribute{
												"test1": actionschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = "test_action"
									},
								}
							},
							func() action.Action {
								return &testprovider.Action{
									SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
										resp.Schema = actionschema.UnlinkedSchema{
											Attributes: map[string]actionschema.Attribute{
												"test2": actionschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = "test_action"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Action Defined",
						"The test_action action type was returned for multiple actions. "+
							"Action types must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        providerschema.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				ProviderMeta:             nil,
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
			},
		},
		"actionschemas-empty-type-name": {
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Action Type Missing",
						"The *testprovider.Action Action returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        providerschema.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				ProviderMeta:             nil,
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
			},
		},
		"actionschemas-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					ActionsMethod: func(_ context.Context) []func() action.Action {
						return []func() action.Action{
							func() action.Action {
								return &testprovider.Action{
									SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
										resp.Schema = actionschema.UnlinkedSchema{
											Attributes: map[string]actionschema.Attribute{
												"test": actionschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
										resp.TypeName = req.ProviderTypeName + "_action"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas: map[string]actionschema.SchemaType{
					"testprovidertype_action": actionschema.UnlinkedSchema{
						Attributes: map[string]actionschema.Attribute{
							"test": actionschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
				Provider:                 providerschema.Schema{},
				ResourceSchemas:          map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasourceschemas": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
						return []func() datasource.DataSource{
							func() datasource.DataSource {
								return &testprovider.DataSource{
									SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
										resp.Schema = datasourceschema.Schema{
											Attributes: map[string]datasourceschema.Attribute{
												"test1": datasourceschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = "test_data_source1"
									},
								}
							},
							func() datasource.DataSource {
								return &testprovider.DataSource{
									SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
										resp.Schema = datasourceschema.Schema{
											Attributes: map[string]datasourceschema.Attribute{
												"test2": datasourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
				ActionSchemas: map[string]actionschema.SchemaType{},
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source1": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test1": datasourceschema.StringAttribute{
								Required: true,
							},
						},
					},
					"test_data_source2": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test2": datasourceschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
				Provider:                 providerschema.Schema{},
				ResourceSchemas:          map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasourceschemas-invalid-attribute-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
						return []func() datasource.DataSource{
							func() datasource.DataSource {
								return &testprovider.DataSource{
									SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
										resp.Schema = datasourceschema.Schema{
											Attributes: map[string]datasourceschema.Attribute{
												"$": datasourceschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = "test_data_source1"
									},
								}
							},
							func() datasource.DataSource {
								return &testprovider.DataSource{
									SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
										resp.Schema = datasourceschema.Schema{
											Attributes: map[string]datasourceschema.Attribute{
												"test2": datasourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
				Provider:        providerschema.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Attribute/Block Name",
						"When validating the schema, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"\"$\" at schema path \"$\" is an invalid attribute/block name. "+
							"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
					),
				},
			},
		},
		"datasourceschemas-duplicate-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
						return []func() datasource.DataSource{
							func() datasource.DataSource {
								return &testprovider.DataSource{
									SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
										resp.Schema = datasourceschema.Schema{
											Attributes: map[string]datasourceschema.Attribute{
												"test1": datasourceschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
										resp.TypeName = "test_data_source"
									},
								}
							},
							func() datasource.DataSource {
								return &testprovider.DataSource{
									SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
										resp.Schema = datasourceschema.Schema{
											Attributes: map[string]datasourceschema.Attribute{
												"test2": datasourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
				Provider:        providerschema.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasourceschemas-empty-type-name": {
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Data Source Type Name Missing",
						"The *testprovider.DataSource DataSource returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        providerschema.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasourceschemas-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
						return []func() datasource.DataSource{
							func() datasource.DataSource {
								return &testprovider.DataSource{
									SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
										resp.Schema = datasourceschema.Schema{
											Attributes: map[string]datasourceschema.Attribute{
												"test": datasourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas: map[string]actionschema.SchemaType{},
				DataSourceSchemas: map[string]fwschema.Schema{
					"testprovidertype_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test": datasourceschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
				Provider:                 providerschema.Schema{},
				ResourceSchemas:          map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralschema": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
						return []func() ephemeral.EphemeralResource{
							func() ephemeral.EphemeralResource {
								return &testprovider.EphemeralResource{
									SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
										resp.Schema = ephemeralschema.Schema{
											Attributes: map[string]ephemeralschema.Attribute{
												"test1": ephemeralschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
										resp.TypeName = "test_ephemeral_resource1"
									},
								}
							},
							func() ephemeral.EphemeralResource {
								return &testprovider.EphemeralResource{
									SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
										resp.Schema = ephemeralschema.Schema{
											Attributes: map[string]ephemeralschema.Attribute{
												"test2": ephemeralschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
										resp.TypeName = "test_ephemeral_resource2"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas:     map[string]actionschema.SchemaType{},
				DataSourceSchemas: map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource1": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test1": ephemeralschema.StringAttribute{
								Required: true,
							},
						},
					},
					"test_ephemeral_resource2": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test2": ephemeralschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				FunctionDefinitions: map[string]function.Definition{},
				ListResourceSchemas: map[string]fwschema.Schema{},
				Provider:            providerschema.Schema{},
				ResourceSchemas:     map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralschema-invalid-attribute-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
						return []func() ephemeral.EphemeralResource{
							func() ephemeral.EphemeralResource {
								return &testprovider.EphemeralResource{
									SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
										resp.Schema = ephemeralschema.Schema{
											Attributes: map[string]ephemeralschema.Attribute{
												"$": ephemeralschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
										resp.TypeName = "test_ephemeral_resource1"
									},
								}
							},
							func() ephemeral.EphemeralResource {
								return &testprovider.EphemeralResource{
									SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
										resp.Schema = ephemeralschema.Schema{
											Attributes: map[string]ephemeralschema.Attribute{
												"test2": ephemeralschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
										resp.TypeName = "test_ephemeral_resource2"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas:   map[string]fwschema.Schema{},
				FunctionDefinitions: map[string]function.Definition{},
				Provider:            providerschema.Schema{},
				ResourceSchemas:     map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Attribute/Block Name",
						"When validating the schema, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"\"$\" at schema path \"$\" is an invalid attribute/block name. "+
							"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
					),
				},
			},
		},
		"ephemeralschema-duplicate-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
						return []func() ephemeral.EphemeralResource{
							func() ephemeral.EphemeralResource {
								return &testprovider.EphemeralResource{
									SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
										resp.Schema = ephemeralschema.Schema{
											Attributes: map[string]ephemeralschema.Attribute{
												"test1": ephemeralschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
										resp.TypeName = "test_ephemeral_resource"
									},
								}
							},
							func() ephemeral.EphemeralResource {
								return &testprovider.EphemeralResource{
									SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
										resp.Schema = ephemeralschema.Schema{
											Attributes: map[string]ephemeralschema.Attribute{
												"test2": ephemeralschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
										resp.TypeName = "test_ephemeral_resource"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Ephemeral Resource Type Defined",
						"The test_ephemeral_resource ephemeral resource type name was returned for multiple ephemeral resources. "+
							"Ephemeral resource type names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				FunctionDefinitions: map[string]function.Definition{},
				Provider:            providerschema.Schema{},
				ResourceSchemas:     map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralschema-empty-type-name": {
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Ephemeral Resource Type Name Missing",
						"The *testprovider.EphemeralResource EphemeralResource returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				FunctionDefinitions: map[string]function.Definition{},
				Provider:            providerschema.Schema{},
				ResourceSchemas:     map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralschema-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
						return []func() ephemeral.EphemeralResource{
							func() ephemeral.EphemeralResource {
								return &testprovider.EphemeralResource{
									SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
										resp.Schema = ephemeralschema.Schema{
											Attributes: map[string]ephemeralschema.Attribute{
												"test": ephemeralschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
										resp.TypeName = req.ProviderTypeName + "_ephemeral_resource"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas:     map[string]actionschema.SchemaType{},
				DataSourceSchemas: map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"testprovidertype_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test": ephemeralschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				FunctionDefinitions: map[string]function.Definition{},
				ListResourceSchemas: map[string]fwschema.Schema{},
				Provider:            providerschema.Schema{},
				ResourceSchemas:     map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functiondefinitions": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{
					FunctionsMethod: func(_ context.Context) []func() function.Function {
						return []func() function.Function{
							func() function.Function {
								return &testprovider.Function{
									DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
										resp.Definition = function.Definition{
											Return: function.StringReturn{},
										}
									},
									MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
										resp.Name = "function1"
									},
								}
							},
							func() function.Function {
								return &testprovider.Function{
									DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
										resp.Definition = function.Definition{
											Return: function.StringReturn{},
										}
									},
									MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
										resp.Name = "function2"
									},
								}
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas:            map[string]actionschema.SchemaType{},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions: map[string]function.Definition{
					"function1": {
						Return: function.StringReturn{},
					},
					"function2": {
						Return: function.StringReturn{},
					},
				},
				ListResourceSchemas: map[string]fwschema.Schema{},
				Provider:            providerschema.Schema{},
				ResourceSchemas:     map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functiondefinitions-invalid-definition": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{
					FunctionsMethod: func(_ context.Context) []func() function.Function {
						return []func() function.Function{
							func() function.Function {
								return &testprovider.Function{
									DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
										resp.Definition = function.Definition{
											Return: nil, // intentional
										}
									},
									MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
										resp.Name = "function1"
									},
								}
							},
							func() function.Function {
								return &testprovider.Function{
									DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
										resp.Definition = function.Definition{
											Return: function.StringReturn{},
										}
									},
									MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
										resp.Name = "function2"
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
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Function \"function1\" - Definition Return field is undefined",
					),
				},
				FunctionDefinitions: nil,
				Provider:            providerschema.Schema{},
				ResourceSchemas:     map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functiondefinitions-duplicate-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{
					FunctionsMethod: func(_ context.Context) []func() function.Function {
						return []func() function.Function{
							func() function.Function {
								return &testprovider.Function{
									DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
										resp.Definition = function.Definition{
											Return: function.StringReturn{},
										}
									},
									MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
										resp.Name = "testfunction" // intentionally duplicate
									},
								}
							},
							func() function.Function {
								return &testprovider.Function{
									DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
										resp.Definition = function.Definition{
											Return: function.StringReturn{},
										}
									},
									MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
										resp.Name = "testfunction" // intentionally duplicate
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
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Function Name Defined",
						"The testfunction function name was returned for multiple functions. "+
							"Function names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				FunctionDefinitions: nil,
				Provider:            providerschema.Schema{},
				ResourceSchemas:     map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functiondefinitions-empty-name": {
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Function Name Missing",
						"The *testprovider.Function Function returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				FunctionDefinitions: nil,
				Provider:            providerschema.Schema{},
				ResourceSchemas:     map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"listresource-schemas": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
						return []func() list.ListResource{
							func() list.ListResource {
								return &testprovider.ListResource{
									ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
										resp.Schema = listschema.Schema{
											Attributes: map[string]listschema.Attribute{
												"test1": listschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
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
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = resourceschema.Schema{
											Attributes: map[string]resourceschema.Attribute{
												"test1": resourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
				ActionSchemas:            map[string]actionschema.SchemaType{},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas: map[string]fwschema.Schema{
					"test_resource": listschema.Schema{
						Attributes: map[string]listschema.Attribute{
							"test1": listschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				Provider: providerschema.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test1": resourceschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"listresource-schemas-invalid-attribute-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
						return []func() list.ListResource{
							func() list.ListResource {
								return &testprovider.ListResource{
									ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
										resp.Schema = listschema.Schema{
											Attributes: map[string]listschema.Attribute{
												"$filter": listschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
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
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = resourceschema.Schema{
											Attributes: map[string]resourceschema.Attribute{
												"name": resourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
				Provider:                 providerschema.Schema{},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"name": resourceschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Attribute/Block Name",
						"When validating the schema, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"\"$filter\" at schema path \"$filter\" is an invalid attribute/block name. "+
							"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
					),
				},
			},
		},
		"provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
						resp.Schema = providerschema.Schema{
							Attributes: map[string]providerschema.Attribute{
								"test": providerschema.StringAttribute{
									Required: true,
								},
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas:            map[string]actionschema.SchemaType{},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test": providerschema.StringAttribute{
							Required: true,
						},
					},
				},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"provider-invalid-attribute-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
						resp.Schema = providerschema.Schema{
							Attributes: map[string]providerschema.Attribute{
								"$": providerschema.StringAttribute{
									Required: true,
								},
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Attribute/Block Name",
						"When validating the schema, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"\"$\" at schema path \"$\" is an invalid attribute/block name. "+
							"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
					),
				},
			},
		},
		"providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithMetaSchema{
					Provider: &testprovider.Provider{},
					MetaSchemaMethod: func(_ context.Context, _ provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
						resp.Schema = metaschema.Schema{
							Attributes: map[string]metaschema.Attribute{
								"test": metaschema.StringAttribute{
									Required: true,
								},
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas:            map[string]actionschema.SchemaType{},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
				Provider:                 providerschema.Schema{},
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test": metaschema.StringAttribute{
							Required: true,
						},
					},
				},
				ResourceSchemas: map[string]fwschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"providermeta-invalid-attribute-name": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithMetaSchema{
					Provider: &testprovider.Provider{},
					MetaSchemaMethod: func(_ context.Context, _ provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
						resp.Schema = metaschema.Schema{
							Attributes: map[string]metaschema.Attribute{
								"$": metaschema.StringAttribute{
									Required: true,
								},
							},
						}
					},
				},
			},
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				Provider: providerschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Attribute/Block Name",
						"When validating the schema, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"\"$\" at schema path \"$\" is an invalid attribute/block name. "+
							"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
					),
				},
			},
		},
		"resourceschemas": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.Resource{
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = resourceschema.Schema{
											Attributes: map[string]resourceschema.Attribute{
												"test1": resourceschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource1"
									},
								}
							},
							func() resource.Resource {
								return &testprovider.Resource{
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = resourceschema.Schema{
											Attributes: map[string]resourceschema.Attribute{
												"test2": resourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
				ActionSchemas:            map[string]actionschema.SchemaType{},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
				Provider:                 providerschema.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource1": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test1": resourceschema.StringAttribute{
								Required: true,
							},
						},
					},
					"test_resource2": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test2": resourceschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resourceschemas-invalid-attribute-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.Resource{
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = resourceschema.Schema{
											Attributes: map[string]resourceschema.Attribute{
												"$": resourceschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource1"
									},
								}
							},
							func() resource.Resource {
								return &testprovider.Resource{
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = resourceschema.Schema{
											Attributes: map[string]resourceschema.Attribute{
												"test2": resourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
				Provider: providerschema.Schema{},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Attribute/Block Name",
						"When validating the schema, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"\"$\" at schema path \"$\" is an invalid attribute/block name. "+
							"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
					),
				},
			},
		},
		"resourceschemas-duplicate-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.Resource{
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = resourceschema.Schema{
											Attributes: map[string]resourceschema.Attribute{
												"test1": resourceschema.StringAttribute{
													Required: true,
												},
											},
										}
									},
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource"
									},
								}
							},
							func() resource.Resource {
								return &testprovider.Resource{
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = resourceschema.Schema{
											Attributes: map[string]resourceschema.Attribute{
												"test2": resourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
				Provider:        providerschema.Schema{},
				ResourceSchemas: nil,
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resourceschemas-empty-type-name": {
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: nil,
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Resource Type Name Missing",
						"The *testprovider.Resource Resource returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				Provider:        providerschema.Schema{},
				ResourceSchemas: nil,
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resourceschemas-provider-type-name": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					MetadataMethod: func(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
						resp.TypeName = "testprovidertype"
					},
					ResourcesMethod: func(_ context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							func() resource.Resource {
								return &testprovider.Resource{
									SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
										resp.Schema = resourceschema.Schema{
											Attributes: map[string]resourceschema.Attribute{
												"test": resourceschema.StringAttribute{
													Required: true,
												},
											},
										}
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
			request: &fwserver.GetProviderSchemaRequest{},
			expectedResponse: &fwserver.GetProviderSchemaResponse{
				ActionSchemas:            map[string]actionschema.SchemaType{},
				DataSourceSchemas:        map[string]fwschema.Schema{},
				EphemeralResourceSchemas: map[string]fwschema.Schema{},
				FunctionDefinitions:      map[string]function.Definition{},
				ListResourceSchemas:      map[string]fwschema.Schema{},
				Provider:                 providerschema.Schema{},
				ResourceSchemas: map[string]fwschema.Schema{
					"testprovidertype_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test": resourceschema.StringAttribute{
								Required: true,
							},
						},
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

			response := &fwserver.GetProviderSchemaResponse{}
			testCase.server.GetProviderSchema(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
