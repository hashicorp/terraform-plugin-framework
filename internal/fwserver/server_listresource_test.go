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

	testResourceValue1 := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_attribute": tftypes.NewValue(tftypes.String, "test-value-1"),
	})

	testResourceValue2 := tftypes.NewValue(testType, map[string]tftypes.Value{
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
		request              *fwserver.ListRequest
		expectedStreamEvents []fwserver.ListResult
	}{
		"success-with-zero-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListRequest{
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) { // TODO
						resp.Results = slices.Values([]list.ListResult{})
					},
				},
			},
			expectedStreamEvents: []fwserver.ListResult{},
		},
		"success-with-nil-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListRequest{
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) { // TODO
						// Do nothing, so that resp.Results is nil
					},
				},
			},
			expectedStreamEvents: []fwserver.ListResult{},
		},

		"success-with-multiple-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListRequest{
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) { // TODO
						resp.Results = slices.Values([]list.ListResult{
							{
								Identity: &tfsdk.ResourceIdentity{
									Schema: testIdentitySchema,
									Raw:    testIdentityValue1,
								},
								Resource: &tfsdk.Resource{
									Schema: testSchema,
									Raw:    testResourceValue1,
								},
								DisplayName: "Test Resource 1",
								Diagnostics: diag.Diagnostics{},
							},
							{
								Identity: &tfsdk.ResourceIdentity{
									Schema: testIdentitySchema,
									Raw:    testIdentityValue2,
								},
								Resource: &tfsdk.Resource{
									Schema: testSchema,
									Raw:    testResourceValue2,
								},
								DisplayName: "Test Resource 2",
								Diagnostics: diag.Diagnostics{},
							},
						})
					},
				},
			},
			expectedStreamEvents: []fwserver.ListResult{
				{
					Identity: &tfsdk.ResourceIdentity{
						Schema: testIdentitySchema,
						Raw:    testIdentityValue1,
					},
					Resource: &tfsdk.Resource{
						Schema: testSchema,
						Raw:    testResourceValue1,
					},
					DisplayName: "Test Resource 1",
					Diagnostics: diag.Diagnostics{},
				},
				{
					Identity: &tfsdk.ResourceIdentity{
						Schema: testIdentitySchema,
						Raw:    testIdentityValue2,
					},
					Resource: &tfsdk.Resource{
						Schema: testSchema,
						Raw:    testResourceValue2,
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

			events := slices.AppendSeq([]fwserver.ListResult{}, response.Results)
			if diff := cmp.Diff(events, testCase.expectedStreamEvents); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
