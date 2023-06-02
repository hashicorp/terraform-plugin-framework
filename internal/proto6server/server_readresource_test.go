// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	testCurrentStateValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, nil),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testEmptyDynamicValue := testNewDynamicValue(t, tftypes.Object{}, nil)

	testNewStateDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-newstate-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testNewStateRemovedDynamicValue, _ := tfprotov6.NewDynamicValue(testType, tftypes.NewValue(testType, nil))

	testProviderMetaDynamicValue := testNewDynamicValue(t,
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"test_optional": tftypes.String,
				"test_required": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"test_optional": tftypes.NewValue(tftypes.String, nil),
			"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
		},
	)

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

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.ReadResourceRequest
		expectedError    error
		expectedResponse *tfprotov6.ReadResourceResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testEmptyDynamicValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testEmptyDynamicValue,
			},
		},
		"request-currentstate": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
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
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testCurrentStateValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testCurrentStateValue,
			},
		},
		"request-providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithMetaSchema{
						Provider: &testprovider.Provider{
							ResourcesMethod: func(_ context.Context) []func() resource.Resource {
								return []func() resource.Resource{
									func() resource.Resource {
										return &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_resource"
											},
											ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
												var data struct {
													TestOptional types.String `tfsdk:"test_optional"`
													TestRequired types.String `tfsdk:"test_required"`
												}

												resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

												if data.TestRequired.ValueString() != "test-config-value" {
													resp.Diagnostics.AddError("Unexpected req.ProviderMeta", data.TestRequired.ValueString())
												}
											},
										}
									},
								}
							},
						},
						MetaSchemaMethod: func(_ context.Context, _ provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
							resp.Schema = metaschema.Schema{
								Attributes: map[string]metaschema.Attribute{
									"test_optional": metaschema.StringAttribute{
										Optional: true,
									},
									"test_required": metaschema.StringAttribute{
										Required: true,
									},
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testEmptyDynamicValue,
				ProviderMeta: testProviderMetaDynamicValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testEmptyDynamicValue,
			},
		},
		"request-private": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
										ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
											expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`
											got, diags := req.Private.GetKey(ctx, "providerKey")

											resp.Diagnostics.Append(diags...)

											if string(got) != expected {
												resp.Diagnostics.AddError(
													"Unexpected req.Private Value",
													fmt.Sprintf("expected %q, got %q", expected, got),
												)
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testEmptyDynamicValue,
				TypeName:     "test_resource",
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKey":   []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testEmptyDynamicValue,
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKey":   []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
										ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
											resp.Diagnostics.AddWarning("warning summary", "warning detail")
											resp.Diagnostics.AddError("error summary", "error detail")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testCurrentStateValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
				},
				NewState: testCurrentStateValue,
			},
		},
		"response-state": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
										ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
											var data struct {
												TestComputed types.String `tfsdk:"test_computed"`
												TestRequired types.String `tfsdk:"test_required"`
											}

											resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

											data.TestComputed = types.StringValue("test-newstate-value")

											resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testCurrentStateValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testNewStateDynamicValue,
			},
		},
		"response-state-removeresource": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
										ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
											resp.State.RemoveResource(ctx)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testCurrentStateValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: &testNewStateRemovedDynamicValue,
			},
		},
		"response-private": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
										ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
											diags := resp.Private.SetKey(ctx, "providerKey", []byte(`{"key": "value"}`))

											resp.Diagnostics.Append(diags...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.ReadResourceRequest{
				CurrentState: testEmptyDynamicValue,
				TypeName:     "test_resource",
			},
			expectedResponse: &tfprotov6.ReadResourceResponse{
				NewState: testEmptyDynamicValue,
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					"providerKey": []byte(`{"key": "value"}`),
				}),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ReadResource(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
