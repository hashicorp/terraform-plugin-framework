// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
)

func TestServerGetResourceIdentitySchemas(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.GetResourceIdentitySchemasRequest
		expectedResponse *fwserver.GetResourceIdentitySchemasResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{},
			},
		},
		"resource-no-identity-schemas": {
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
			request: &fwserver.GetResourceIdentitySchemasRequest{},
			expectedResponse: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{},
			},
		},
		"resource-identity-schemas": {
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
								return &testprovider.ResourceWithIdentity{
									Resource: &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource2"
										},
									},
									IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
										resp.IdentitySchema = identityschema.Schema{
											Attributes: map[string]identityschema.Attribute{
												"test2": identityschema.StringAttribute{
													RequiredForImport: true,
												},
											},
										}
									},
								}
							},
							func() resource.Resource {
								return &testprovider.Resource{
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource3"
									},
								}
							},
							func() resource.Resource {
								return &testprovider.ResourceWithIdentity{
									Resource: &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource4"
										},
									},
									IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
										resp.IdentitySchema = identityschema.Schema{
											Attributes: map[string]identityschema.Attribute{
												"test4": identityschema.BoolAttribute{
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
			request: &fwserver.GetResourceIdentitySchemasRequest{},
			expectedResponse: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource2": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test2": identityschema.StringAttribute{
								RequiredForImport: true,
							},
						},
					},
					"test_resource4": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test4": identityschema.BoolAttribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
		},
		"resource-identity-schemas-invalid-attribute-name": {
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
								return &testprovider.ResourceWithIdentity{
									Resource: &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource2"
										},
									},
									IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
										resp.IdentitySchema = identityschema.Schema{
											Attributes: map[string]identityschema.Attribute{
												"$": identityschema.StringAttribute{
													RequiredForImport: true,
												},
											},
										}
									},
								}
							},
							func() resource.Resource {
								return &testprovider.Resource{
									MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
										resp.TypeName = "test_resource3"
									},
								}
							},
							func() resource.Resource {
								return &testprovider.ResourceWithIdentity{
									Resource: &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource4"
										},
									},
									IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
										resp.IdentitySchema = identityschema.Schema{
											Attributes: map[string]identityschema.Attribute{
												"test4": identityschema.BoolAttribute{
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
			request: &fwserver.GetResourceIdentitySchemasRequest{},
			expectedResponse: &fwserver.GetResourceIdentitySchemasResponse{
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
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.GetResourceIdentitySchemasResponse{}
			testCase.server.GetResourceIdentitySchemas(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
