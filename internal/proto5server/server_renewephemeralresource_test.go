// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func TestServerRenewEphemeralResource(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.RenewEphemeralResourceRequest
		expectedError    error
		expectedResponse *tfprotov5.RenewEphemeralResourceResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResourceWithRenew{
										EphemeralResource: &testprovider.EphemeralResource{
											SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
												resp.Schema = schema.Schema{}
											},
											MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
												resp.TypeName = "test_ephemeral_resource"
											},
										},
										RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.RenewEphemeralResourceRequest{
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov5.RenewEphemeralResourceResponse{},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResourceWithRenew{
										EphemeralResource: &testprovider.EphemeralResource{
											SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
												resp.Schema = testSchema
											},
											MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
												resp.TypeName = "test_ephemeral_resource"
											},
										},
										RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
											resp.Diagnostics.AddWarning("warning summary", "warning detail")
											resp.Diagnostics.AddError("error summary", "error detail")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.RenewEphemeralResourceRequest{
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov5.RenewEphemeralResourceResponse{
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
		"response-renew-at": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResourceWithRenew{
										EphemeralResource: &testprovider.EphemeralResource{
											SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
												resp.Schema = schema.Schema{}
											},
											MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
												resp.TypeName = "test_ephemeral_resource"
											},
										},
										RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
											resp.RenewAt = time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.RenewEphemeralResourceRequest{
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov5.RenewEphemeralResourceResponse{
				RenewAt: time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.RenewEphemeralResource(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
