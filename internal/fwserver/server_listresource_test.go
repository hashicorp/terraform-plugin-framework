// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerListResource(t *testing.T) {
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

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_attribute": tftypes.String,
		},
	}

	testResourceObjectValue1 := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_attribute": tftypes.NewValue(tftypes.String, "test-value-1"),
	})

	testResourceObjectValue2 := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_attribute": tftypes.NewValue(tftypes.String, "test-value-2"),
	})

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}

	testIdentityType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id": tftypes.String,
		},
	}

	testIdentityValue1 := tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
		"test_id": tftypes.NewValue(tftypes.String, "new-id-123"),
	})

	testIdentityValue2 := tftypes.NewValue(testIdentityType, map[string]tftypes.Value{
		"test_id": tftypes.NewValue(tftypes.String, "new-id-456"),
	})

	// nilIdentityValue := tftypes.NewValue(testIdentityType, nil)

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.ListResourceRequest
		expectedStreamEvents []fwserver.ListResourceEvent
	}{
		"success-with-zero-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListResourceRequest{
				ListResource: &testprovider.ListResource{
					ListResourceMethod: func(ctx context.Context, req list.ListResourceRequest, resp *list.ListResourceResponse) { // TODO
						resp.Results = slices.Values([]list.ListResourceEvent{})
					},
				},
			},
			expectedStreamEvents: []fwserver.ListResourceEvent{},
		},
		"success-with-nil-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListResourceRequest{
				ListResource: &testprovider.ListResource{
					ListResourceMethod: func(ctx context.Context, req list.ListResourceRequest, resp *list.ListResourceResponse) { // TODO
						// Do nothing, so that resp.Results is nil
					},
				},
			},
			expectedStreamEvents: []fwserver.ListResourceEvent{},
		},

		"success-with-multiple-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListResourceRequest{
				ListResource: &testprovider.ListResource{
					ListResourceMethod: func(ctx context.Context, req list.ListResourceRequest, resp *list.ListResourceResponse) { // TODO
						resp.Results = slices.Values([]list.ListResourceEvent{
							{
								Identity: &tfsdk.ResourceIdentity{
									Schema: testIdentitySchema,
									Raw:    testIdentityValue1,
								},
								ResourceObject: &tfsdk.ResourceObject{
									Schema: testSchema,
									Raw:    testResourceObjectValue1,
								},
								DisplayName: "Test Resource 1",
								Diagnostics: diag.Diagnostics{},
							},
							{
								Identity: &tfsdk.ResourceIdentity{
									Schema: testIdentitySchema,
									Raw:    testIdentityValue2,
								},
								ResourceObject: &tfsdk.ResourceObject{
									Schema: testSchema,
									Raw:    testResourceObjectValue2,
								},
								DisplayName: "Test Resource 2",
								Diagnostics: diag.Diagnostics{},
							},
						})
					},
				},
			},
			expectedStreamEvents: []fwserver.ListResourceEvent{
				{
					Identity: &tfsdk.ResourceIdentity{
						Schema: testIdentitySchema,
						Raw:    testIdentityValue1,
					},
					ResourceObject: &tfsdk.ResourceObject{
						Schema: testSchema,
						Raw:    testResourceObjectValue1,
					},
					DisplayName: "Test Resource 1",
					Diagnostics: diag.Diagnostics{},
				},
				{
					Identity: &tfsdk.ResourceIdentity{
						Schema: testIdentitySchema,
						Raw:    testIdentityValue2,
					},
					ResourceObject: &tfsdk.ResourceObject{
						Schema: testSchema,
						Raw:    testResourceObjectValue2,
					},
					DisplayName: "Test Resource 2",
					Diagnostics: diag.Diagnostics{},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ListResourceStream{}
			testCase.server.ListResource(context.Background(), testCase.request, response)

			events := slices.AppendSeq([]fwserver.ListResourceEvent{}, response.Results)
			if diff := cmp.Diff(events, testCase.expectedStreamEvents); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
