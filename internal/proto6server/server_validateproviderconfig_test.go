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

func TestServerValidateProviderConfig(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		config       tftypes.Value
		provider     tfsdk.Provider
		providerType tftypes.Type

		// response expectations
		expectedDiags []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"no_validation": {
			config: tftypes.NewValue(testServeProviderProviderType, map[string]tftypes.Value{
				"required":          tftypes.NewValue(tftypes.String, "this is a required value"),
				"optional":          tftypes.NewValue(tftypes.String, nil),
				"computed":          tftypes.NewValue(tftypes.String, nil),
				"optional_computed": tftypes.NewValue(tftypes.String, "they filled this one out"),
				"sensitive":         tftypes.NewValue(tftypes.String, "hunter42"),
				"deprecated":        tftypes.NewValue(tftypes.String, "oops"),
				"string":            tftypes.NewValue(tftypes.String, "a new string value"),
				"number":            tftypes.NewValue(tftypes.Number, 1234),
				"bool":              tftypes.NewValue(tftypes.Bool, true),
				"int64":             tftypes.NewValue(tftypes.Number, 1234),
				"float64":           tftypes.NewValue(tftypes.Number, 1234),
				"list-string": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"list-list-string": tftypes.NewValue(tftypes.List{ElementType: tftypes.List{ElementType: tftypes.String}}, []tftypes.Value{
					tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
					tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "rojo"),
						tftypes.NewValue(tftypes.String, "azul"),
						tftypes.NewValue(tftypes.String, "verde"),
					}),
				}),
				"list-object": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Bool,
					"baz": tftypes.Number,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Bool,
						"baz": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "hello, world"),
						"bar": tftypes.NewValue(tftypes.Bool, true),
						"baz": tftypes.NewValue(tftypes.Number, 4567),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Bool,
						"baz": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "goodnight, moon"),
						"bar": tftypes.NewValue(tftypes.Bool, false),
						"baz": tftypes.NewValue(tftypes.Number, 8675309),
					}),
				}),
				"object": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo":  tftypes.String,
					"bar":  tftypes.Bool,
					"baz":  tftypes.Number,
					"quux": tftypes.List{ElementType: tftypes.String},
				}}, map[string]tftypes.Value{
					"foo": tftypes.NewValue(tftypes.String, "testing123"),
					"bar": tftypes.NewValue(tftypes.Bool, true),
					"baz": tftypes.NewValue(tftypes.Number, 123),
					"quux": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
				}),
				"set-string": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"set-set-string": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Set{ElementType: tftypes.String}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
					tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "rojo"),
						tftypes.NewValue(tftypes.String, "azul"),
						tftypes.NewValue(tftypes.String, "verde"),
					}),
				}),
				"set-object": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Bool,
					"baz": tftypes.Number,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Bool,
						"baz": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "hello, world"),
						"bar": tftypes.NewValue(tftypes.Bool, true),
						"baz": tftypes.NewValue(tftypes.Number, 4567),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Bool,
						"baz": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "goodnight, moon"),
						"bar": tftypes.NewValue(tftypes.Bool, false),
						"baz": tftypes.NewValue(tftypes.Number, 8675309),
					}),
				}),
				"empty-object": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
				"single-nested-attributes": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}, map[string]tftypes.Value{
					"foo": tftypes.NewValue(tftypes.String, "almost done"),
					"bar": tftypes.NewValue(tftypes.Number, 12),
				}),
				"list-nested-attributes": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "let's do the math"),
						"bar": tftypes.NewValue(tftypes.Number, 18973),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "this is why we can't have nice things"),
						"bar": tftypes.NewValue(tftypes.Number, 14554216),
					}),
				}),
				"list-nested-blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "let's do the math"),
						"bar": tftypes.NewValue(tftypes.Number, 18973),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "this is why we can't have nice things"),
						"bar": tftypes.NewValue(tftypes.Number, 14554216),
					}),
				}),
				"map": tftypes.NewValue(tftypes.Map{ElementType: tftypes.Number}, map[string]tftypes.Value{
					"foo": tftypes.NewValue(tftypes.Number, 123),
					"bar": tftypes.NewValue(tftypes.Number, 456),
					"baz": tftypes.NewValue(tftypes.Number, 789),
				}),
				"map-nested-attributes": tftypes.NewValue(tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"bar": tftypes.Number,
					"foo": tftypes.String,
				}}}, map[string]tftypes.Value{
					"hello": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"bar": tftypes.Number,
						"foo": tftypes.String,
					}}, map[string]tftypes.Value{
						"bar": tftypes.NewValue(tftypes.Number, 123456),
						"foo": tftypes.NewValue(tftypes.String, "world"),
					}),
					"goodnight": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"bar": tftypes.Number,
						"foo": tftypes.String,
					}}, map[string]tftypes.Value{
						"bar": tftypes.NewValue(tftypes.Number, 56789),
						"foo": tftypes.NewValue(tftypes.String, "moon"),
					}),
				}),
				"set-nested-attributes": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "let's do the math"),
						"bar": tftypes.NewValue(tftypes.Number, 18973),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "this is why we can't have nice things"),
						"bar": tftypes.NewValue(tftypes.Number, 14554216),
					}),
				}),
				"set-nested-blocks": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "let's do the math"),
						"bar": tftypes.NewValue(tftypes.Number, 18973),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "this is why we can't have nice things"),
						"bar": tftypes.NewValue(tftypes.Number, 14554216),
					}),
				}),
			}),
			provider:     &testServeProvider{},
			providerType: testServeProviderProviderType,

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Summary:   "Attribute Deprecated",
					Detail:    `Deprecated, please use "optional" instead`,
					Attribute: tftypes.NewAttributePath().WithAttributeName("deprecated"),
				},
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "Deprecated",
					Detail:   "Deprecated in favor of other_resource",
				},
			},
		},
		"config_validators_no_diags": {
			config: tftypes.NewValue(testServeResourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			provider: &testServeProviderWithConfigValidators{
				&testServeProvider{
					validateProviderConfigImpl: func(_ context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
					},
				},
			},
			providerType: testServeProviderWithConfigValidatorsType,
		},
		"config_validators_one_diag": {
			config: tftypes.NewValue(testServeResourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			provider: &testServeProviderWithConfigValidators{
				&testServeProvider{
					validateProviderConfigImpl: func(_ context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
						if len(resp.Diagnostics) == 0 {
							resp.Diagnostics.AddError(
								"This is an error",
								"Oops.",
							)
						} else {
							resp.Diagnostics.AddError(
								"This is another error",
								"Oops again.",
							)
						}
					},
				},
			},
			providerType: testServeProviderWithConfigValidatorsType,

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Summary:  "This is an error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops.",
				},
				// ConfigValidators includes multiple calls
				{
					Summary:  "This is another error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops again.",
				},
			},
		},
		"config_validators_two_diags": {
			config: tftypes.NewValue(testServeResourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			provider: &testServeProviderWithConfigValidators{
				&testServeProvider{
					validateProviderConfigImpl: func(_ context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
						if len(resp.Diagnostics) == 0 {
							resp.Diagnostics.AddAttributeWarning(
								tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
								"This is a warning",
								"This is your final warning",
							)
							resp.Diagnostics.AddError(
								"This is an error",
								"Oops.",
							)
						} else {
							resp.Diagnostics.AddAttributeWarning(
								tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
								"This is another warning",
								"This is really your final warning",
							)
							resp.Diagnostics.AddError(
								"This is another error",
								"Oops again.",
							)
						}
					},
				},
			},
			providerType: testServeProviderWithConfigValidatorsType,

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
				// ConfigValidators includes multiple calls
				{
					Summary:   "This is another warning",
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Detail:    "This is really your final warning",
					Attribute: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
				},
				{
					Summary:  "This is another error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops again.",
				},
			},
		},
		"validate_config_no_diags": {
			config: tftypes.NewValue(testServeResourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			provider: &testServeProviderWithValidateConfig{
				&testServeProvider{
					validateProviderConfigImpl: func(_ context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
					},
				},
			},
			providerType: testServeProviderWithValidateConfigType,
		},
		"validate_config_one_diag": {
			config: tftypes.NewValue(testServeResourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			provider: &testServeProviderWithValidateConfig{
				&testServeProvider{
					validateProviderConfigImpl: func(_ context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
						resp.Diagnostics.AddError(
							"This is an error",
							"Oops.",
						)
					},
				},
			},
			providerType: testServeProviderWithValidateConfigType,

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Summary:  "This is an error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops.",
				},
			},
		},
		"validate_config_two_diags": {
			config: tftypes.NewValue(testServeResourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			provider: &testServeProviderWithValidateConfig{
				&testServeProvider{
					validateProviderConfigImpl: func(_ context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
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
				},
			},
			providerType: testServeProviderWithValidateConfigType,

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

			testServer := &Server{
				FrameworkServer: fwserver.Server{
					Provider: tc.provider,
				},
			}

			dv, err := tfprotov6.NewDynamicValue(tc.providerType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ValidateProviderConfigRequest{
				Config: &dv,
			}
			got, err := testServer.ValidateProviderConfig(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}
