package tfsdk

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerCancelInFlightContexts(t *testing.T) {
	t.Parallel()

	// let's test and make sure the code we use to Stop will actually
	// cancel in flight contexts how we expect and not, y'know, crash or
	// something

	// first, let's create a bunch of goroutines
	wg := new(sync.WaitGroup)
	s := &server{}
	testCtx := context.Background()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			ctx = s.registerContext(ctx)
			select {
			case <-time.After(time.Second * 10):
				t.Error("timed out waiting to be canceled")
				return
			case <-ctx.Done():
				return
			}
		}()
	}
	// avoid any race conditions around canceling the contexts before
	// they're all set up
	//
	// we don't need this in prod as, presumably, Terraform would not keep
	// sending us requests after it told us to stop
	time.Sleep(200 * time.Millisecond)

	s.cancelRegisteredContexts(testCtx)

	wg.Wait()
	// if we got here, that means that either all our contexts have been
	// canceled, or we have an error reported
}

func TestMarkComputedNilsAsUnknown(t *testing.T) {
	t.Parallel()

	s := Schema{
		Attributes: map[string]Attribute{
			// values should be left alone
			"string-value": {
				Type:     types.StringType,
				Required: true,
			},
			// nil, uncomputed values should be left alone
			"string-nil": {
				Type:     types.StringType,
				Optional: true,
			},
			// nil computed values should be turned into unknown
			"string-nil-computed": {
				Type:     types.StringType,
				Computed: true,
			},
			// nil computed values should be turned into unknown
			"string-nil-optional-computed": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			// non-nil computed values should be left alone
			"string-value-optional-computed": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			// nil objects should be unknown
			"object-nil-optional-computed": {
				Type: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"string-nil": types.StringType,
						"string-set": types.StringType,
					},
				},
				Optional: true,
				Computed: true,
			},
			// non-nil objects should be left alone
			"object-value-optional-computed": {
				Type: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						// nil attributes of objects
						// should be let alone, as they
						// don't have a schema of their
						// own
						"string-nil": types.StringType,
						"string-set": types.StringType,
					},
				},
				Optional: true,
				Computed: true,
			},
			// nil nested attributes should be unknown
			"nested-nil-optional-computed": {
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"string-nil": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"string-set": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
				}),
				Optional: true,
				Computed: true,
			},
			// non-nil nested attributes should be left alone on the top level
			"nested-value-optional-computed": {
				Attributes: SingleNestedAttributes(map[string]Attribute{
					// nested computed attributes should be unknown
					"string-nil": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					// nested non-nil computed attributes should be left alone
					"string-set": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
				}),
				Optional: true,
				Computed: true,
			},
		},
	}
	input := tftypes.NewValue(s.TerraformType(context.Background()), map[string]tftypes.Value{
		"string-value":                   tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":                     tftypes.NewValue(tftypes.String, nil),
		"string-nil-computed":            tftypes.NewValue(tftypes.String, nil),
		"string-nil-optional-computed":   tftypes.NewValue(tftypes.String, nil),
		"string-value-optional-computed": tftypes.NewValue(tftypes.String, "hello, world"),
		"object-nil-optional-computed":   tftypes.NewValue(s.Attributes["object-nil-optional-computed"].Type.TerraformType(context.Background()), nil),
		"object-value-optional-computed": tftypes.NewValue(s.Attributes["object-value-optional-computed"].Type.TerraformType(context.Background()), map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "foo"),
		}),
		"nested-nil-optional-computed": tftypes.NewValue(s.Attributes["nested-nil-optional-computed"].Attributes.AttributeType().TerraformType(context.Background()), nil),
		"nested-value-optional-computed": tftypes.NewValue(s.Attributes["nested-value-optional-computed"].Attributes.AttributeType().TerraformType(context.Background()), map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "bar"),
		}),
	})
	expected := tftypes.NewValue(s.TerraformType(context.Background()), map[string]tftypes.Value{
		"string-value":                   tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":                     tftypes.NewValue(tftypes.String, nil),
		"string-nil-computed":            tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"string-nil-optional-computed":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"string-value-optional-computed": tftypes.NewValue(tftypes.String, "hello, world"),
		"object-nil-optional-computed":   tftypes.NewValue(s.Attributes["object-nil-optional-computed"].Type.TerraformType(context.Background()), tftypes.UnknownValue),
		"object-value-optional-computed": tftypes.NewValue(s.Attributes["object-value-optional-computed"].Type.TerraformType(context.Background()), map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "foo"),
		}),
		"nested-nil-optional-computed": tftypes.NewValue(s.Attributes["nested-nil-optional-computed"].Attributes.AttributeType().TerraformType(context.Background()), tftypes.UnknownValue),
		"nested-value-optional-computed": tftypes.NewValue(s.Attributes["nested-value-optional-computed"].Attributes.AttributeType().TerraformType(context.Background()), map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"string-set": tftypes.NewValue(tftypes.String, "bar"),
		}),
	})

	got, err := tftypes.Transform(input, markComputedNilsAsUnknown(context.Background(), input, s))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	diff, err := expected.Diff(got)
	if err != nil {
		t.Errorf("Error diffing values: %s", err)
		return
	}
	if len(diff) > 0 {
		t.Errorf("Unexpected diff (value1 expected, value2 got): %v", diff)
	}
}

