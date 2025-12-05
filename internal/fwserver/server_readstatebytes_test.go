// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
)

func TestServerReadStateBytesResource(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.ReadStateBytesRequest
		expectedStreamEvents []fwserver.ReadStateBytesResponse
		expectedError        string
	}{
		"success-with-zero-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateId: "test_id",
			},
			expectedStreamEvents: []fwserver.ReadStateBytesResponse{},
		},
		"success-with-nil-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateId: "test_id",
			},
			expectedStreamEvents: []fwserver.ReadStateBytesResponse{},
		},
		"success-with-multiple-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateId: "test_id",
			},
			expectedStreamEvents: []fwserver.ReadStateBytesResponse{},
		},
		"zero-results-on-empty-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateId: "",
			},
			expectedStreamEvents: []fwserver.ReadStateBytesResponse{},
			expectedError:        "config cannot be nil",
		},
		"zero-results-with-warning-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateId: "test_id",
			},
			expectedStreamEvents: []fwserver.ReadStateBytesResponse{
				{},
			},
		},
		"empty-id": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadStateBytesRequest{
				StateId: "",
			},
			expectedStreamEvents: []fwserver.ReadStateBytesResponse{
				{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			//response := &fwserver.ReadStateBytesResponse{}
			//testCase.server.ReadStateBytesResource(context.Background(), testCase.request, response)

			opts := cmp.Options{
				cmp.Comparer(func(a, b diag.Diagnostics) bool {
					for i := range a {
						if a[i].Severity() != b[i].Severity() || a[i].Summary() != b[i].Summary() {
							return false
						}
					}
					return true
				}),
			}

			events := slices.AppendSeq([]fwserver.ReadStateBytesResponse{}, nil)
			if diff := cmp.Diff(events, testCase.expectedStreamEvents, opts); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
