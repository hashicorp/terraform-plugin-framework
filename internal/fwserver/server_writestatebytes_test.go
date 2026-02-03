// Copyright IBM Corp. 2021, 2025
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

func TestServerWriteStateBytes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.WriteStateBytesRequest
		expectedResponse *fwserver.WriteStateBytesResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.WriteStateBytesResponse{},
		},
		"request": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.WriteStateBytesRequest{
				StateID:    "test-state-123",
				StateBytes: []byte(`{"version": 4, "terraform_version": "1.15.0"}`),
				StateStore: &testprovider.StateStore{
					WriteMethod: func(ctx context.Context, req statestore.WriteRequest, resp *statestore.WriteResponse) {
						if req.StateID != "test-state-123" {
							resp.Diagnostics.AddError(
								"Unexpected req.StateID",
								fmt.Sprintf("expected \"test-state-123\", got: %q", req.StateID),
							)
							return
						}

						if string(req.StateBytes) != `{"version": 4, "terraform_version": "1.15.0"}` {
							resp.Diagnostics.AddError(
								"Unexpected req.StateBytes",
								fmt.Sprintf("expected \"test-state-123\", got: %q", string(req.StateBytes)),
							)
							return
						}
					},
				},
			},
			expectedResponse: &fwserver.WriteStateBytesResponse{},
		},
		"statestore-configure-data": {
			server: &fwserver.Server{
				StateStoreConfigureData: fwserver.StateStoreConfigureData{
					StateStoreConfigureData: "test-statestore-configure-value",
				},
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.WriteStateBytesRequest{
				StateStore: &testprovider.StateStoreWithConfigure{
					StateStore: &testprovider.StateStore{},
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
			},
			expectedResponse: &fwserver.WriteStateBytesResponse{},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.WriteStateBytesRequest{
				StateStore: &testprovider.StateStore{
					WriteMethod: func(ctx context.Context, req statestore.WriteRequest, resp *statestore.WriteResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.WriteStateBytesResponse{
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

			response := &fwserver.WriteStateBytesResponse{}
			testCase.server.WriteStateBytes(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
