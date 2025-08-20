// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func TestServerValidateListResourceConfig(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test": tftypes.String,
		},
	}

	testValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testDynamicValue, err := tfprotov5.NewDynamicValue(testType, testValue)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.ValidateListResourceConfigRequest
		expectedError    error
		expectedResponse *tfprotov5.ValidateListResourceConfigResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
							return []func() list.ListResource{
								func() list.ListResource {
									return &testprovider.ListResource{
										MetadataMethod: func(_ context.Context, _ list.MetadataRequest, resp *list.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
										ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {},
									}
								},
							}
						},
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
											response.TypeName = "test_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ValidateListResourceConfigRequest{
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ValidateListResourceConfigResponse{},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
							return []func() list.ListResource{
								func() list.ListResource {
									return &testprovider.ListResource{
										ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ list.MetadataRequest, resp *list.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
							}
						},
						ResourcesMethod: func(ctx context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
											response.TypeName = "test_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ValidateListResourceConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ValidateListResourceConfigResponse{},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ListResourcesMethod: func(_ context.Context) []func() list.ListResource {
							return []func() list.ListResource{
								func() list.ListResource {
									return &testprovider.ListResourceWithValidateConfig{
										ListResource: &testprovider.ListResource{
											ListResourceConfigSchemaMethod: func(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ list.MetadataRequest, resp *list.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
										},
										ValidateConfigMethod: func(ctx context.Context, req list.ValidateConfigRequest, resp *list.ValidateConfigResponse) {
											resp.Diagnostics.AddWarning("warning summary", "warning detail")
											resp.Diagnostics.AddError("error summary", "error detail")
										},
									}
								},
							}
						},
						ResourcesMethod: func(ctx context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
											response.TypeName = "test_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ValidateListResourceConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ValidateListResourceConfigResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ValidateListResourceConfig(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
