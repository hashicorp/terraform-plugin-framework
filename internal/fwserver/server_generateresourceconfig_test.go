// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerGenerateResourceConfig(t *testing.T) {
	t.Parallel()

	testNestedBlockType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_nested_block_attr": tftypes.String,
		},
	}

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"test_computed":     tftypes.String,
			"test_optional":     tftypes.String,
			"test_required":     tftypes.String,
			"test_deprecated":   tftypes.List{ElementType: tftypes.String},
			"test_false_bool":   tftypes.Bool,
			"test_empty_string": tftypes.String,
			"test_deprecated_block": tftypes.List{
				ElementType: testNestedBlockType,
			},
		},
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_optional": schema.StringAttribute{
				Optional: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
			"test_deprecated": schema.ListAttribute{
				ElementType:        types.StringType,
				Optional:           true,
				DeprecationMessage: "deprecated",
			},
			"test_false_bool": schema.BoolAttribute{
				Optional: true,
			},
			"test_empty_string": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"test_deprecated_block": schema.ListNestedBlock{
				DeprecationMessage: "deprecated",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"test_nested_block_attr": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.GenerateResourceConfigRequest
		expectedResponse     *fwserver.GenerateResourceConfigResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{},
		},
		"request-state-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Generate Config Request",
						"An unexpected error was encountered when generating resource configuration. The current state was missing.\n\n"+
							"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
					),
				},
			},
		},
		"response-default-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: testSchema,
				State: &tfsdk.State{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"id":            tftypes.NewValue(tftypes.String, "test-id-val"),
						"test_computed": tftypes.NewValue(tftypes.String, "test-computed-val"),
						"test_optional": tftypes.NewValue(tftypes.String, "test-optional-val"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_deprecated": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-deprecated-a"),
							tftypes.NewValue(tftypes.String, "test-deprecated-b"),
						}),
						"test_false_bool":   tftypes.NewValue(tftypes.Bool, false),
						"test_empty_string": tftypes.NewValue(tftypes.String, ""),
						"test_deprecated_block": tftypes.NewValue(tftypes.List{ElementType: testNestedBlockType}, []tftypes.Value{
							tftypes.NewValue(testNestedBlockType, map[string]tftypes.Value{
								"test_nested_block_attr": tftypes.NewValue(tftypes.String, "test-nested-block-val-a"),
							}),
							tftypes.NewValue(testNestedBlockType, map[string]tftypes.Value{
								"test_nested_block_attr": tftypes.NewValue(tftypes.String, "test-nested-block-val-b"),
							}),
						}),
					}),
					Schema: testSchema,
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"id":                    tftypes.NewValue(tftypes.String, nil),
						"test_computed":         tftypes.NewValue(tftypes.String, nil),
						"test_optional":         tftypes.NewValue(tftypes.String, "test-optional-val"),
						"test_required":         tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_deprecated":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_false_bool":       tftypes.NewValue(tftypes.Bool, false),
						"test_empty_string":     tftypes.NewValue(tftypes.String, nil),
						"test_deprecated_block": tftypes.NewValue(tftypes.List{ElementType: testNestedBlockType}, nil),
					}),
					Schema: testSchema,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.configureProviderReq != nil {
				configureProviderResp := &provider.ConfigureResponse{}
				testCase.server.ConfigureProvider(t.Context(), testCase.configureProviderReq, configureProviderResp)
			}

			response := &fwserver.GenerateResourceConfigResponse{}
			testCase.server.GenerateResourceConfig(t.Context(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
