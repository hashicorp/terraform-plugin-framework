// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

func TestServerLockState(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.LockStateRequest
		expectedResponse *fwserver.LockStateResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.LockStateResponse{},
		},
		"configure-success": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
				StateStoreConfigureData: fwserver.StateStoreConfigureData{
					StateStoreConfigureData: "test-statestore-configure-data",
				},
			},
			request: &fwserver.LockStateRequest{
				StateID:   "test-state-123",
				Operation: "apply",
				StateStore: &testprovider.StateStoreWithConfigure{
					StateStore: &testprovider.StateStore{
						LockMethod: func(ctx context.Context, req statestore.LockRequest, resp *statestore.LockResponse) {
							resp.LockID = "test-lock-with-configure-123"
						},
					},
					ConfigureMethod: func(ctx context.Context, req statestore.ConfigureRequest, resp *statestore.ConfigureResponse) {
						if req.StateStoreData != "test-statestore-configure-data" {
							resp.Diagnostics.AddError(
								"Unexpected req.StateStoreData value",
								fmt.Sprintf("Expected \"test-statestore-configure-data\", got: %q", req.StateStoreData),
							)
						}
					},
				},
			},
			expectedResponse: &fwserver.LockStateResponse{
				LockID: "test-lock-with-configure-123",
			},
		},
		"configure-diags": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.LockStateRequest{
				StateID:   "test-state-123",
				Operation: "apply",
				StateStore: &testprovider.StateStoreWithConfigure{
					StateStore: &testprovider.StateStore{
						LockMethod: func(ctx context.Context, req statestore.LockRequest, resp *statestore.LockResponse) {
							// This is never called
							resp.LockID = "test-lock-with-configure-123"
						},
					},
					ConfigureMethod: func(ctx context.Context, req statestore.ConfigureRequest, resp *statestore.ConfigureResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.LockStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic(
						"warning summary",
						"warning detail",
					),
					diag.NewErrorDiagnostic(
						"error summary",
						"error detail",
					),
				},
			},
		},
		"lock-success": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.LockStateRequest{
				StateID:   "test-state-123",
				Operation: "apply",
				StateStore: &testprovider.StateStore{
					LockMethod: func(ctx context.Context, req statestore.LockRequest, resp *statestore.LockResponse) {
						if req.StateID != "test-state-123" {
							resp.Diagnostics.AddError(
								"Unexpected req.StateID value",
								fmt.Sprintf("Expected \"test-state-123\", got: %q", req.StateID),
							)
						}
						if req.Operation != "apply" {
							resp.Diagnostics.AddError(
								"Unexpected req.Operation value",
								fmt.Sprintf("Expected \"apply\", got: %q", req.Operation),
							)
						}

						if resp.Diagnostics.HasError() {
							return
						}

						resp.LockID = "test-lock-123"
					},
				},
			},
			expectedResponse: &fwserver.LockStateResponse{
				LockID: "test-lock-123",
			},
		},
		"lock-diags": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.LockStateRequest{
				StateID:   "test-state-123",
				Operation: "apply",
				StateStore: &testprovider.StateStore{
					LockMethod: func(ctx context.Context, req statestore.LockRequest, resp *statestore.LockResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.LockStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic(
						"warning summary",
						"warning detail",
					),
					diag.NewErrorDiagnostic(
						"error summary",
						"error detail",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.LockStateResponse{}
			testCase.server.LockState(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
