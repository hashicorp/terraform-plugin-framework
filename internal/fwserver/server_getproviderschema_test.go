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
				DataSourceSchemas: map[string]*tfsdk.Schema{},
				Provider:          &tfsdk.Schema{},
				ResourceSchemas:   map[string]*tfsdk.Schema{},
			},
		},
		"datasourceschemas": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetDataSourcesMethod: func(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
						return map[string]tfsdk.DataSourceType{
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
				DataSourceSchemas: map[string]*tfsdk.Schema{
					"test_data_source1": {
						Attributes: map[string]tfsdk.Attribute{
							"test1": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
					"test_data_source2": {
						Attributes: map[string]tfsdk.Attribute{
							"test2": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
				Provider:        &tfsdk.Schema{},
				ResourceSchemas: map[string]*tfsdk.Schema{},
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
				DataSourceSchemas: map[string]*tfsdk.Schema{},
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Required: true,
							Type:     types.StringType,
						},
					},
				},
				ResourceSchemas: map[string]*tfsdk.Schema{},
			},
		},
		"providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithProviderMeta{
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
				DataSourceSchemas: map[string]*tfsdk.Schema{},
				Provider:          &tfsdk.Schema{},
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Required: true,
							Type:     types.StringType,
						},
					},
				},
				ResourceSchemas: map[string]*tfsdk.Schema{},
			},
		},
		"resourceschemas": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
						return map[string]tfsdk.ResourceType{
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
				DataSourceSchemas: map[string]*tfsdk.Schema{},
				Provider:          &tfsdk.Schema{},
				ResourceSchemas: map[string]*tfsdk.Schema{
					"test_resource1": {
						Attributes: map[string]tfsdk.Attribute{
							"test1": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
					"test_resource2": {
						Attributes: map[string]tfsdk.Attribute{
							"test2": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
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