func TestNormaliseRequiresReplace(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*tftypes.AttributePath
		expected []*tftypes.AttributePath
	}

	tests := map[string]testCase{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"no-duplicates": {
			input: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("name2"),
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
				tftypes.NewAttributePath().WithAttributeName("name1").WithElementKeyString("elementkey"),
			},
			expected: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithAttributeName("name1").WithElementKeyString("elementkey"),
				tftypes.NewAttributePath().WithAttributeName("name2"),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
			},
		},
		"duplicates": {
			input: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
			},
			expected: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := normaliseRequiresReplace(tc.input)

			if diff := cmp.Diff(actual, tc.expected, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
				return
			}
		})
	}
}

func TestServerGetProviderSchema(t *testing.T) {
	t.Parallel()

	s := new(testServeProvider)
	testServer := &server{
		p: s,
	}
	got, err := testServer.GetProviderSchema(context.Background(), new(tfprotov6.GetProviderSchemaRequest))
	if err != nil {
		t.Errorf("Got unexpected error: %s", err)
		return
	}
	expected := &tfprotov6.GetProviderSchemaResponse{
		Provider: testServeProviderProviderSchema,
		ResourceSchemas: map[string]*tfprotov6.Schema{
			"test_one":                      testServeResourceTypeOneSchema,
			"test_two":                      testServeResourceTypeTwoSchema,
			"test_three":                    testServeResourceTypeThreeSchema,
			"test_attribute_plan_modifiers": testServeResourceTypeAttributePlanModifiersSchema,
			"test_config_validators":        testServeResourceTypeConfigValidatorsSchema,
			"test_import_state":             testServeResourceTypeImportStateSchema,
			"test_validate_config":          testServeResourceTypeValidateConfigSchema,
		},
		DataSourceSchemas: map[string]*tfprotov6.Schema{
			"test_one":               testServeDataSourceTypeOneSchema,
			"test_two":               testServeDataSourceTypeTwoSchema,
			"test_config_validators": testServeDataSourceTypeConfigValidatorsSchema,
			"test_validate_config":   testServeDataSourceTypeValidateConfigSchema,
		},
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
	}
}

func TestServerGetProviderSchemaWithProviderMeta(t *testing.T) {
	t.Parallel()

	s := new(testServeProviderWithMetaSchema)
	testServer := &server{
		p: s,
	}
	got, err := testServer.GetProviderSchema(context.Background(), new(tfprotov6.GetProviderSchemaRequest))
	if err != nil {
		t.Errorf("Got unexpected error: %s", err)
		return
	}
	expected := &tfprotov6.GetProviderSchemaResponse{
		Provider: testServeProviderProviderSchema,
		ResourceSchemas: map[string]*tfprotov6.Schema{
			"test_one":                      testServeResourceTypeOneSchema,
			"test_two":                      testServeResourceTypeTwoSchema,
			"test_three":                    testServeResourceTypeThreeSchema,
			"test_attribute_plan_modifiers": testServeResourceTypeAttributePlanModifiersSchema,
			"test_config_validators":        testServeResourceTypeConfigValidatorsSchema,
			"test_import_state":             testServeResourceTypeImportStateSchema,
			"test_validate_config":          testServeResourceTypeValidateConfigSchema,
		},
		DataSourceSchemas: map[string]*tfprotov6.Schema{
			"test_one":               testServeDataSourceTypeOneSchema,
			"test_two":               testServeDataSourceTypeTwoSchema,
			"test_config_validators": testServeDataSourceTypeConfigValidatorsSchema,
			"test_validate_config":   testServeDataSourceTypeValidateConfigSchema,
		},
		ProviderMeta: &tfprotov6.Schema{
			Version: 2,
			Block: &tfprotov6.SchemaBlock{
				Attributes: []*tfprotov6.SchemaAttribute{
					{
						Name:            "foo",
						Required:        true,
						Type:            tftypes.String,
						Description:     "A **string**",
						DescriptionKind: tfprotov6.StringKindMarkdown,
					},
				},
			},
		},
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
	}
}

