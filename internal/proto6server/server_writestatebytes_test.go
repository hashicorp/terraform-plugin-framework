// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

type streamedChunk struct {
	Chunk       *tfprotov6.WriteStateBytesChunk
	Diagnostics []*tfprotov6.Diagnostic
}

func TestServerWriteStateBytes(t *testing.T) {
	t.Parallel()

	defaultConfigureData := fwserver.StateStoreConfigureData{
		ServerCapabilities: fwserver.StateStoreServerCapabilities{
			ChunkSize: 8 << 20, // 8MB is the core default
		},
	}

	testCases := map[string]struct {
		server           *Server
		streamedChunks   []streamedChunk
		expectedError    error
		expectedResponse *tfprotov6.WriteStateBytesResponse
	}{
		"state-bytes-single-chunk": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 45,
						},
					},
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
										MetadataMethod: func(_ context.Context, _ statestore.MetadataRequest, resp *statestore.MetadataResponse) {
											resp.TypeName = "test_state_store"
										},
										WriteMethod: func(ctx context.Context, req statestore.WriteRequest, resp *statestore.WriteResponse) {
											expectedStateBytes := `{"version": 4, "terraform_version": "1.15.0"}`
											if string(req.StateBytes) != `{"version": 4, "terraform_version": "1.15.0"}` {
												resp.Diagnostics.AddError(
													"Unexpected req.StateBytes",
													fmt.Sprintf("expected %q, got: %q", expectedStateBytes, string(req.StateBytes)),
												)
												return
											}
										},
									}
								},
							}
						},
					},
				},
			},
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							TypeName: "test_state_store",
							StateID:  "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`{"version": 4, "terraform_version": "1.15.0"}`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   45,
							},
						},
					},
				},
			},
			expectedResponse: &tfprotov6.WriteStateBytesResponse{},
		},
		"state-bytes-multiple-chunks": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 10,
						},
					},
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
										MetadataMethod: func(_ context.Context, _ statestore.MetadataRequest, resp *statestore.MetadataResponse) {
											resp.TypeName = "test_state_store"
										},
										WriteMethod: func(ctx context.Context, req statestore.WriteRequest, resp *statestore.WriteResponse) {
											expectedStateBytes := `{"version": 4, "terraform_version": "1.15.0"}`
											if string(req.StateBytes) != `{"version": 4, "terraform_version": "1.15.0"}` {
												resp.Diagnostics.AddError(
													"Unexpected req.StateBytes",
													fmt.Sprintf("expected %q, got: %q", expectedStateBytes, string(req.StateBytes)),
												)
												return
											}
										},
									}
								},
							}
						},
					},
				},
			},
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							TypeName: "test_state_store",
							StateID:  "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`{"version"`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   9,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`: 4, "terr`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 10,
								End:   19,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`aform_vers`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 20,
								End:   29,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`ion": "1.1`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 30,
								End:   39,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
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
			expectedResponse: &tfprotov6.WriteStateBytesResponse{},
		},
		"diags-chunk-size-not-configured": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					// Omitting this simulates Terraform not calling ConfigureStateStore RPC first
					// StateStoreConfigureData: defaultConfigureData,
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{}
								},
							}
						},
					},
				},
			},
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							TypeName: "test_state_store",
							StateID:  "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`{"version": 4, "terraform_version": "1.15.0"}`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   45,
							},
						},
					},
				},
			},
			expectedResponse: &tfprotov6.WriteStateBytesResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Error Writing State",
						Detail:   "The provider server does not have a chunk size configured. This is a bug in either Terraform or terraform-plugin-framework and should be reported.",
					},
				},
			},
		},
		"diags-grpc-error": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: defaultConfigureData,
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{}
								},
							}
						},
					},
				},
			},
			streamedChunks: []streamedChunk{
				{
					Diagnostics: []*tfprotov6.Diagnostic{
						{
							Severity: tfprotov6.DiagnosticSeverityError,
							Summary:  "Fake GRPC error",
							Detail:   "Something went wrong at the GRPC layer!",
						},
					},
				},
			},
			expectedResponse: &tfprotov6.WriteStateBytesResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Fake GRPC error",
						Detail:   "Something went wrong at the GRPC layer!",
					},
				},
			},
		},
		"diags-invalid-middle-chunk": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 10,
						},
					},
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{}
								},
							}
						},
					},
				},
			},
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							TypeName: "test_state_store",
							StateID:  "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`{"version"`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   9,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`CHUNKTOOLARGE`), // This chunk is larger than the configured size (10 bytes)
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 10,
								End:   19,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`aform_vers`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 20,
								End:   29,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`ion": "1.1`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 30,
								End:   39,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
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
			expectedResponse: &tfprotov6.WriteStateBytesResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Error Writing State",
						Detail:   "Unexpected chunk of size 13 was received from Terraform, expected chunk size was 10. This is a bug and should be reported.",
					},
				},
			},
		},
		"diags-invalid-last-chunk": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 10,
						},
					},
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{}
								},
							}
						},
					},
				},
			},
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							TypeName: "test_state_store",
							StateID:  "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`{"version"`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   9,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`: 4, "terr`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 10,
								End:   19,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`aform_vers`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 20,
								End:   29,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`ion": "1.1`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 30,
								End:   39,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`CHUNKTOOLARGE`), // This chunk is larger than the configured size (10 bytes)
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 40,
								End:   44,
							},
						},
					},
				},
			},
			expectedResponse: &tfprotov6.WriteStateBytesResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Error Writing State",
						Detail:   "Unexpected final chunk of size 13 was received from Terraform, which exceeds the configured chunk size of 10. This is a bug and should be reported.",
					},
				},
			},
		},
		"diags-empty-state": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 10,
						},
					},
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{}
								},
							}
						},
					},
				},
			},
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							TypeName: "test_state_store",
							StateID:  "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(""),
							TotalLength: 0,
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   0,
							},
						},
					},
				},
			},
			expectedResponse: &tfprotov6.WriteStateBytesResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Error Writing State",
						Detail:   "No state data was received from Terraform. This is a bug and should be reported.",
					},
				},
			},
		},
		"diags-invalid-total-length": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 10,
						},
					},
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{}
								},
							}
						},
					},
				},
			},
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							TypeName: "test_state_store",
							StateID:  "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`{"version"`),
							TotalLength: 50, // total length is actually 45
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   9,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`: 4, "terr`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 10,
								End:   19,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`aform_vers`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 20,
								End:   29,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`ion": "1.1`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 30,
								End:   39,
							},
						},
					},
				},
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
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
			expectedResponse: &tfprotov6.WriteStateBytesResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Error Writing State",
						Detail:   "Unexpected size of state data received from Terraform, got: 45, expected: 50. This is a bug and should be reported.",
					},
				},
			},
		},
		"diags-empty-state-id": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 45,
						},
					},
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
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
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							TypeName: "test_state_store",
							// Omitting this field
							// StateID:  "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`{"version": 4, "terraform_version": "1.15.0"}`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   45,
							},
						},
					},
				},
			},
			expectedResponse: &tfprotov6.WriteStateBytesResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Error Writing State",
						Detail:   "No state ID was received from Terraform. This is a bug and should be reported.",
					},
				},
			},
		},
		"diags-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: fwserver.StateStoreConfigureData{
						ServerCapabilities: fwserver.StateStoreServerCapabilities{
							ChunkSize: 45,
						},
					},
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
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
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							// Omitting this field
							// TypeName: "test_state_store",
							StateID: "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`{"version": 4, "terraform_version": "1.15.0"}`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   45,
							},
						},
					},
				},
			},
			expectedResponse: &tfprotov6.WriteStateBytesResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "State Store Type Not Found",
						Detail:   "No state store type named \"\" was found in the provider.",
					},
				},
			},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					StateStoreConfigureData: defaultConfigureData,
					Provider: &testprovider.Provider{
						StateStoresMethod: func(_ context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
										MetadataMethod: func(_ context.Context, _ statestore.MetadataRequest, resp *statestore.MetadataResponse) {
											resp.TypeName = "test_state_store"
										},
										WriteMethod: func(ctx context.Context, req statestore.WriteRequest, resp *statestore.WriteResponse) {
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
			streamedChunks: []streamedChunk{
				{
					Chunk: &tfprotov6.WriteStateBytesChunk{
						Meta: &tfprotov6.WriteStateChunkMeta{
							TypeName: "test_state_store",
							StateID:  "test-state-123",
						},
						StateByteChunk: tfprotov6.StateByteChunk{
							Bytes:       []byte(`{"version": 4, "terraform_version": "1.15.0"}`),
							TotalLength: 45,
							Range: tfprotov6.StateByteRange{
								Start: 0,
								End:   45,
							},
						},
					},
				},
			},
			expectedResponse: &tfprotov6.WriteStateBytesResponse{
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
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Ensure the order of the streamed chunks
			requestStream := &tfprotov6.WriteStateBytesStream{
				Chunks: func(yield func(*tfprotov6.WriteStateBytesChunk, []*tfprotov6.Diagnostic) bool) {
					for _, streamedChunk := range testCase.streamedChunks {
						if !yield(streamedChunk.Chunk, streamedChunk.Diagnostics) {
							return
						}
					}
				},
			}

			got, err := testCase.server.WriteStateBytes(context.Background(), requestStream)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
