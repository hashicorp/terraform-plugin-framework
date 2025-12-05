// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tfsdklogtest"
)

func TestServerGetResourceIdentitySchemas(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.GetResourceIdentitySchemasRequest
		expectedError    error
		expectedResponse *tfprotov6.GetResourceIdentitySchemasResponse
	}{
		"resource-identity-schemas": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithIdentity{
										Resource: &testprovider.Resource{
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource1"
											},
										},
										IdentitySchemaMethod: func(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
											resp.IdentitySchema = identityschema.Schema{
												Attributes: map[string]identityschema.Attribute{
													"test1": identityschema.StringAttribute{
														RequiredForImport: true,
													},
												},
											}
										},
									}
								},
								func() resource.Resource {
									return &testprovider.ResourceWithIdentity{
										Resource: &testprovider.Resource{
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource2"
											},
										},
										IdentitySchemaMethod: func(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
											resp.IdentitySchema = identityschema.Schema{
												Attributes: map[string]identityschema.Attribute{
													"test2": identityschema.BoolAttribute{
														RequiredForImport: true,
													},
												},
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.GetResourceIdentitySchemasRequest{},
			expectedResponse: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource1": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test1",
								RequiredForImport: true,
								Type:              tftypes.String,
							},
						},
					},
					"test_resource2": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test2",
								RequiredForImport: true,
								Type:              tftypes.Bool,
							},
						},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.GetResourceIdentitySchemas(context.Background(), new(tfprotov6.GetResourceIdentitySchemasRequest))

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}

func TestServerGetResourceIdentitySchemas_logging(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tfsdklogtest.RootLogger(context.Background(), &output)
	ctx = logging.InitContext(ctx)

	testServer := &Server{
		FrameworkServer: fwserver.Server{
			Provider: &testprovider.Provider{
				ResourcesMethod: func(ctx context.Context) []func() resource.Resource {
					return []func() resource.Resource{
						func() resource.Resource {
							return &testprovider.ResourceWithIdentity{
								Resource: &testprovider.Resource{
									MetadataMethod: func(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "examplecloud_thing"
									},
								},
								IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
									resp.IdentitySchema = identityschema.Schema{}
								},
							}
						},
					}
				},
			},
		},
	}

	_, err := testServer.GetResourceIdentitySchemas(ctx, new(tfprotov6.GetResourceIdentitySchemasRequest))

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
			"@message": "Checking ResourceTypes lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ProviderTypeName lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Calling provider defined Provider Metadata",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Called provider defined Provider Metadata",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Calling provider defined Provider Resources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Called provider defined Provider Resources",
			"@module":  "sdk.framework",
		},
		{
			"@level":           "trace",
			"@message":         "Found resource type",
			"@module":          "sdk.framework",
			"tf_resource_type": "examplecloud_thing",
		},
		{
			"@level":           "trace",
			"@message":         "Calling provider defined Resource IdentitySchema method",
			"@module":          "sdk.framework",
			"tf_resource_type": "examplecloud_thing",
		},
		{
			"@level":           "trace",
			"@message":         "Called provider defined Resource IdentitySchema method",
			"@module":          "sdk.framework",
			"tf_resource_type": "examplecloud_thing",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}
