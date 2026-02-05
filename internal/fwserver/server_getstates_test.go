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

func TestServerGetStates(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.GetStatesRequest
		expectedResponse *fwserver.GetStatesResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.GetStatesResponse{},
		},
		"configure-success": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
				StateStoreConfigureData: fwserver.StateStoreConfigureData{
					StateStoreConfigureData: "test-statestore-configure-data",
				},
			},
			request: &fwserver.GetStatesRequest{
				StateStore: &testprovider.StateStoreWithConfigure{
					StateStore: &testprovider.StateStore{
						GetStatesMethod: func(ctx context.Context, req statestore.GetStatesRequest, resp *statestore.GetStatesResponse) {},
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
			expectedResponse: &fwserver.GetStatesResponse{},
		},
		"configure-response-diags": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GetStatesRequest{
				StateStore: &testprovider.StateStoreWithConfigure{
					StateStore: &testprovider.StateStore{
						GetStatesMethod: func(ctx context.Context, req statestore.GetStatesRequest, resp *statestore.GetStatesResponse) {},
					},
					ConfigureMethod: func(ctx context.Context, req statestore.ConfigureRequest, resp *statestore.ConfigureResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.GetStatesResponse{
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
		"response-stateids": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GetStatesRequest{
				StateStore: &testprovider.StateStore{
					GetStatesMethod: func(ctx context.Context, req statestore.GetStatesRequest, resp *statestore.GetStatesResponse) {
						resp.StateIDs = []string{"hello", "world"}
					},
				},
			},
			expectedResponse: &fwserver.GetStatesResponse{
				StateIDs: []string{"hello", "world"},
			},
		},
		"response-diags": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GetStatesRequest{
				StateStore: &testprovider.StateStore{
					GetStatesMethod: func(ctx context.Context, req statestore.GetStatesRequest, resp *statestore.GetStatesResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.GetStatesResponse{
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

			response := &fwserver.GetStatesResponse{}
			testCase.server.GetStates(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
