// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestServerRenewEphemeralResource(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testPrivateFrameworkMap := map[string][]byte{
		".frameworkKey": []byte(`{"fk": "framework value"}`),
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testPrivate := &privatestate.Data{
		Framework: testPrivateFrameworkMap,
		Provider:  testProviderData,
	}

	testPrivateProvider := &privatestate.Data{
		Provider: testProviderData,
	}

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testEmptyPrivate := &privatestate.Data{
		Provider: testEmptyProviderData,
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.RenewEphemeralResourceRequest
		expectedResponse     *fwserver.RenewEphemeralResourceResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.RenewEphemeralResourceResponse{},
		},
		"request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.RenewEphemeralResourceRequest{
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithRenew{
					RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
						expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

						key := "providerKeyOne"
						got, diags := req.Private.GetKey(ctx, key)

						resp.Diagnostics.Append(diags...)

						if string(got) != expected {
							resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
						}
					},
				},
				Private: testPrivate,
			},
			expectedResponse: &fwserver.RenewEphemeralResourceResponse{
				Private: testPrivate,
			},
		},
		"request-private-nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.RenewEphemeralResourceRequest{
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithRenew{
					RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
						var expected []byte

						key := "providerKeyOne"
						got, diags := req.Private.GetKey(ctx, key)

						resp.Diagnostics.Append(diags...)

						if !bytes.Equal(got, expected) {
							resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
						}
					},
				},
			},
			expectedResponse: &fwserver.RenewEphemeralResourceResponse{
				Private: testEmptyPrivate,
			},
		},
		"ephemeralresource-no-renew-implementation-diagnostic": {
			server: &fwserver.Server{
				EphemeralResourceConfigureData: "test-provider-configure-value",
				Provider:                       &testprovider.Provider{},
			},
			request: &fwserver.RenewEphemeralResourceRequest{
				EphemeralResourceSchema: testSchema,
				// Doesn't implement Renew interface
				EphemeralResource: &testprovider.EphemeralResource{},
			},
			expectedResponse: &fwserver.RenewEphemeralResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Ephemeral Resource Renew Not Implemented",
						"An unexpected error was encountered when renewing the ephemeral resource. Terraform sent a renewal request for an "+
							"ephemeral resource that has not implemented renewal logic.\n\n"+
							"Please report this to the provider developer.",
					),
				},
			},
		},
		"ephemeralresource-configure-data": {
			server: &fwserver.Server{
				EphemeralResourceConfigureData: "test-provider-configure-value",
				Provider:                       &testprovider.Provider{},
			},
			request: &fwserver.RenewEphemeralResourceRequest{
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithConfigureAndRenew{
					ConfigureMethod: func(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
						providerData, ok := req.ProviderData.(string)

						if !ok {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.ProviderData",
								fmt.Sprintf("Expected string, got: %T", req.ProviderData),
							)
							return
						}

						if providerData != "test-provider-configure-value" {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.ProviderData",
								fmt.Sprintf("Expected test-provider-configure-value, got: %q", providerData),
							)
						}
					},
					RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
						// In practice, the Configure method would save the
						// provider data to the EphemeralResource implementation and
						// use it here. The fact that Configure is able to
						// read the data proves this can work.
					},
				},
			},
			expectedResponse: &fwserver.RenewEphemeralResourceResponse{
				Private: testEmptyPrivate,
			},
		},
		"response-default-values": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.RenewEphemeralResourceRequest{
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithRenew{
					RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {},
				},
			},
			expectedResponse: &fwserver.RenewEphemeralResourceResponse{
				Private: testEmptyPrivate,
				RenewAt: *new(time.Time),
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.RenewEphemeralResourceRequest{
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithRenew{
					RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.RenewEphemeralResourceResponse{
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
				Private: testEmptyPrivate,
			},
		},
		"response-renew-at": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.RenewEphemeralResourceRequest{
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithRenew{
					RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
						resp.RenewAt = time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC)
					},
				},
			},
			expectedResponse: &fwserver.RenewEphemeralResourceResponse{
				Private: testEmptyPrivate,
				RenewAt: time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC),
			},
		},
		"response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.RenewEphemeralResourceRequest{
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithRenew{
					RenewMethod: func(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.RenewEphemeralResourceResponse{
				Private: testPrivateProvider,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.configureProviderReq != nil {
				configureProviderResp := &provider.ConfigureResponse{}
				testCase.server.ConfigureProvider(context.Background(), testCase.configureProviderReq, configureProviderResp)
			}

			response := &fwserver.RenewEphemeralResourceResponse{}
			testCase.server.RenewEphemeralResource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
