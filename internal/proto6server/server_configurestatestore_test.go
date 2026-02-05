// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-framework/statestore/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerConfigureStateStore(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test": tftypes.String,
		},
	}

	testConfigDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testEmptyDynamicValue := testNewDynamicValue(t, tftypes.Object{}, nil)

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test": schema.StringAttribute{
				Required: true,
			},
		},
	}
	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.ConfigureStateStoreRequest
		expectedError    error
		expectedResponse *tfprotov6.ConfigureStateStoreResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
										SchemaMethod: func(_ context.Context, _ statestore.SchemaRequest, resp *statestore.SchemaResponse) {
											resp.Schema = schema.Schema{}
										},
										MetadataMethod: func(_ context.Context, _ statestore.MetadataRequest, resp *statestore.MetadataResponse) {
											resp.TypeName = "test_state_store"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ConfigureStateStoreRequest{
				Config:   testEmptyDynamicValue,
				TypeName: "test_state_store",
			},
			expectedResponse: &tfprotov6.ConfigureStateStoreResponse{
				Capabilities: &tfprotov6.StateStoreServerCapabilities{
					ChunkSize: 8 << 20,
				},
			},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
										SchemaMethod: func(_ context.Context, _ statestore.SchemaRequest, resp *statestore.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ statestore.MetadataRequest, resp *statestore.MetadataResponse) {
											resp.TypeName = "test_state_store"
										},
										InitializeMethod: func(ctx context.Context, req statestore.InitializeRequest, resp *statestore.InitializeResponse) {
											var config struct {
												Test types.String `tfsdk:"test"`
											}

											resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

											if config.Test.ValueString() != "test-config-value" {
												resp.Diagnostics.AddError("unexpected req.Config value: %s", config.Test.ValueString())
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ConfigureStateStoreRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_state_store",
			},
			expectedResponse: &tfprotov6.ConfigureStateStoreResponse{
				Capabilities: &tfprotov6.StateStoreServerCapabilities{
					ChunkSize: 8 << 20,
				},
			},
		},
		"capabilities": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
										SchemaMethod: func(_ context.Context, _ statestore.SchemaRequest, resp *statestore.SchemaResponse) {
											resp.Schema = schema.Schema{}
										},
										MetadataMethod: func(_ context.Context, _ statestore.MetadataRequest, resp *statestore.MetadataResponse) {
											resp.TypeName = "test_state_store"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ConfigureStateStoreRequest{
				Config:   testEmptyDynamicValue,
				TypeName: "test_state_store",
				Capabilities: &tfprotov6.ConfigureStateStoreClientCapabilities{
					ChunkSize: 4 << 20,
				},
			},
			expectedResponse: &tfprotov6.ConfigureStateStoreResponse{
				Capabilities: &tfprotov6.StateStoreServerCapabilities{
					ChunkSize: 4 << 20,
				},
			},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
										SchemaMethod: func(_ context.Context, _ statestore.SchemaRequest, resp *statestore.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ statestore.MetadataRequest, resp *statestore.MetadataResponse) {
											resp.TypeName = "test_state_store"
										},
										InitializeMethod: func(_ context.Context, _ statestore.InitializeRequest, resp *statestore.InitializeResponse) {
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
			request: &tfprotov6.ConfigureStateStoreRequest{
				Config:   testConfigDynamicValue,
				TypeName: "test_state_store",
			},
			expectedResponse: &tfprotov6.ConfigureStateStoreResponse{
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
				Capabilities: &tfprotov6.StateStoreServerCapabilities{
					ChunkSize: 8 << 20,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ConfigureStateStore(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
