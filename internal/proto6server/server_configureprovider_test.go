package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerConfigureProvider(t *testing.T) {
	t.Parallel()

	type testCase struct {
		tfVersion     string
		config        tftypes.Value
		expectedDiags []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"basic": {
			tfVersion: "1.0.0",
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
		},
		"config-unknown-value": {
			tfVersion: "1.0.0",
			config: tftypes.NewValue(testServeProviderProviderType, map[string]tftypes.Value{
				"required":          tftypes.NewValue(tftypes.String, "this is a required value"),
				"optional":          tftypes.NewValue(tftypes.String, nil),
				"computed":          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"optional_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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
				"list-list-string": tftypes.NewValue(tftypes.List{ElementType: tftypes.List{ElementType: tftypes.String}}, tftypes.UnknownValue),
				"list-object": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Bool,
					"baz": tftypes.Number,
				}}}, tftypes.UnknownValue),
				"object": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo":  tftypes.String,
					"bar":  tftypes.Bool,
					"baz":  tftypes.Number,
					"quux": tftypes.List{ElementType: tftypes.String},
				}}, map[string]tftypes.Value{
					"foo":  tftypes.NewValue(tftypes.String, "testing123"),
					"bar":  tftypes.NewValue(tftypes.Bool, true),
					"baz":  tftypes.NewValue(tftypes.Number, 123),
					"quux": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue),
				}),
				"set-string": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"set-set-string": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Set{ElementType: tftypes.String}}, tftypes.UnknownValue),
				"set-object": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Bool,
					"baz": tftypes.Number,
				}}}, tftypes.UnknownValue),
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
				}}}, tftypes.UnknownValue),
				"map": tftypes.NewValue(tftypes.Map{ElementType: tftypes.Number}, map[string]tftypes.Value{
					"foo": tftypes.NewValue(tftypes.Number, 123),
					"bar": tftypes.NewValue(tftypes.Number, 456),
					"baz": tftypes.NewValue(tftypes.Number, 789),
				}),
				"list-nested-blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}}, tftypes.UnknownValue),
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
				}}}, tftypes.UnknownValue),
				"set-nested-blocks": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}}, tftypes.UnknownValue),
			}),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := new(testServeProvider)
			testServer := &Server{
				FrameworkServer: fwserver.Server{
					Provider: s,
				},
			}
			dv, err := tfprotov6.NewDynamicValue(testServeProviderProviderType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			providerSchema, diags := s.GetSchema(context.Background())
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}
			got, err := testServer.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{
				TerraformVersion: tc.tfVersion,
				Config:           &dv,
			})
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.configuredTFVersion != tc.tfVersion {
				t.Errorf("Expected Terraform version to be %q, got %q", tc.tfVersion, s.configuredTFVersion)
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			if diff := cmp.Diff(s.configuredVal, tc.config); diff != "" {
				t.Errorf("Unexpected diff in config (+wanted, -got): %s", diff)
				return
			}
			if diff := cmp.Diff(s.configuredSchema, providerSchema); diff != "" {
				t.Errorf("Unexpected diff in schema (+wanted, -got): %s", diff)
				return
			}
		})
	}
}
