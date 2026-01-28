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

func TestServerUnlockState(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.UnlockStateRequest
		expectedResponse *fwserver.UnlockStateResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.UnlockStateResponse{},
		},
		"configure-success": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
				StateStoreConfigureData: fwserver.StateStoreConfigureData{
					StateStoreConfigureData: "test-statestore-configure-data",
				},
			},
			request: &fwserver.UnlockStateRequest{
				StateID: "test-state-123",
				LockID:  "test-lock-123",
				StateStore: &testprovider.StateStoreWithConfigure{
					StateStore: &testprovider.StateStore{
						UnlockMethod: func(ctx context.Context, req statestore.UnlockRequest, resp *statestore.UnlockResponse) {},
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
			expectedResponse: &fwserver.UnlockStateResponse{},
		},
		"configure-diags": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UnlockStateRequest{
				StateID: "test-state-123",
				LockID:  "test-lock-123",
				StateStore: &testprovider.StateStoreWithConfigure{
					StateStore: &testprovider.StateStore{
						UnlockMethod: func(ctx context.Context, req statestore.UnlockRequest, resp *statestore.UnlockResponse) {},
					},
					ConfigureMethod: func(ctx context.Context, req statestore.ConfigureRequest, resp *statestore.ConfigureResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.UnlockStateResponse{
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
		"unlock-success": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UnlockStateRequest{
				StateID: "test-state-123",
				LockID:  "test-lock-123",
				StateStore: &testprovider.StateStore{
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
				},
			},
			expectedResponse: &fwserver.UnlockStateResponse{},
		},
		"unlock-diags": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.UnlockStateRequest{
				StateID: "test-state-123",
				LockID:  "test-lock-123",
				StateStore: &testprovider.StateStore{
					UnlockMethod: func(ctx context.Context, req statestore.UnlockRequest, resp *statestore.UnlockResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.UnlockStateResponse{
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

			response := &fwserver.UnlockStateResponse{}
			testCase.server.UnlockState(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
