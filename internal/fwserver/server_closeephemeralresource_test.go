// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerCloseEphemeralResource(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

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

	testState := &tfsdk.EphemeralState{
		Raw:    testStateValue,
		Schema: testSchema,
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

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.CloseEphemeralResourceRequest
		expectedResponse     *fwserver.CloseEphemeralResourceResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.CloseEphemeralResourceResponse{},
		},
		"request-state-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CloseEphemeralResourceRequest{
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithClose{
					CloseMethod: func(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {},
				},
			},
			expectedResponse: &fwserver.CloseEphemeralResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Close Request",
						"An unexpected error was encountered when closing the ephemeral resource. The state was missing.\n\n"+
							"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
					),
				},
			},
		},
		"request-state": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CloseEphemeralResourceRequest{
				State:                   testState,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithClose{
					CloseMethod: func(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
						var data struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestRequired.ValueString())
						}

						if data.TestComputed.ValueString() != "test-state-value" {
							resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestComputed.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.CloseEphemeralResourceResponse{},
		},
		"request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CloseEphemeralResourceRequest{
				State:                   testState,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithClose{
					CloseMethod: func(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
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
			expectedResponse: &fwserver.CloseEphemeralResourceResponse{},
		},
		"request-private-nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CloseEphemeralResourceRequest{
				State:                   testState,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithClose{
					CloseMethod: func(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
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
			expectedResponse: &fwserver.CloseEphemeralResourceResponse{},
		},
		"ephemeralresource-no-close-implementation-diagnostic": {
			server: &fwserver.Server{
				EphemeralResourceConfigureData: "test-provider-configure-value",
				Provider:                       &testprovider.Provider{},
			},
			request: &fwserver.CloseEphemeralResourceRequest{
				State:                   testState,
				EphemeralResourceSchema: testSchema,
				// Doesn't implement Close interface
				EphemeralResource: &testprovider.EphemeralResource{},
			},
			expectedResponse: &fwserver.CloseEphemeralResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Ephemeral Resource Close Not Implemented",
						"This ephemeral resource does not support close. Please contact the provider developer for additional information.",
					),
				},
			},
		},
		"ephemeralresource-configure-data": {
			server: &fwserver.Server{
				EphemeralResourceConfigureData: "test-provider-configure-value",
				Provider:                       &testprovider.Provider{},
			},
			request: &fwserver.CloseEphemeralResourceRequest{
				State:                   testState,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithConfigureAndClose{
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
					CloseMethod: func(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
						// In practice, the Configure method would save the
						// provider data to the EphemeralResource implementation and
						// use it here. The fact that Configure is able to
						// read the data proves this can work.
					},
				},
			},
			expectedResponse: &fwserver.CloseEphemeralResourceResponse{},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.CloseEphemeralResourceRequest{
				State:                   testState,
				EphemeralResourceSchema: testSchema,
				EphemeralResource: &testprovider.EphemeralResourceWithClose{
					CloseMethod: func(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.CloseEphemeralResourceResponse{
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
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.configureProviderReq != nil {
				configureProviderResp := &provider.ConfigureResponse{}
				testCase.server.ConfigureProvider(context.Background(), testCase.configureProviderReq, configureProviderResp)
			}

			response := &fwserver.CloseEphemeralResourceResponse{}
			testCase.server.CloseEphemeralResource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
