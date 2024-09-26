// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerRenewEphemeralResource(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testEmptyDynamicValue := testNewDynamicValue(t, tftypes.Object{}, nil)

	testStateDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

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
		request          *tfprotov6.RenewEphemeralResourceRequest
		expectedError    error
		expectedResponse *tfprotov6.RenewEphemeralResourceResponse
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
			request: &tfprotov6.RenewEphemeralResourceRequest{
				State:    testEmptyDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov6.RenewEphemeralResourceResponse{},
		},
		"request-state": {
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
											var data struct {
												TestComputed types.String `tfsdk:"test_computed"`
												TestRequired types.String `tfsdk:"test_required"`
											}

											resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

											if data.TestRequired.ValueString() != "test-config-value" {
												resp.Diagnostics.AddError("unexpected req.State value for test_required: %s", data.TestRequired.ValueString())
											}

											if data.TestComputed.ValueString() != "test-state-value" {
												resp.Diagnostics.AddError("unexpected req.State value for test_computed: %s", data.TestComputed.ValueString())
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.RenewEphemeralResourceRequest{
				State:    testStateDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov6.RenewEphemeralResourceResponse{},
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
			request: &tfprotov6.RenewEphemeralResourceRequest{
				State:    testStateDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov6.RenewEphemeralResourceResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov6.DiagnosticSeverityError,
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
											resp.RenewAt = time.Date(2024, 8, 29, 6, 10, 32, 0, time.UTC)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.RenewEphemeralResourceRequest{
				State:    testEmptyDynamicValue,
				TypeName: "test_ephemeral_resource",
			},
			expectedResponse: &tfprotov6.RenewEphemeralResourceResponse{
				RenewAt: time.Date(2024, 8, 29, 6, 10, 32, 0, time.UTC),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

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
