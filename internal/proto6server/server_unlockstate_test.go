// Copyright IBM Corp. 2021, 2026
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

func TestServerUnlockState(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.UnlockStateRequest
		expectedError    error
		expectedResponse *tfprotov6.UnlockStateResponse
	}{
		"unlockstate-success": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						StateStoresMethod: func(ctx context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
										MetadataMethod: func(_ context.Context, _ statestore.MetadataRequest, resp *statestore.MetadataResponse) {
											resp.TypeName = "test_statestore"
										},
										UnlockMethod: func(ctx context.Context, req statestore.UnlockRequest, resp *statestore.UnlockResponse) {
											if req.StateID != "test-state-123" {
												resp.Diagnostics.AddError(
													"Unexpected req.StateID value",
													fmt.Sprintf("Expected \"test-state-123\", got: %q", req.StateID),
												)
											}
											if req.LockID != "test-lock-123" {
												resp.Diagnostics.AddError(
													"Unexpected req.LockID value",
													fmt.Sprintf("Expected \"test-lock-123\", got: %q", req.LockID),
												)
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.UnlockStateRequest{
				TypeName: "test_statestore",
				StateID:  "test-state-123",
				LockID:   "test-lock-123",
			},
			expectedResponse: &tfprotov6.UnlockStateResponse{},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						StateStoresMethod: func(ctx context.Context) []func() statestore.StateStore {
							return []func() statestore.StateStore{
								func() statestore.StateStore {
									return &testprovider.StateStore{
										MetadataMethod: func(_ context.Context, _ statestore.MetadataRequest, resp *statestore.MetadataResponse) {
											resp.TypeName = "test_statestore"
										},
										UnlockMethod: func(ctx context.Context, req statestore.UnlockRequest, resp *statestore.UnlockResponse) {
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
			request: &tfprotov6.UnlockStateRequest{
				TypeName: "test_statestore",
			},
			expectedResponse: &tfprotov6.UnlockStateResponse{
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

			got, err := testCase.server.UnlockState(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
