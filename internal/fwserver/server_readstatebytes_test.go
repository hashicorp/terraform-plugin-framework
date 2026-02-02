// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
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
					ReadMethod: func(ctx context.Context, req statestore.ReadStateBytesRequest, resp *statestore.ReadStateBytesResponse) {
						resp.StateBytes = []byte{}
					},
				},
				StateID: "test_id",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				Bytes: []byte{},
			},
		},
		"success-with-nil-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStore{
					ReadMethod: func(ctx context.Context, req statestore.ReadStateBytesRequest, resp *statestore.ReadStateBytesResponse) {
						resp.StateBytes = nil
					},
				},
				StateID: "test_id",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				Bytes: nil,
			},
		},
		"success-with-data": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStore{
					ReadMethod: func(ctx context.Context, req statestore.ReadStateBytesRequest, resp *statestore.ReadStateBytesResponse) {
						resp.StateBytes = []byte("test-data")
					},
				},
				StateID: "test_id",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				Bytes: []byte("test-data"),
			},
		},
		"empty-state-id": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStore{
					ReadMethod: func(ctx context.Context, req statestore.ReadStateBytesRequest, resp *statestore.ReadStateBytesResponse) {
						resp.StateBytes = []byte{}
					},
				},
				StateID: "",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				Bytes: []byte{},
			},
		},
		"with-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateStore: &testprovider.StateStore{
					ReadMethod: func(ctx context.Context, req statestore.ReadStateBytesRequest, resp *statestore.ReadStateBytesResponse) {
						resp.StateBytes = []byte("test-data")
						resp.Diagnostics.AddWarning("Test Warning", "This is a test warning")
					},
				},
				StateID: "test_id",
			},
			expectedResponse: &fwserver.ReadStateBytesResponse{
				Bytes: []byte("test-data"),
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
