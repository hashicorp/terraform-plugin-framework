// Copyright (c) HashiCorp, Inc.
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
)

func TestServerGetStates(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.GetStatesRequest
		expectedResponse *fwserver.GetStatesResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.GetStatesResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"StateStore Type Not Found",
						`No statestore type named "" was found in the provider.`,
					),
				},
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.GetStatesResponse{}
			testCase.server.GetStates(context.Background(), testCase.request, response)

			opts := cmp.Options{
				cmpopts.EquateEmpty(),
			}

			if diff := cmp.Diff(response, testCase.expectedResponse, opts...); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
