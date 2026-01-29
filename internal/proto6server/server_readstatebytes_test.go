// Copyright IBM Corp. 2021, 2025
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
	// TODO: Test the actual chunking after configure is established
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testCases := map[string]struct {
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
										ReadMethod: func(ctx context.Context, req statestore.ReadStateBytesRequest, resp *statestore.ReadStateResponse) {
											resp.Bytes = []byte("test-config-value")
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
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       []byte("test-config-value"),
						TotalLength: 17,
						Range: tfprotov6.StateByteRange{
							Start: 0,
							End:   17,
						},
					},
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
											resp.TypeName = "test_statestore"
										},
										ReadMethod: func(ctx context.Context, req statestore.ReadStateBytesRequest, resp *statestore.ReadStateResponse) {
											resp.Bytes = []byte("test-config-value")
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
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       []byte("test-config-value"),
						TotalLength: 17,
						Range: tfprotov6.StateByteRange{
							Start: 0,
							End:   17,
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
