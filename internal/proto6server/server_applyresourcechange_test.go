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

func TestServerApplyResourceChange(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		priorState     tftypes.Value
		plannedState   tftypes.Value
		config         tftypes.Value
		plannedPrivate []byte
		providerMeta   tftypes.Value
		resource       string
		action         string
		resourceType   tftypes.Type

		create  func(context.Context, tfsdk.CreateResourceRequest, *tfsdk.CreateResourceResponse)
		update  func(context.Context, tfsdk.UpdateResourceRequest, *tfsdk.UpdateResourceResponse)
		destroy func(context.Context, tfsdk.DeleteResourceRequest, *tfsdk.DeleteResourceResponse)

		// response expectations
		expectedNewState tftypes.Value
		expectedDiags    []*tfprotov6.Diagnostic
		expectedPrivate  []byte
	}

	tests := map[string]testCase{
		"one_create": {
			plannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			action:       "create",
			resourceType: testServeResourceTypeOneType,
			create: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "hello, world"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
				})
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
		},
		"one_create_diags": {
			plannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			action:       "create",
			resourceType: testServeResourceTypeOneType,
			create: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "hello, world"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
				})
				resp.Diagnostics.AddAttributeWarning(
					tftypes.NewAttributePath().WithAttributeName("favorite_colors").WithElementKeyInt(0),
					"This is a warning",
					"I'm warning you",
				)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Summary:   "This is a warning",
					Detail:    "I'm warning you",
					Attribute: tftypes.NewAttributePath().WithAttributeName("favorite_colors").WithElementKeyInt(0),
				},
			},
		},
		"one_update": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			plannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			action:       "update",
			resourceType: testServeResourceTypeOneType,
			update: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "hello, world"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "orange"),
						tftypes.NewValue(tftypes.String, "yellow"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
				})
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
		},
		"one_update_diags": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			plannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			action:       "update",
			resourceType: testServeResourceTypeOneType,
			update: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "hello, world"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "orange"),
						tftypes.NewValue(tftypes.String, "yellow"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
				})
				resp.Diagnostics.AddAttributeWarning(
					tftypes.NewAttributePath().WithAttributeName("name"),
					"I'm warning you...",
					"This is a warning!",
				)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Summary:   "I'm warning you...",
					Detail:    "This is a warning!",
					Attribute: tftypes.NewAttributePath().WithAttributeName("name"),
				},
			},
		},
		"one_update_diags_error": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			plannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			action:       "update",
			resourceType: testServeResourceTypeOneType,
			update: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "hello, world"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "orange"),
						tftypes.NewValue(tftypes.String, "yellow"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
				})
				resp.Diagnostics.AddAttributeError(
					tftypes.NewAttributePath().WithAttributeName("name"),
					"Oops!",
					"This is an error! Don't update the state!",
				)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity:  tfprotov6.DiagnosticSeverityError,
					Summary:   "Oops!",
					Detail:    "This is an error! Don't update the state!",
					Attribute: tftypes.NewAttributePath().WithAttributeName("name"),
				},
			},
		},
		"one_delete": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			resource:     "test_one",
			action:       "delete",
			resourceType: testServeResourceTypeOneType,
			destroy: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
				// Removing the state prior to the framework should not generate errors
				resp.State.RemoveResource(ctx)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, nil),
		},
		"one_delete_diags": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			resource:     "test_one",
			action:       "delete",
			resourceType: testServeResourceTypeOneType,
			destroy: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
				// Removing the state prior to the framework should not generate errors
				resp.State.RemoveResource(ctx)
				resp.Diagnostics.AddAttributeWarning(
					tftypes.NewAttributePath().WithAttributeName("created_timestamp"),
					"This is a warning",
					"just a warning diagnostic, no behavior changes",
				)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, nil),
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Summary:   "This is a warning",
					Detail:    "just a warning diagnostic, no behavior changes",
					Attribute: tftypes.NewAttributePath().WithAttributeName("created_timestamp"),
				},
			},
		},
		"one_delete_diags_error": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			resource:     "test_one",
			action:       "delete",
			resourceType: testServeResourceTypeOneType,
			destroy: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
				resp.Diagnostics.AddError(
					"This is an error",
					"Something went wrong, keep the old state around",
				)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  "This is an error",
					Detail:   "Something went wrong, keep the old state around",
				},
			},
		},
		"one_delete_automatic_removeresource": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			resource:     "test_one",
			action:       "delete",
			resourceType: testServeResourceTypeOneType,
			destroy: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
				// The framework should automatically call resp.State.RemoveResource()
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, nil),
		},
		"one_delete_diags_warning_automatic_removeresource": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			resource:     "test_one",
			action:       "delete",
			resourceType: testServeResourceTypeOneType,
			destroy: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
				// The framework should automatically call resp.State.RemoveResource()
				resp.Diagnostics.AddAttributeWarning(
					tftypes.NewAttributePath().WithAttributeName("created_timestamp"),
					"This is a warning",
					"just a warning diagnostic, no behavior changes",
				)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, nil),
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Summary:   "This is a warning",
					Detail:    "just a warning diagnostic, no behavior changes",
					Attribute: tftypes.NewAttributePath().WithAttributeName("created_timestamp"),
				},
			},
		},
		"one_delete_diags_error_no_automatic_removeresource": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			resource:     "test_one",
			action:       "delete",
			resourceType: testServeResourceTypeOneType,
			destroy: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
				// The framework should NOT automatically call resp.State.RemoveResource()
				resp.Diagnostics.AddError(
					"This is an error",
					"Something went wrong, keep the old state around",
				)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  "This is an error",
					Detail:   "Something went wrong, keep the old state around",
				},
			},
		},
		"two_create": {
			plannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, tftypes.UnknownValue),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, tftypes.UnknownValue),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, nil),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, nil),
			}),
			resource:     "test_two",
			action:       "create",
			resourceType: testServeResourceTypeTwoType,
			create: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "test-instance"),
					"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 123),
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
							"required_string": tftypes.NewValue(tftypes.String, "stringvalue"),
						}),
					}),
				})
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 123),
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
						"required_string": tftypes.NewValue(tftypes.String, "stringvalue"),
					}),
				}),
			}),
		},
		"two_update": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 123),
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
			plannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1234),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 2345),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
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
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, false),
						"required_number": tftypes.NewValue(tftypes.Number, 456),
						"required_string": tftypes.NewValue(tftypes.String, "newvalue"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1234),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 2345),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
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
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, false),
						"required_number": tftypes.NewValue(tftypes.Number, 456),
						"required_string": tftypes.NewValue(tftypes.String, "newvalue"),
					}),
				}),
			}),
			resource:     "test_two",
			action:       "update",
			resourceType: testServeResourceTypeTwoType,
			update: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "test-instance"),
					"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 1234),
							"boot":    tftypes.NewValue(tftypes.Bool, true),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 2345),
							"boot":    tftypes.NewValue(tftypes.Bool, false),
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
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_bool":   tftypes.Bool,
								"required_number": tftypes.Number,
								"required_string": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_bool":   tftypes.NewValue(tftypes.Bool, false),
							"required_number": tftypes.NewValue(tftypes.Number, 456),
							"required_string": tftypes.NewValue(tftypes.String, "newvalue"),
						}),
					}),
				})
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1234),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 2345),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
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
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, false),
						"required_number": tftypes.NewValue(tftypes.Number, 456),
						"required_string": tftypes.NewValue(tftypes.String, "newvalue"),
					}),
				}),
			}),
		},
		"two_delete": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1234),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 2345),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
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
			resource:     "test_two",
			action:       "delete",
			resourceType: testServeResourceTypeTwoType,
			destroy: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, nil)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeTwoType, nil),
		},
		"one_meta_create": {
			plannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),
			resource:     "test_one",
			action:       "create",
			resourceType: testServeResourceTypeOneType,
			create: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "hello, world"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
				})
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
		},
		"one_meta_update": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			plannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),
			resource:     "test_one",
			action:       "update",
			resourceType: testServeResourceTypeOneType,
			update: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "hello, world"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "orange"),
						tftypes.NewValue(tftypes.String, "yellow"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
				})
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
		},
		"one_meta_delete": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "right now I guess"),
			}),
			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),
			resource:     "test_one",
			action:       "delete",
			resourceType: testServeResourceTypeOneType,
			destroy: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, nil)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, nil),
		},
		"two_meta_create": {
			plannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, tftypes.UnknownValue),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, tftypes.UnknownValue),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, nil),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, nil),
			}),
			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),
			resource:     "test_two",
			action:       "create",
			resourceType: testServeResourceTypeTwoType,
			create: func(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "test-instance"),
					"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 123),
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
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 123),
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
		"two_meta_update": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 123),
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
			plannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1234),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 2345),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
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
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1234),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 2345),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
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
			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),
			resource:     "test_two",
			action:       "update",
			resourceType: testServeResourceTypeTwoType,
			update: func(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "test-instance"),
					"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 1234),
							"boot":    tftypes.NewValue(tftypes.Bool, true),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 2345),
							"boot":    tftypes.NewValue(tftypes.Bool, false),
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
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1234),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 2345),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
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
		"two_meta_delete": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-instance"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1234),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-other-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 2345),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
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
			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),
			resource:     "test_two",
			action:       "delete",
			resourceType: testServeResourceTypeTwoType,
			destroy: func(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, nil)
			},
			expectedNewState: tftypes.NewValue(testServeResourceTypeTwoType, nil),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := &testServeProvider{
				createFunc: tc.create,
				updateFunc: tc.update,
				deleteFunc: tc.destroy,
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

			priorState, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.priorState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			plannedState, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.plannedState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			config, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ApplyResourceChangeRequest{
				TypeName:       tc.resource,
				PlannedPrivate: tc.plannedPrivate,
				PriorState:     &priorState,
				PlannedState:   &plannedState,
				Config:         &config,
			}
			if tc.providerMeta.Type() != nil {
				providerMeta, err := tfprotov6.NewDynamicValue(testServeProviderMetaType, tc.providerMeta)
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				req.ProviderMeta = &providerMeta
			}
			got, err := testServer.ApplyResourceChange(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			if s.applyResourceChangeCalledResourceType != tc.resource {
				t.Errorf("Called wrong resource. Expected to call %q, actually called %q", tc.resource, s.applyResourceChangeCalledResourceType)
				return
			}
			if s.applyResourceChangeCalledAction != tc.action {
				t.Errorf("Called wrong action. Expected to call %q, actually called %q", tc.action, s.applyResourceChangeCalledAction)
				return
			}
			if tc.priorState.Type() != nil {
				if diff := cmp.Diff(s.applyResourceChangePriorStateValue, tc.priorState); diff != "" {
					t.Errorf("Unexpected diff in prior state (+wanted, -got): %s", diff)
					return
				}
				if diff := cmp.Diff(s.applyResourceChangePriorStateSchema, schema); diff != "" {
					t.Errorf("Unexpected diff in prior state schema (+wanted, -got): %s", diff)
					return
				}
			}
			if tc.plannedState.Type() != nil {
				if diff := cmp.Diff(s.applyResourceChangePlannedStateValue, tc.plannedState); diff != "" {
					t.Errorf("Unexpected diff in planned state (+wanted, -got): %s", diff)
					return
				}
				if diff := cmp.Diff(s.applyResourceChangePlannedStateSchema, schema); diff != "" {
					t.Errorf("Unexpected diff in planned state schema (+wanted, -got): %s", diff)
					return
				}
			}
			if tc.config.Type() != nil {
				if diff := cmp.Diff(s.applyResourceChangeConfigValue, tc.config); diff != "" {
					t.Errorf("Unexpected diff in config (+wanted, -got): %s", diff)
					return
				}
				if diff := cmp.Diff(s.applyResourceChangeConfigSchema, schema); diff != "" {
					t.Errorf("Unexpected diff in config schema (+wanted, -got): %s", diff)
					return
				}
			}
			if tc.providerMeta.Type() != nil {
				if diff := cmp.Diff(s.applyResourceChangeProviderMetaValue, tc.providerMeta); diff != "" {
					t.Errorf("Unexpected diff in provider meta (+wanted, -got): %s", diff)
					return
				}
				if diff := cmp.Diff(s.applyResourceChangeProviderMetaSchema, pmSchema); diff != "" {
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
