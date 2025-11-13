// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
)

func TestServerStateBytesResource(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.StateBytesRequest
		expectedStreamEvents []fwserver.StateBytesResult
		expectedError        string
	}{
		"success-with-zero-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.StateBytesRequest{
				TypeName: "test_type",
				StateId:  "test_id",
			},
			expectedStreamEvents: []fwserver.StateBytesResult{},
		},
		"success-with-nil-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.StateBytesRequest{
				TypeName: "test_type",
				StateId:  "test_id",
			},
			expectedStreamEvents: []fwserver.StateBytesResult{},
		},
		"success-with-multiple-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.StateBytesRequest{
				TypeName: "test_type",
				StateId:  "test_id",
			},
			expectedStreamEvents: []fwserver.StateBytesResult{},
		},
		"zero-results-on-empty-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.StateBytesRequest{
				TypeName: "",
				StateId:  "",
			},
			expectedStreamEvents: []fwserver.StateBytesResult{},
			expectedError:        "config cannot be nil",
		},
		"zero-results-with-warning-diagnostic": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.StateBytesRequest{
				TypeName: "test_type",
				StateId:  "test_id",
			},
			expectedStreamEvents: []fwserver.StateBytesResult{
				{
					Diagnostics: diag.Diagnostics{
						diag.NewWarningDiagnostic("Test Warning", "This is a test warning diagnostic"),
					},
				},
			},
		},
		"empty-id": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.StateBytesRequest{
				TypeName: "test_type",
				StateId:  "",
			},
			expectedStreamEvents: []fwserver.StateBytesResult{
				{
					Diagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("Incomplete StateBytes Result", "When reading statestore, an implementation issue was found. This is always a problem with the provider. Please report this to the provider developers."),
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.StateBytesStream{}
			testCase.server.StateBytesResource(context.Background(), testCase.request, response)

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

			events := slices.AppendSeq([]fwserver.StateBytesResult{}, response.Chunks)
			if diff := cmp.Diff(events, testCase.expectedStreamEvents, opts); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