func TestServerValidateProviderConfig(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		config       tftypes.Value
		provider     Provider
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
					validateProviderConfigImpl: func(_ context.Context, req ValidateProviderConfigRequest, resp *ValidateProviderConfigResponse) {},
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
					validateProviderConfigImpl: func(_ context.Context, req ValidateProviderConfigRequest, resp *ValidateProviderConfigResponse) {
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
					validateProviderConfigImpl: func(_ context.Context, req ValidateProviderConfigRequest, resp *ValidateProviderConfigResponse) {
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
					validateProviderConfigImpl: func(_ context.Context, req ValidateProviderConfigRequest, resp *ValidateProviderConfigResponse) {},
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
					validateProviderConfigImpl: func(_ context.Context, req ValidateProviderConfigRequest, resp *ValidateProviderConfigResponse) {
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
					validateProviderConfigImpl: func(_ context.Context, req ValidateProviderConfigRequest, resp *ValidateProviderConfigResponse) {
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

			testServer := &server{
				p: tc.provider,
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
			}),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := new(testServeProvider)
			testServer := &server{
				p: s,
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

func TestServerValidateResourceConfig(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		config       tftypes.Value
		resource     string
		resourceType tftypes.Type

		impl func(context.Context, ValidateResourceConfigRequest, *ValidateResourceConfigResponse)

		// response expectations
		expectedDiags []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"no_validation": {
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, ""),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, ""),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,
		},
		"config_validators_no_diags": {
			config: tftypes.NewValue(testServeResourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_config_validators",
			resourceType: testServeResourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req ValidateResourceConfigRequest, resp *ValidateResourceConfigResponse) {},
		},
		"config_validators_one_diag": {
			config: tftypes.NewValue(testServeResourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_config_validators",
			resourceType: testServeResourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req ValidateResourceConfigRequest, resp *ValidateResourceConfigResponse) {
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
			resource:     "test_config_validators",
			resourceType: testServeResourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req ValidateResourceConfigRequest, resp *ValidateResourceConfigResponse) {
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
			resource:     "test_validate_config",
			resourceType: testServeResourceTypeValidateConfigType,

			impl: func(_ context.Context, req ValidateResourceConfigRequest, resp *ValidateResourceConfigResponse) {},
		},
		"validate_config_one_diag": {
			config: tftypes.NewValue(testServeResourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_validate_config",
			resourceType: testServeResourceTypeValidateConfigType,

			impl: func(_ context.Context, req ValidateResourceConfigRequest, resp *ValidateResourceConfigResponse) {
				resp.Diagnostics.AddError(
					"This is an error",
					"Oops.",
				)
			},

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
			resource:     "test_validate_config",
			resourceType: testServeResourceTypeValidateConfigType,

			impl: func(_ context.Context, req ValidateResourceConfigRequest, resp *ValidateResourceConfigResponse) {
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
				validateResourceConfigImpl: tc.impl,
			}
			testServer := &server{
				p: s,
			}

			dv, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ValidateResourceConfigRequest{
				TypeName: tc.resource,
				Config:   &dv,
			}
			got, err := testServer.ValidateResourceConfig(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.validateResourceConfigCalledResourceType != tc.resource && !(tc.resource == "test_one" && s.validateResourceConfigCalledResourceType == "") {
				t.Errorf("Called wrong resource. Expected to call %q, actually called %q", tc.resource, s.readDataSourceCalledDataSourceType)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestServerReadResource(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		currentState tftypes.Value
		providerMeta tftypes.Value
		private      []byte
		resource     string
		resourceType tftypes.Type

		impl func(context.Context, ReadResourceRequest, *ReadResourceResponse)

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

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
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

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
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

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
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
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
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
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
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
			testServer := &server{
				p: s,
			}
			var pmSchema Schema
			if tc.providerMeta.Type() != nil {
				sWithMeta := &testServeProviderWithMetaSchema{s}
				testServer.p = sWithMeta
				schema, diags := sWithMeta.GetMetaSchema(context.Background())
				if len(diags) > 0 {
					t.Errorf("Unexpected diags: %+v", diags)
					return
				}
				pmSchema = schema
			}

			rt, diags := testServer.getResourceType(context.Background(), tc.resource)
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

func TestServerPlanResourceChange(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		priorState       tftypes.Value
		proposedNewState tftypes.Value
		config           tftypes.Value
		priorPrivate     []byte
		providerMeta     tftypes.Value
		resource         string
		resourceType     tftypes.Type

		modifyPlanFunc func(context.Context, ModifyResourcePlanRequest, *ModifyResourcePlanResponse)

		// response expectations
		expectedPlannedState    tftypes.Value
		expectedRequiresReplace []*tftypes.AttributePath
		expectedPlannedPrivate  []byte
		expectedDiags           []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"one_changed": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
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
			resourceType: testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		},
		"one_not_changed": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
		},
		"one_nil_state_and_config": {
			priorState:           tftypes.NewValue(testServeResourceTypeOneType, nil),
			proposedNewState:     tftypes.NewValue(testServeResourceTypeOneType, nil),
			config:               tftypes.NewValue(testServeResourceTypeOneType, nil),
			resource:             "test_one",
			resourceType:         testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, nil),
		},
		"two_nil_state_and_config": {
			priorState:           tftypes.NewValue(testServeResourceTypeTwoType, nil),
			proposedNewState:     tftypes.NewValue(testServeResourceTypeTwoType, nil),
			config:               tftypes.NewValue(testServeResourceTypeTwoType, nil),
			resource:             "test_two",
			resourceType:         testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, nil),
		},
		"two_delete": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			proposedNewState:     tftypes.NewValue(testServeResourceTypeTwoType, nil),
			config:               tftypes.NewValue(testServeResourceTypeTwoType, nil),
			resource:             "test_two",
			resourceType:         testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, nil),
		},
		"three_nested_computed_no_changes": {
			resource:     "test_three",
			resourceType: testServeResourceTypeThreeType,
			priorState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
		},
		"three_nested_computed_configuration_change": {
			resource:     "test_three",
			resourceType: testServeResourceTypeThreeType,
			priorState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "newname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "newname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "newname"),
				"last_updated":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"first_updated": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
		},
		"three_nested_computed_nested_configuration_change": {
			resource:     "test_three",
			resourceType: testServeResourceTypeThreeType,
			priorState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, nil),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, nil),
					}),
				}),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"first_updated": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"string":          tftypes.NewValue(tftypes.String, nil),
					}),
				}),
			}),
		},
		"one_add": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, nil),
			proposedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		},
		"two_modifyplan_add_list_elem": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "auto-boot-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
			}),
			modifyPlanFunc: func(ctx context.Context, req ModifyResourcePlanRequest, resp *ModifyResourcePlanResponse) {
				resp.Plan.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "123456"),
					"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						}}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 10),
							"boot":    tftypes.NewValue(tftypes.Bool, false),
						}),
						tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						}}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "auto-boot-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 1),
							"boot":    tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
				})
			},
		},
		"two_modifyplan_requires_replace": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "1234567"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "1234567"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "1234567"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			modifyPlanFunc: func(ctx context.Context, req ModifyResourcePlanRequest, resp *ModifyResourcePlanResponse) {
				resp.RequiresReplace = []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")}
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")},
		},
		"two_modifyplan_diags_warning": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			modifyPlanFunc: func(ctx context.Context, req ModifyResourcePlanRequest, resp *ModifyResourcePlanResponse) {
				resp.RequiresReplace = []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")}
				resp.AddWarning("I'm warning you", "You have been warned")
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")},
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "I'm warning you",
					Detail:   "You have been warned",
				},
			},
		},
		"two_modifyplan_diags_error": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			modifyPlanFunc: func(ctx context.Context, req ModifyResourcePlanRequest, resp *ModifyResourcePlanResponse) {
				resp.RequiresReplace = []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")}
				resp.AddError("This is an error", "More details about the error")
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")},
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  "This is an error",
					Detail:   "More details about the error",
				},
			},
		},
		"attr_plan_modifiers_nil_state_and_config": {
			priorState:           tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, nil),
			proposedNewState:     tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, nil),
			config:               tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, nil),
			resource:             "test_attribute_plan_modifiers",
			resourceType:         testServeResourceTypeAttributePlanModifiersType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, nil),
		},
		"attr_plan_modifiers_requiresreplace": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, map[string]tftypes.Value{
						"size":   tftypes.NewValue(tftypes.Number, 1),
						"format": tftypes.NewValue(tftypes.String, "ext4"),
					}),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, map[string]tftypes.Value{
						"size":   tftypes.NewValue(tftypes.Number, 1),
						"format": tftypes.NewValue(tftypes.String, "ext4"),
					}),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, map[string]tftypes.Value{
						"size":   tftypes.NewValue(tftypes.Number, 1),
						"format": tftypes.NewValue(tftypes.String, "ext4"),
					}),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, map[string]tftypes.Value{
						"size":   tftypes.NewValue(tftypes.Number, 1),
						"format": tftypes.NewValue(tftypes.String, "ext4"),
					}),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedRequiresReplace: []*tftypes.AttributePath{

				tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("filesystem").WithAttributeName("format"),
				tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface"),
			},
		},
		"attr_plan_modifiers_requiresreplaceif_true": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 999),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 999),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 999),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface"), tftypes.NewAttributePath().WithAttributeName("size")},
		},
		"attr_plan_modifiers_requiresreplaceif_false": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 1),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 1),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 1),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
		"attr_plan_modifiers_diags": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "TESTDIAG"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "TESTDIAG"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "TESTDIAG"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "TESTDIAG"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "Warning diag",
					Detail:   "This is a warning",
				},
				{
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  "Error diag",
					Detail:   "This is an error",
				},
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
		"attr_plan_modifiers_chained_modifiers": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:                "test_attribute_plan_modifiers",
			resourceType:            testServeResourceTypeAttributePlanModifiersType,
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
		"attr_plan_modifiers_default_value_modifier": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, nil),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, nil),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
			}),
			resource:                "test_attribute_plan_modifiers",
			resourceType:            testServeResourceTypeAttributePlanModifiersType,
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
		// TODO: Attribute plan modifiers should run before plan unknown marking.
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/183
		// "attr_plan_modifiers_trigger_computed_unknown": {
		// 	resource:     "test_attribute_plan_modifiers",
		// 	resourceType: testServeResourceTypeAttributePlanModifiersType,
		// 	priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
		// 		"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
		// 		"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
		// 		"size":                         tftypes.NewValue(tftypes.Number, 3),
		// 		"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 			"id":        tftypes.String,
		// 			"interface": tftypes.String,
		// 			"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}},
		// 		}}, map[string]tftypes.Value{
		// 			"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
		// 			"interface": tftypes.NewValue(tftypes.String, "scsi"),
		// 			"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}}, nil),
		// 		}),
		// 		"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
		// 	}),
		// 	config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
		// 		"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
		// 		"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
		// 		"size":                         tftypes.NewValue(tftypes.Number, 3),
		// 		"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 			"id":        tftypes.String,
		// 			"interface": tftypes.String,
		// 			"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}},
		// 		}}, map[string]tftypes.Value{
		// 			"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
		// 			"interface": tftypes.NewValue(tftypes.String, "scsi"),
		// 			"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}}, nil),
		// 		}),
		// 		"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
		// 	}),
		// 	proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
		// 		"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
		// 		"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
		// 		"size":                         tftypes.NewValue(tftypes.Number, 3),
		// 		"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 			"id":        tftypes.String,
		// 			"interface": tftypes.String,
		// 			"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}},
		// 		}}, map[string]tftypes.Value{
		// 			"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
		// 			"interface": tftypes.NewValue(tftypes.String, "scsi"),
		// 			"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}}, nil),
		// 		}),
		// 		"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
		// 	}),
		// 	expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
		// 		"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		// 		"name":                         tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
		// 		"size":                         tftypes.NewValue(tftypes.Number, 3),
		// 		"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 			"id":        tftypes.String,
		// 			"interface": tftypes.String,
		// 			"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}},
		// 		}}, map[string]tftypes.Value{
		// 			"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
		// 			"interface": tftypes.NewValue(tftypes.String, "scsi"),
		// 			"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}}, nil),
		// 		}),
		// 		"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
		// 	}),
		// 	expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		// },
		"attr_plan_modifiers_nested_modifier": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "TESTATTRTWO"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "TESTATTRTWO"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:                "test_attribute_plan_modifiers",
			resourceType:            testServeResourceTypeAttributePlanModifiersType,
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := &testServeProvider{
				modifyPlanFunc: tc.modifyPlanFunc,
			}
			testServer := &server{
				p: s,
			}

			priorStateDV, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.priorState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			proposedStateDV, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.proposedNewState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			configDV, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.PlanResourceChangeRequest{
				TypeName:         tc.resource,
				PriorPrivate:     tc.priorPrivate,
				PriorState:       &priorStateDV,
				ProposedNewState: &proposedStateDV,
				Config:           &configDV,
			}
			if tc.providerMeta.Type() != nil {
				providerMeta, err := tfprotov6.NewDynamicValue(testServeProviderMetaType, tc.providerMeta)
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				req.ProviderMeta = &providerMeta
			}
			got, err := testServer.PlanResourceChange(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			gotPlannedState, err := got.PlannedState.Unmarshal(tc.resourceType)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(gotPlannedState, tc.expectedPlannedState); diff != "" {
				t.Errorf("Unexpected diff in planned state (+wanted, -got): %s", diff)
				return
			}
			if string(got.PlannedPrivate) != string(tc.expectedPlannedPrivate) {
				t.Errorf("Expected planned private to be %q, got %q", tc.expectedPlannedPrivate, got.PlannedPrivate)
				return
			}
			if diff := cmp.Diff(got.RequiresReplace, tc.expectedRequiresReplace, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Unexpected diff in requires replace (+wanted, -got): %s", diff)
				return
			}
		})
	}
}

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

		create  func(context.Context, CreateResourceRequest, *CreateResourceResponse)
		update  func(context.Context, UpdateResourceRequest, *UpdateResourceResponse)
		destroy func(context.Context, DeleteResourceRequest, *DeleteResourceResponse)

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
			create: func(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
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
			create: func(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
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
			update: func(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
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
			update: func(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
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
			update: func(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
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
			destroy: func(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, nil)
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
			destroy: func(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, nil)
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
			destroy: func(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
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
			}),
			resource:     "test_two",
			action:       "create",
			resourceType: testServeResourceTypeTwoType,
			create: func(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
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
			}),
			resource:     "test_two",
			action:       "update",
			resourceType: testServeResourceTypeTwoType,
			update: func(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
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
			}),
			resource:     "test_two",
			action:       "delete",
			resourceType: testServeResourceTypeTwoType,
			destroy: func(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
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
			create: func(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
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
			update: func(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
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
			destroy: func(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
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
			}),
			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),
			resource:     "test_two",
			action:       "create",
			resourceType: testServeResourceTypeTwoType,
			create: func(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
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
			}),
			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),
			resource:     "test_two",
			action:       "update",
			resourceType: testServeResourceTypeTwoType,
			update: func(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
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
			}),
			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),
			resource:     "test_two",
			action:       "delete",
			resourceType: testServeResourceTypeTwoType,
			destroy: func(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
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
			testServer := &server{
				p: s,
			}
			var pmSchema Schema
			if tc.providerMeta.Type() != nil {
				sWithMeta := &testServeProviderWithMetaSchema{s}
				testServer.p = sWithMeta
				schema, diags := sWithMeta.GetMetaSchema(context.Background())
				if len(diags) > 0 {
					t.Errorf("Unexpected diags: %+v", diags)
					return
				}
				pmSchema = schema
			}

			rt, diags := testServer.getResourceType(context.Background(), tc.resource)
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

func TestServerValidateDataResourceConfig(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		config         tftypes.Value
		dataSource     string
		dataSourceType tftypes.Type

		impl func(context.Context, ValidateDataSourceConfigRequest, *ValidateDataSourceConfigResponse)

		// response expectations
		expectedDiags []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"no_validation": {
			config: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, nil),
				"current_time": tftypes.NewValue(tftypes.String, nil),
				"is_dst":       tftypes.NewValue(tftypes.Bool, nil),
			}),
			dataSource:     "test_one",
			dataSourceType: testServeDataSourceTypeOneType,
		},
		"config_validators_no_diags": {
			config: tftypes.NewValue(testServeDataSourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_config_validators",
			dataSourceType: testServeDataSourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req ValidateDataSourceConfigRequest, resp *ValidateDataSourceConfigResponse) {},
		},
		"config_validators_one_diag": {
			config: tftypes.NewValue(testServeDataSourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_config_validators",
			dataSourceType: testServeDataSourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req ValidateDataSourceConfigRequest, resp *ValidateDataSourceConfigResponse) {
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
			config: tftypes.NewValue(testServeDataSourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_config_validators",
			dataSourceType: testServeDataSourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req ValidateDataSourceConfigRequest, resp *ValidateDataSourceConfigResponse) {
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
			config: tftypes.NewValue(testServeDataSourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_validate_config",
			dataSourceType: testServeDataSourceTypeValidateConfigType,

			impl: func(_ context.Context, req ValidateDataSourceConfigRequest, resp *ValidateDataSourceConfigResponse) {},
		},
		"validate_config_one_diag": {
			config: tftypes.NewValue(testServeDataSourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_validate_config",
			dataSourceType: testServeDataSourceTypeValidateConfigType,

			impl: func(_ context.Context, req ValidateDataSourceConfigRequest, resp *ValidateDataSourceConfigResponse) {
				resp.Diagnostics.AddError(
					"This is an error",
					"Oops.",
				)
			},

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Summary:  "This is an error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops.",
				},
			},
		},
		"validate_config_two_diags": {
			config: tftypes.NewValue(testServeDataSourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_validate_config",
			dataSourceType: testServeDataSourceTypeValidateConfigType,

			impl: func(_ context.Context, req ValidateDataSourceConfigRequest, resp *ValidateDataSourceConfigResponse) {
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
				validateDataSourceConfigImpl: tc.impl,
			}
			testServer := &server{
				p: s,
			}

			dv, err := tfprotov6.NewDynamicValue(tc.dataSourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ValidateDataResourceConfigRequest{
				TypeName: tc.dataSource,
				Config:   &dv,
			}
			got, err := testServer.ValidateDataResourceConfig(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.validateDataSourceConfigCalledDataSourceType != tc.dataSource && !(tc.dataSource == "test_one" && s.validateDataSourceConfigCalledDataSourceType == "") {
				t.Errorf("Called wrong data source. Expected to call %q, actually called %q", tc.dataSource, s.readDataSourceCalledDataSourceType)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestServerReadDataSource(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		config         tftypes.Value
		providerMeta   tftypes.Value
		dataSource     string
		dataSourceType tftypes.Type

		impl func(context.Context, ReadDataSourceRequest, *ReadDataSourceResponse)

		// response expectations
		expectedNewState tftypes.Value
		expectedDiags    []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"one_basic": {
			config: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, nil),
				"current_time": tftypes.NewValue(tftypes.String, nil),
				"is_dst":       tftypes.NewValue(tftypes.Bool, nil),
			}),
			dataSource:     "test_one",
			dataSourceType: testServeDataSourceTypeOneType,

			impl: func(_ context.Context, req ReadDataSourceRequest, resp *ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
					"current_date": tftypes.NewValue(tftypes.String, "today"),
					"current_time": tftypes.NewValue(tftypes.String, "now"),
					"is_dst":       tftypes.NewValue(tftypes.Bool, true),
				})
			},

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, "today"),
				"current_time": tftypes.NewValue(tftypes.String, "now"),
				"is_dst":       tftypes.NewValue(tftypes.Bool, true),
			}),
		},
		"one_provider_meta": {
			config: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, nil),
				"current_time": tftypes.NewValue(tftypes.String, nil),
				"is_dst":       tftypes.NewValue(tftypes.Bool, nil),
			}),
			dataSource:     "test_one",
			dataSourceType: testServeDataSourceTypeOneType,

			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),

			impl: func(_ context.Context, req ReadDataSourceRequest, resp *ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
					"current_date": tftypes.NewValue(tftypes.String, "today"),
					"current_time": tftypes.NewValue(tftypes.String, "now"),
					"is_dst":       tftypes.NewValue(tftypes.Bool, true),
				})
			},

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, "today"),
				"current_time": tftypes.NewValue(tftypes.String, "now"),
				"is_dst":       tftypes.NewValue(tftypes.Bool, true),
			}),
		},
		"one_remove": {
			config: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, nil),
				"current_time": tftypes.NewValue(tftypes.String, nil),
				"is_dst":       tftypes.NewValue(tftypes.Bool, nil),
			}),
			dataSource:     "test_one",
			dataSourceType: testServeDataSourceTypeOneType,

			impl: func(_ context.Context, req ReadDataSourceRequest, resp *ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeOneType, nil)
			},

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeOneType, nil),
		},
		"two_basic": {
			config: tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
				"family": tftypes.NewValue(tftypes.String, "123foo"),
				"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
				"id":     tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_two",
			dataSourceType: testServeDataSourceTypeTwoType,

			impl: func(_ context.Context, req ReadDataSourceRequest, resp *ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
					"family": tftypes.NewValue(tftypes.String, "123foo"),
					"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
					"id":     tftypes.NewValue(tftypes.String, "a random id or something I dunno"),
				})
			},

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
				"family": tftypes.NewValue(tftypes.String, "123foo"),
				"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
				"id":     tftypes.NewValue(tftypes.String, "a random id or something I dunno"),
			}),
		},
		"two_diags": {
			config: tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
				"family": tftypes.NewValue(tftypes.String, "123foo"),
				"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
				"id":     tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_two",
			dataSourceType: testServeDataSourceTypeTwoType,

			impl: func(_ context.Context, req ReadDataSourceRequest, resp *ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
					"family": tftypes.NewValue(tftypes.String, "123foo"),
					"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
					"id":     tftypes.NewValue(tftypes.String, "a random id or something I dunno"),
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

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
				"family": tftypes.NewValue(tftypes.String, "123foo"),
				"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
				"id":     tftypes.NewValue(tftypes.String, "a random id or something I dunno"),
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
				readDataSourceImpl: tc.impl,
			}
			testServer := &server{
				p: s,
			}
			var pmSchema Schema
			if tc.providerMeta.Type() != nil {
				sWithMeta := &testServeProviderWithMetaSchema{s}
				testServer.p = sWithMeta
				schema, diags := sWithMeta.GetMetaSchema(context.Background())
				if len(diags) > 0 {
					t.Errorf("Unexpected diags: %+v", diags)
					return
				}
				pmSchema = schema
			}

			rt, diags := testServer.getDataSourceType(context.Background(), tc.dataSource)
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}
			schema, diags := rt.GetSchema(context.Background())
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}

			dv, err := tfprotov6.NewDynamicValue(tc.dataSourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ReadDataSourceRequest{
				TypeName: tc.dataSource,
				Config:   &dv,
			}
			if tc.providerMeta.Type() != nil {
				providerMeta, err := tfprotov6.NewDynamicValue(testServeProviderMetaType, tc.providerMeta)
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				req.ProviderMeta = &providerMeta
			}
			got, err := testServer.ReadDataSource(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.readDataSourceCalledDataSourceType != tc.dataSource {
				t.Errorf("Called wrong dataSource. Expected to call %q, actually called %q", tc.dataSource, s.readDataSourceCalledDataSourceType)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			if diff := cmp.Diff(s.readDataSourceConfigValue, tc.config); diff != "" {
				t.Errorf("Unexpected diff in config (+wanted, -got): %s", diff)
				return
			}
			if diff := cmp.Diff(s.readDataSourceConfigSchema, schema); diff != "" {
				t.Errorf("Unexpected diff in config schema (+wanted, -got): %s", diff)
				return
			}
			if tc.providerMeta.Type() != nil {
				if diff := cmp.Diff(s.readDataSourceProviderMetaValue, tc.providerMeta); diff != "" {
					t.Errorf("Unexpected diff in provider meta (+wanted, -got): %s", diff)
					return
				}
				if diff := cmp.Diff(s.readDataSourceProviderMetaSchema, pmSchema); diff != "" {
					t.Errorf("Unexpected diff in provider meta schema (+wanted, -got): %s", diff)
					return
				}
			}
			gotNewState, err := got.State.Unmarshal(tc.dataSourceType)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(gotNewState, tc.expectedNewState); diff != "" {
				t.Errorf("Unexpected diff in new state (+wanted, -got): %s", diff)
				return
			}
		})
	}
}
