// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

func TestServerReadStateBytesResource(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.ReadStateBytesRequest
		expectedResponse *fwserver.ReadStateBytesResponse
	}{
		"success-with-zero-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStore{
					ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
						resp.StateBytes = []byte{}
					},
				},
				StateID: "test_id",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				StateBytes: []byte{},
			},
		},
		"success-with-nil-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStore{
					ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
						resp.StateBytes = nil
					},
				},
				StateID: "test_id",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				StateBytes: nil,
			},
		},
		"success-with-data": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStore{
					ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
						resp.StateBytes = []byte("test-data")
					},
				},
				StateID: "test_id",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				StateBytes: []byte("test-data"),
			},
		},
		"success-with-configure": {
			server: &fwserver.Server{
				StateStoreConfigureData: fwserver.StateStoreConfigureData{
					StateStoreConfigureData: "test-statestore-configure-value",
				},
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStoreWithConfigure{
					StateStore: &testprovider.StateStore{
						ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
							resp.StateBytes = []byte("test-data")
						},
					},
					ConfigureMethod: func(ctx context.Context, req statestore.ConfigureRequest, resp *statestore.ConfigureResponse) {
						stateStoreData, ok := req.StateStoreData.(string)

						if !ok {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.StateStoreData",
								fmt.Sprintf("Expected string, got: %T", req.StateStoreData),
							)
							return
						}

						if stateStoreData != "test-statestore-configure-value" {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.StateStoreData",
								fmt.Sprintf("Expected test-statestore-configure-value, got: %q", stateStoreData),
							)
						}
					},
				},
				StateID: "test_id",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				StateBytes: []byte("test-data"),
			},
		},
		"empty-state-id": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStore{
					ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
						resp.StateBytes = []byte{}
					},
				},
				StateID: "",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				StateBytes: []byte{},
			},
		},
		"with-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStore{
					ReadMethod: func(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
						resp.StateBytes = []byte("test-data")
						resp.Diagnostics.AddWarning("Test Warning", "This is a test warning")
					},
				},
				StateID: "test_id",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				StateBytes: []byte("test-data"),
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("Test Warning", "This is a test warning"),
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ReadStateBytesResponse{}
			testCase.server.ReadStateBytes(context.Background(), testCase.request, response)

			opts := cmp.Options{
				cmpopts.EquateEmpty(),
			}

			if diff := cmp.Diff(response, testCase.expectedResponse, opts...); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
