package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerReadResource(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		currentState tftypes.Value
		providerMeta tftypes.Value
		private      []byte
		resource     string
		resourceType tftypes.Type

		impl func(context.Context, tfsdk.ReadResourceRequest, *tfsdk.ReadResourceResponse)

		// response expectations
		expectedNewState tftypes.Value
		expectedDiags    []*tfprotov6.Diagnostic
		expectedPrivate  []byte
	}

	tests := map[string]testCase{
		"one_basic": {
			currentState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "foo"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, "a minute ago, but like, as a timestamp"),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,

			impl: func(_ context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "foo"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "orange"),
						tftypes.NewValue(tftypes.String, "yellow"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "now"),
				})
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "foo"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "now"),
			}),
		},
		"one_provider_meta": {
			currentState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "my name"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "a long, long time ago"),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,

			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),

			impl: func(_ context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "my name"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "a long, long time ago"),
				})
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "my name"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "blue"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "a long, long time ago"),
			}),
		},
		"one_remove": {
			currentState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "my name"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "a long, long time ago"),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,

			impl: func(_ context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, nil)
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, nil),
		},
		"two_basic": {
			currentState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123foo"),
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					},
				}, tftypes.UnknownValue),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, tftypes.UnknownValue),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,

			impl: func(_ context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "123foo"),
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 100),
							"boot":    tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
					"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_bool":   tftypes.Bool,
								"required_number": tftypes.Number,
								"required_string": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_bool":   tftypes.NewValue(tftypes.Bool, true),
							"required_number": tftypes.NewValue(tftypes.Number, 123),
							"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
						}),
					}),
				})
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123foo"),
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 100),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
		},
		"two_diags": {
			currentState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123foo"),
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					},
				}, tftypes.UnknownValue),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, tftypes.UnknownValue),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,

			impl: func(_ context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "123foo"),
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 100),
							"boot":    tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
					"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_bool":   tftypes.Bool,
								"required_number": tftypes.Number,
								"required_string": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_bool":   tftypes.NewValue(tftypes.Bool, true),
							"required_number": tftypes.NewValue(tftypes.Number, 123),
							"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
						}),
					}),
				})
				resp.Diagnostics.AddAttributeWarning(
					tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
					"This is a warning",
					"This is your final warning",
				)
				resp.Diagnostics.AddError(
					"This is an error",
					"Oops.",
				)
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123foo"),
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 100),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Summary:   "This is a warning",
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Detail:    "This is your final warning",
					Attribute: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
				},
				{
					Summary:  "This is an error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops.",
				},
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := &testServeProvider{
				readResourceImpl: tc.impl,
			}
			testServer := &Server{
				FrameworkServer: fwserver.Server{
					Provider: s,
				},
			}
			var pmSchema tfsdk.Schema
			if tc.providerMeta.Type() != nil {
				testServer.FrameworkServer.Provider = &testServeProviderWithMetaSchema{s}
				schema, diags := testServer.FrameworkServer.ProviderMetaSchema(context.Background())
				if len(diags) > 0 {
					t.Errorf("Unexpected diags: %+v", diags)
					return
				}
				pmSchema = *schema
			}

			rt, diags := testServer.FrameworkServer.ResourceType(context.Background(), tc.resource)
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}
			schema, diags := rt.GetSchema(context.Background())
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}

			dv, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.currentState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ReadResourceRequest{
				TypeName:     tc.resource,
				Private:      tc.private,
				CurrentState: &dv,
			}
			if tc.providerMeta.Type() != nil {
				providerMeta, err := tfprotov6.NewDynamicValue(testServeProviderMetaType, tc.providerMeta)
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				req.ProviderMeta = &providerMeta
			}
			got, err := testServer.ReadResource(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.readResourceCalledResourceType != tc.resource {
				t.Errorf("Called wrong resource. Expected to call %q, actually called %q", tc.resource, s.readResourceCalledResourceType)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			if diff := cmp.Diff(s.readResourceCurrentStateValue, tc.currentState); diff != "" {
				t.Errorf("Unexpected diff in current state (+wanted, -got): %s", diff)
				return
			}
			if diff := cmp.Diff(s.readResourceCurrentStateSchema, schema); diff != "" {
				t.Errorf("Unexpected diff in state schema (+wanted, -got): %s", diff)
				return
			}
			if tc.providerMeta.Type() != nil {
				if diff := cmp.Diff(s.readResourceProviderMetaValue, tc.providerMeta); diff != "" {
					t.Errorf("Unexpected diff in provider meta (+wanted, -got): %s", diff)
					return
				}
				if diff := cmp.Diff(s.readResourceProviderMetaSchema, pmSchema); diff != "" {
					t.Errorf("Unexpected diff in provider meta schema (+wanted, -got): %s", diff)
					return
				}
			}
			gotNewState, err := got.NewState.Unmarshal(tc.resourceType)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(gotNewState, tc.expectedNewState); diff != "" {
				t.Errorf("Unexpected diff in new state (+wanted, -got): %s", diff)
				return
			}
			if string(got.Private) != string(tc.expectedPrivate) {
				t.Errorf("Expected private to be %q, got %q", tc.expectedPrivate, got.Private)
				return
			}
		})
	}
}
