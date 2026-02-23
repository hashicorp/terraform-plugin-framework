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
	"github.com/hashicorp/terraform-plugin-framework/statestore/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerConfigureStateStore(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test": tftypes.String,
		},
	}

	testValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test": schema.StringAttribute{
				Required: true,
			},
		},
	}

	type testSchemaData struct {
		Test types.String `tfsdk:"test"`
	}

	testConfig := tfsdk.Config{
		Raw:    testValue,
		Schema: testSchema,
	}

	testCases := map[string]struct {
		server                          *fwserver.Server
		request                         *fwserver.ConfigureStateStoreRequest
		expectedResponse                *fwserver.ConfigureStateStoreResponse
		expectedStateStoreConfigureData fwserver.StateStoreConfigureData
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ConfigureStateStoreResponse{},
		},
		"request-nil-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ConfigureStateStoreRequest{
				StateStoreSchema: testSchema,
				StateStore: &testprovider.StateStore{
					InitializeMethod: func(ctx context.Context, req statestore.InitializeRequest, resp *statestore.InitializeResponse) {
						if !req.Config.Raw.IsNull() {
							resp.Diagnostics.AddError("Unexpected req.Config Value, expected <null>", fmt.Sprintf("Got: %s", req.Config.Raw.String()))
						}
					},
				},
			},
			expectedResponse: &fwserver.ConfigureStateStoreResponse{
				ServerCapabilities: &fwserver.StateStoreServerCapabilities{},
			},
		},
		"request-provider-data": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ConfigureStateStoreRequest{
				StateStoreSchema: testSchema,
				StateStore: &testprovider.StateStore{
					InitializeMethod: func(ctx context.Context, req statestore.InitializeRequest, resp *statestore.InitializeResponse) {
						if req.ProviderData != "provider data for state stores" {
							resp.Diagnostics.AddError("Unexpected req.ProviderData Value", fmt.Sprintf("Got: %s", req.ProviderData))
						}
					},
				},
			},
			expectedResponse: &fwserver.ConfigureStateStoreResponse{
				ServerCapabilities: &fwserver.StateStoreServerCapabilities{},
			},
		},
		"request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ConfigureStateStoreRequest{
				Config:           &testConfig,
				StateStoreSchema: testSchema,
				StateStore: &testprovider.StateStore{
					InitializeMethod: func(ctx context.Context, req statestore.InitializeRequest, resp *statestore.InitializeResponse) {
						var data testSchemaData
						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						if data.Test.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.Test.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.ConfigureStateStoreResponse{
				ServerCapabilities: &fwserver.StateStoreServerCapabilities{},
			},
		},
		"request-state-store-server-capabilities": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ConfigureStateStoreRequest{
				StateStoreSchema: testSchema,
				StateStore:       &testprovider.StateStore{},
				ClientCapabilities: fwserver.ConfigureStateStoreClientCapabilities{
					ChunkSize: 4 << 20,
				},
			},
			expectedResponse: &fwserver.ConfigureStateStoreResponse{
				ServerCapabilities: &fwserver.StateStoreServerCapabilities{
					ChunkSize: 4 << 20,
				},
			},
			expectedStateStoreConfigureData: fwserver.StateStoreConfigureData{
				ServerCapabilities: fwserver.StateStoreServerCapabilities{
					ChunkSize: 4 << 20,
				},
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ConfigureStateStoreRequest{
				Config:           &testConfig,
				StateStoreSchema: testSchema,
				StateStore: &testprovider.StateStore{
					InitializeMethod: func(ctx context.Context, req statestore.InitializeRequest, resp *statestore.InitializeResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.ConfigureStateStoreResponse{
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
				ServerCapabilities: &fwserver.StateStoreServerCapabilities{},
			},
		},
		"response-statestoredata": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ConfigureStateStoreRequest{
				Config:           &testConfig,
				StateStoreSchema: testSchema,
				StateStore: &testprovider.StateStore{
					InitializeMethod: func(ctx context.Context, req statestore.InitializeRequest, resp *statestore.InitializeResponse) {
						resp.StateStoreData = req.ProviderData
					},
				},
			},
			expectedResponse: &fwserver.ConfigureStateStoreResponse{
				ServerCapabilities: &fwserver.StateStoreServerCapabilities{},
			},
			expectedStateStoreConfigureData: fwserver.StateStoreConfigureData{
				StateStoreConfigureData: "provider data for state stores",
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testCase.server.StateStoreProviderData = "provider data for state stores"
			response := &fwserver.ConfigureStateStoreResponse{}
			testCase.server.ConfigureStateStore(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.server.StateStoreConfigureData, testCase.expectedStateStoreConfigureData); diff != "" {
				t.Errorf("unexpected difference in StateStoreConfigureData: %s", diff)
			}
		})
	}
}
