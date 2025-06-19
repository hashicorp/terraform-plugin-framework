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

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.ListRequest
		expectedStreamEvents []fwserver.ListResult
		expectedError        string
	}{
		"success-with-zero-results": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListRequest{
				Config: &tfsdk.Config{},
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
						resp.Results = list.NoListResults
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
				Config: &tfsdk.Config{},
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
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
				Config: &tfsdk.Config{},
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
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
		"error-on-nil-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListRequest{
				Config: nil,
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
						resp.Results = list.NoListResults
					},
				},
			},
			expectedError:        "Invalid ListResource request: Config cannot be nil",
			expectedStreamEvents: []fwserver.ListResult{},
		},
		"error-on-nil-resource-identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListRequest{
				Config: &tfsdk.Config{},
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
						resp.Results = slices.Values([]list.ListResult{
							{
								Identity: nil,
								Resource: &tfsdk.Resource{
									Schema: testSchema,
									Raw:    testResourceValue1,
								},
								DisplayName: "Test Resource 1",
							},
						})
					},
				},
			},
			expectedStreamEvents: []fwserver.ListResult{
				{
					Diagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("Incomplete List Result", "..."),
					},
				},
			},
		},
		"error-on-null-resource-identity": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListRequest{
				Config: &tfsdk.Config{},
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
						resp.Results = slices.Values([]list.ListResult{
							{
								Identity: &tfsdk.ResourceIdentity{},
								Resource: &tfsdk.Resource{
									Schema: testSchema,
									Raw:    testResourceValue1,
								},
								DisplayName: "Test Resource 1",
							},
						})
					},
				},
			},
			expectedStreamEvents: []fwserver.ListResult{
				{
					Diagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("Incomplete List Result", "..."),
					},
				},
			},
		},
		"warning-on-missing-resource": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListRequest{
				Config:          &tfsdk.Config{},
				IncludeResource: true,
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
						resp.Results = slices.Values([]list.ListResult{
							{
								Identity: &tfsdk.ResourceIdentity{
									Schema: testIdentitySchema,
									Raw:    testIdentityValue1,
								},
								Resource:    nil,
								DisplayName: "Test Resource 1",
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
					DisplayName: "Test Resource 1",
					Diagnostics: diag.Diagnostics{
						diag.NewWarningDiagnostic("Incomplete List Result", "..."),
					},
				},
			},
		},
		"warning-on-null-resource": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ListRequest{
				Config:          &tfsdk.Config{},
				IncludeResource: true,
				ListResource: &testprovider.ListResource{
					ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
						resp.Results = slices.Values([]list.ListResult{
							{
								Identity: &tfsdk.ResourceIdentity{
									Schema: testIdentitySchema,
									Raw:    testIdentityValue1,
								},
								Resource:    &tfsdk.Resource{},
								DisplayName: "Test Resource 1",
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
					Resource:    &tfsdk.Resource{},
					DisplayName: "Test Resource 1",
					Diagnostics: diag.Diagnostics{
						diag.NewWarningDiagnostic("Incomplete List Result", "..."),
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ListResultsStream{}
			err := testCase.server.ListResource(context.Background(), testCase.request, response)
			if err != nil && err.Error() != testCase.expectedError {
				t.Fatalf("unexpected error: %s", err)
			}

			opts := cmp.Options{
				cmp.Comparer(func(a, b diag.Diagnostics) bool {
					// Differences in Detail() are not relevant to correctness of logic
					for i := range a {
						if a[i].Severity() != b[i].Severity() || a[i].Summary() != b[i].Summary() {
							return false
						}
					}
					return true
				}),
			}
			events := slices.AppendSeq([]fwserver.ListResult{}, response.Results)
			if diff := cmp.Diff(events, testCase.expectedStreamEvents, opts); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
