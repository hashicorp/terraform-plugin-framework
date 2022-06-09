package proto5server

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tfsdklogtest"
)

func TestServerGetProviderSchema(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.GetProviderSchemaRequest
		expectedError    error
		expectedResponse *tfprotov5.GetProviderSchemaResponse
	}{
		"datasourceschemas": {
			server: &Server{
				FrameworkServer: fwserver.Server{
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
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov5.Schema{
					"test_data_source1": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test1",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_data_source2": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test2",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
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
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov5.Schema{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{
						Attributes: []*tfprotov5.SchemaAttribute{
							{
								Name:     "test",
								Required: true,
								Type:     tftypes.String,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
			},
		},
		"providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
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
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov5.Schema{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ProviderMeta: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{
						Attributes: []*tfprotov5.SchemaAttribute{
							{
								Name:     "test",
								Required: true,
								Type:     tftypes.String,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
			},
		},
		"resourceschemas": {
			server: &Server{
				FrameworkServer: fwserver.Server{
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
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov5.Schema{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{
					"test_resource1": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test1",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_resource2": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test2",
									Required: true,
									Type:     tftypes.String,
								},
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

			got, err := testCase.server.GetProviderSchema(context.Background(), new(tfprotov5.GetProviderSchemaRequest))

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

	_, err := testServer.GetProviderSchema(ctx, new(tfprotov5.GetProviderSchemaRequest))

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
			"@message": "Calling provider defined Provider GetResources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Called provider defined Provider GetResources",
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
			"@message": "Calling provider defined Provider GetDataSources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Called provider defined Provider GetDataSources",
			"@module":  "sdk.framework",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}
