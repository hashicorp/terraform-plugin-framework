package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/emptyprovider"
)

// TODO: Migrate tfsdk.Provider bits of proto6server.testProviderServer to
// new internal/testing/provider.Provider that allows customization of all
// method implementations via struct fields. Then, create additional test
// cases in this unit test.
//
// For now this testing is covered by proto6server.ValidateProviderConfig.
//
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func TestServerValidateProviderConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.ValidateProviderConfigRequest
		expectedResponse *fwserver.ValidateProviderConfigResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &emptyprovider.Provider{},
			},
			expectedResponse: &fwserver.ValidateProviderConfigResponse{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ValidateProviderConfigResponse{}
			testCase.server.ValidateProviderConfig(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
