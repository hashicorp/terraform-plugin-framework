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
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerReadResource(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testCurrentStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, nil),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testNewStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-newstate-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
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

	testConfig := &tfsdk.Config{
		Raw:    testCurrentStateValue,
		Schema: testSchema,
	}

	testCurrentState := &tfsdk.State{
		Raw:    testCurrentStateValue,
		Schema: testSchema,
	}

	testNewState := &tfsdk.State{
		Raw:    testNewStateValue,
		Schema: testSchema,
	}

	testNewStateRemoved := &tfsdk.State{
		Raw:    tftypes.NewValue(testType, nil),
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

	testPrivateFramework := &privatestate.Data{
		Framework: testPrivateFrameworkMap,
	}

	testPrivateProvider := &privatestate.Data{
		Provider: testProviderData,
	}

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testEmptyPrivate := &privatestate.Data{
		Provider: testEmptyProviderData,
	}

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.ReadResourceRequest
		expectedResponse *fwserver.ReadResourceResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ReadResourceResponse{},
		},
		"request-currentstate-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{},
			expectedResponse: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Read Request",
						"An unexpected error was encountered when reading the resource. The current state was missing.\n\n"+
							"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
					),
				},
			},
		},
		"request-currentstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var data struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-currentstate-value" {
							resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var config struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &config)...)

						if config.TestRequired.ValueString() != "test-currentstate-value" {
							resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", config.TestRequired.ValueString())
						}
					},
				},
				ProviderMeta: testConfig,
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testPrivate,
			},
		},
		"request-private-nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"resource-configure-data": {
			server: &fwserver.Server{
				Provider:              &testprovider.Provider{},
				ResourceConfigureData: "test-provider-configure-value",
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.ResourceWithConfigure{
					ConfigureMethod: func(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
					Resource: &testprovider.Resource{
						ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
							// In practice, the Configure method would save the
							// provider data to the Resource implementation and
							// use it here. The fact that Configure is able to
							// read the data proves this can work.
						},
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
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
				NewState: testCurrentState,
				Private:  testEmptyPrivate,
			},
		},
		"response-state": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						var data struct {
							TestComputed types.String `tfsdk:"test_computed"`
							TestRequired types.String `tfsdk:"test_required"`
						}

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						data.TestComputed = types.StringValue("test-newstate-value")

						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testNewState,
				Private:  testEmptyPrivate,
			},
		},
		"response-state-removeresource": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						resp.State.RemoveResource(ctx)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testNewStateRemoved,
				Private:  testEmptyPrivate,
			},
		},
		"response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testPrivateProvider,
			},
		},
		"response-private-updated": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				Resource: &testprovider.Resource{
					ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
				Private: testPrivateFramework,
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
				Private:  testPrivate,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ReadResourceResponse{}
			testCase.server.ReadResource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
