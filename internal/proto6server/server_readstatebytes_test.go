// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-framework/statestore/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestServerReadStateBytes(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	var testCases = map[string]struct {
		server         *Server
		request        *tfprotov6.ReadStateBytesRequest
		expectedError  error
		expectedChunks []tfprotov6.ReadStateByteChunk
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
											resp.TypeName = "test_statestore"
										},
										ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
											resp.StateBytes = []byte(`{"version": 1, "terraform_version": "1.15.0"}`)
										},
									}
								},
							}
						},
					},
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 50,
						},
					},
				},
			},
			request: &tfprotov6.ReadStateBytesRequest{
				TypeName: "test_statestore",
				StateID:  "test_statestore",
			},
			expectedChunks: []tfprotov6.ReadStateByteChunk{
				{
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       []byte(`{"version": 1, "terraform_version": "1.15.0"}`), // total lenth is 47
						TotalLength: 45,
						Range: tfprotov6.StateByteRange{
							Start: 0,
							End:   44,
						},
					},
				},
			},
		},
		"no-typename": {
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
										},
										ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
											resp.StateBytes = []byte(`{"version": 1, "terraform_version": "1.15.0"}`)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadStateBytesRequest{
				TypeName: "",
				StateID:  "test_statestore",
			},
			expectedChunks: []tfprotov6.ReadStateByteChunk{
				{
					Diagnostics: []*tfprotov6.Diagnostic{
						{
							Severity: tfprotov6.DiagnosticSeverityError,
							Summary:  "State Store Type Missing",
							Detail:   "The *testprovider.StateStore state store returned an empty string from the Metadata method. This is always an issue with the provider and should be reported to the provider developers.",
						},
						{
							Severity: tfprotov6.DiagnosticSeverityError,
							Summary:  "State Store Type Not Found",
							Detail:   "No state store type named \"\" was found in the provider.",
						},
					},
				},
			},
		},
		"no-stateid": {
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
											resp.TypeName = "test_statestore"
										},
										ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
											resp.StateBytes = []byte(`{"version": 2, "terraform_version": "1.15.0"}`)
										},
									}
								},
							}
						},
					},
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 8 << 20, // 8 MB
						},
					},
				},
			},
			request: &tfprotov6.ReadStateBytesRequest{
				TypeName: "test_statestore",
				StateID:  "",
			},
			expectedChunks: []tfprotov6.ReadStateByteChunk{
				{
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       []byte(`{"version": 2, "terraform_version": "1.15.0"}`),
						TotalLength: 45,
						Range: tfprotov6.StateByteRange{
							Start: 0,
							End:   44,
						},
					},
				},
			},
		},
		"no-config": {
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
											resp.TypeName = "test_statestore"
										},
										ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
											resp.StateBytes = []byte(`{"version": 3, "terraform_version": "1.15.0"}`)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadStateBytesRequest{
				TypeName: "test_statestore",
				StateID:  "test_statestore",
			},
			expectedChunks: []tfprotov6.ReadStateByteChunk{
				{
					Diagnostics: []*tfprotov6.Diagnostic{
						{
							Severity: tfprotov6.DiagnosticSeverityError,
							Summary:  "Error reading state",
							Detail:   "The provider server does not have a chunk size configured. This is a bug in either Terraform core or terraform-plugin-framework and should be reported."},
					},
				},
			},
		},
		"chunking-config-chunk-size": {
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
											resp.TypeName = "test_statestore"
										},
										ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
											resp.StateBytes = []byte(`{"version": 4, "terraform_version": "1.15.0"}`)
										},
									}
								},
							}
						},
					},
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 10,
						},
					},
				},
			},
			request: &tfprotov6.ReadStateBytesRequest{
				TypeName: "test_statestore",
				StateID:  "test_statestore",
			},
			expectedChunks: []tfprotov6.ReadStateByteChunk{
				{
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       []byte(`{"version"`),
						TotalLength: 45,
						Range: tfprotov6.StateByteRange{
							Start: 0,
							End:   9,
						},
					},
				},
				{
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       []byte(`: 4, "terr`),
						TotalLength: 45,
						Range: tfprotov6.StateByteRange{
							Start: 10,
							End:   19,
						},
					},
				},
				{
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       []byte("aform_vers"),
						TotalLength: 45,
						Range: tfprotov6.StateByteRange{
							Start: 20,
							End:   29,
						},
					},
				},
				{
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       []byte(`ion": "1.1`),
						TotalLength: 45,
						Range: tfprotov6.StateByteRange{
							Start: 30,
							End:   39,
						},
					},
				},
				{
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       []byte(`5.0"}`),
						TotalLength: 45,
						Range: tfprotov6.StateByteRange{
							Start: 40,
							End:   44,
						},
					},
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ReadStateBytes(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedChunks, slices.Collect(got.Chunks)); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
