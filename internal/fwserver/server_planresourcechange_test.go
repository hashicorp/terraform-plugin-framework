// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMarkComputedNilsAsUnknown(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			// values should be left alone
			"string-value": schema.StringAttribute{
				Required: true,
			},
			// nil, uncomputed values should be left alone
			"string-nil": schema.StringAttribute{
				Optional: true,
			},
			// nil computed values should be turned into unknown
			"string-nil-computed": schema.StringAttribute{
				Computed: true,
			},
			// nil computed values should be turned into unknown
			"string-nil-optional-computed": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			// non-nil computed values should be left alone
			"string-value-optional-computed": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			// values should be left alone
			"dynamic-value": schema.DynamicAttribute{
				Required: true,
			},
			// nil, uncomputed values should be left alone
			"dynamic-nil": schema.DynamicAttribute{
				Optional: true,
			},
			// nil computed values should be turned into unknown
			"dynamic-nil-computed": schema.DynamicAttribute{
				Computed: true,
			},
			// underlying nil computed values should be turned into unknown, this should not preserve the original type
			"dynamic-underlying-string-nil-computed": schema.DynamicAttribute{
				Computed: true,
			},
			// nil computed values should be turned into unknown
			"dynamic-nil-optional-computed": schema.DynamicAttribute{
				Optional: true,
				Computed: true,
			},
			// non-nil computed values should be left alone
			"dynamic-value-optional-computed": schema.DynamicAttribute{
				Optional: true,
				Computed: true,
			},
			// non-nil computed values should be left alone
			// each element of this dynamic value will be visited, then skipped
			"dynamic-value-with-underlying-list-optional-computed": schema.DynamicAttribute{
				Optional: true,
				Computed: true,
			},
			// nil objects should be unknown
			"object-nil-optional-computed": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"string-nil": types.StringType,
					"string-set": types.StringType,
				},
				Optional: true,
				Computed: true,
			},
			// non-nil objects should be left alone
			"object-value-optional-computed": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					// nil attributes of objects
					// should be let alone, as they
					// don't have a schema of their
					// own
					"string-nil": types.StringType,
					"string-set": types.StringType,
				},
				Optional: true,
				Computed: true,
			},
			// nil nested attributes should be unknown
			"nested-nil-optional-computed": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string-nil": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
					"string-set": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
				},
				Optional: true,
				Computed: true,
			},
			// non-nil nested attributes should be left alone on the top level
			"nested-value-optional-computed": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					// nested computed attributes should be unknown
					"string-nil": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
					// nested non-nil computed attributes should be left alone
					"string-set": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
				},
				Optional: true,
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			// nil blocks should remain nil
			"block-nil-optional-computed": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string-nil": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"string-set": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"block-value-optional-computed": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						// nested computed attributes should be unknown
						"string-nil": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						// nested non-nil computed attributes should be left alone
						"string-set": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
	input := tftypes.NewValue(s.Type().TerraformType(context.Background()), map[string]tftypes.Value{
		"string-value":                           tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":                             tftypes.NewValue(tftypes.String, nil),
		"string-nil-computed":                    tftypes.NewValue(tftypes.String, nil),
		"string-nil-optional-computed":           tftypes.NewValue(tftypes.String, nil),
		"string-value-optional-computed":         tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-value":                          tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-nil":                            tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"dynamic-nil-computed":                   tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"dynamic-underlying-string-nil-computed": tftypes.NewValue(tftypes.String, nil),
		"dynamic-nil-optional-computed":          tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"dynamic-value-optional-computed":        tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-value-with-underlying-list-optional-computed": tftypes.NewValue(
			tftypes.List{
				ElementType: tftypes.Bool,
			},
			[]tftypes.Value{
				tftypes.NewValue(tftypes.Bool, true),
				tftypes.NewValue(tftypes.Bool, false),
			},
		),
		"object-nil-optional-computed": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, nil),
		"object-value-optional-computed": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "foo"),
		}),
		"nested-nil-optional-computed": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, nil),
		"nested-value-optional-computed": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "bar"),
		}),
		"block-nil-optional-computed": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			},
		}, nil),
		"block-value-optional-computed": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"string-nil": tftypes.NewValue(tftypes.String, nil),
				"string-set": tftypes.NewValue(tftypes.String, "bar"),
			}),
		}),
	})
	expected := tftypes.NewValue(s.Type().TerraformType(context.Background()), map[string]tftypes.Value{
		"string-value":                           tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":                             tftypes.NewValue(tftypes.String, nil),
		"string-nil-computed":                    tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"string-nil-optional-computed":           tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"string-value-optional-computed":         tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-value":                          tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-nil":                            tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"dynamic-nil-computed":                   tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
		"dynamic-underlying-string-nil-computed": tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue), // doesn't preserve original type
		"dynamic-nil-optional-computed":          tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
		"dynamic-value-optional-computed":        tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-value-with-underlying-list-optional-computed": tftypes.NewValue(
			tftypes.List{
				ElementType: tftypes.Bool,
			},
			[]tftypes.Value{
				tftypes.NewValue(tftypes.Bool, true),
				tftypes.NewValue(tftypes.Bool, false),
			},
		),
		"object-nil-optional-computed": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, tftypes.UnknownValue),
		"object-value-optional-computed": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "foo"),
		}),
		"nested-nil-optional-computed": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, tftypes.UnknownValue),
		"nested-value-optional-computed": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"string-set": tftypes.NewValue(tftypes.String, "bar"),
		}),
		"block-nil-optional-computed": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			},
		}, nil),
		"block-value-optional-computed": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"string-nil": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"string-set": tftypes.NewValue(tftypes.String, "bar"),
			}),
		}),
	})

	got, err := tftypes.Transform(input, fwserver.MarkComputedNilsAsUnknown(context.Background(), input, s))
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
		input    path.Paths
		expected path.Paths
	}

	tests := map[string]testCase{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"no-duplicates": {
			input: path.Paths{
				path.Root("name2"),
				path.Root("name1"),
				path.Empty().AtListIndex(1234),
				path.Root("name1").AtMapKey("elementkey"),
			},
			expected: path.Paths{
				path.Empty().AtListIndex(1234),
				path.Root("name1"),
				path.Root("name1").AtMapKey("elementkey"),
				path.Root("name2"),
			},
		},
		"duplicates": {
			input: path.Paths{
				path.Root("name1"),
				path.Root("name1"),
				path.Empty().AtListIndex(1234),
				path.Empty().AtListIndex(1234),
			},
			expected: path.Paths{
				path.Empty().AtListIndex(1234),
				path.Root("name1"),
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := fwserver.NormaliseRequiresReplace(context.Background(), tc.input)

			if diff := cmp.Diff(actual, tc.expected, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
				return
			}
		})
	}
}

func TestServerPlanResourceChange(t *testing.T) {
	t.Parallel()

	testSchemaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testSchemaTypeDefault := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed_bool":                    tftypes.Bool,
			"test_computed_float32":                 tftypes.Number,
			"test_computed_float64":                 tftypes.Number,
			"test_computed_int32":                   tftypes.Number,
			"test_computed_int64":                   tftypes.Number,
			"test_computed_list":                    tftypes.List{ElementType: tftypes.String},
			"test_computed_map":                     tftypes.Map{ElementType: tftypes.String},
			"test_computed_number":                  tftypes.Number,
			"test_computed_object":                  tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
			"test_computed_set":                     tftypes.Set{ElementType: tftypes.String},
			"test_computed_string":                  tftypes.String,
			"test_computed_string_custom_type":      tftypes.String,
			"test_computed_dynamic":                 tftypes.DynamicPseudoType,
			"test_computed_nested_list":             tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_list_attribute":   tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_map":              tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_map_attribute":    tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_set":              tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_set_attribute":    tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_single":           tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}},
			"test_computed_nested_single_attribute": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}},
			"test_configured_bool":                  tftypes.Bool,
			"test_configured_float32":               tftypes.Number,
			"test_configured_float64":               tftypes.Number,
			"test_configured_int32":                 tftypes.Number,
			"test_configured_int64":                 tftypes.Number,
			"test_configured_list":                  tftypes.List{ElementType: tftypes.String},
			"test_configured_map":                   tftypes.Map{ElementType: tftypes.String},
			"test_configured_number":                tftypes.Number,
			"test_configured_object":                tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
			"test_configured_set":                   tftypes.Set{ElementType: tftypes.String},
			"test_configured_string":                tftypes.String,
			"test_configured_string_custom_type":    tftypes.String,
			"test_configured_dynamic":               tftypes.DynamicPseudoType,
			"test_configured_nested_list": tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"computed_attribute":     tftypes.String,
						"configurable_attribute": tftypes.String,
					},
				},
			},
			"test_configured_nested_list_attribute": tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"computed_attribute":     tftypes.String,
						"configurable_attribute": tftypes.String,
					},
				},
			},
			"test_configured_nested_map": tftypes.Map{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"computed_attribute":     tftypes.String,
						"configurable_attribute": tftypes.String,
					},
				},
			},
			"test_configured_nested_map_attribute": tftypes.Map{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"computed_attribute":     tftypes.String,
						"configurable_attribute": tftypes.String,
					},
				},
			},
			"test_configured_nested_set": tftypes.Set{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"computed_attribute":     tftypes.String,
						"configurable_attribute": tftypes.String,
					},
				},
			},
			"test_configured_nested_set_attribute": tftypes.Set{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"computed_attribute":     tftypes.String,
						"configurable_attribute": tftypes.String,
					},
				},
			},
			"test_configured_nested_single": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"computed_attribute":     tftypes.String,
					"configurable_attribute": tftypes.String,
				},
			},
			"test_configured_nested_single_attribute": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"computed_attribute":     tftypes.String,
					"configurable_attribute": tftypes.String,
				},
			},
		},
	}

	testSchemaBlockType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_required": tftypes.String,
			"test_optional_block": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test_optional_one": tftypes.String,
					"test_optional_two": tftypes.String,
				},
			},
		},
	}

	testSchemaTypeComputedRequired := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

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

	testSchemaDefault := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed_bool": schema.BoolAttribute{
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"test_computed_float32": schema.Float32Attribute{
				Computed: true,
				Default:  float32default.StaticFloat32(1.2345),
			},
			"test_computed_float64": schema.Float64Attribute{
				Computed: true,
				Default:  float64default.StaticFloat64(1.2345),
			},
			"test_computed_int32": schema.Int32Attribute{
				Computed: true,
				Default:  int32default.StaticInt32(12345),
			},
			"test_computed_int64": schema.Int64Attribute{
				Computed: true,
				Default:  int64default.StaticInt64(12345),
			},
			"test_computed_list": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default")})),
			},
			"test_computed_map": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Default:     mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{"a": types.StringValue("default")})),
			},
			"test_computed_number": schema.NumberAttribute{
				Computed: true,
				Default:  numberdefault.StaticBigFloat(big.NewFloat(1.2345)),
			},
			"test_computed_object": schema.ObjectAttribute{
				Computed:       true,
				AttributeTypes: map[string]attr.Type{"a": types.StringType},
				Default:        objectdefault.StaticValue(types.ObjectValueMust(map[string]attr.Type{"a": types.StringType}, map[string]attr.Value{"a": types.StringValue("default")})),
			},
			"test_computed_set": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("default")})),
			},
			"test_computed_string": schema.StringAttribute{
				Computed: true,
				Default:  stringdefault.StaticString("one"),
			},
			"test_computed_string_custom_type": schema.StringAttribute{
				Computed:   true,
				CustomType: testtypes.StringTypeWithSemanticEquals{},
				Default:    stringdefault.StaticString("one"),
			},
			"test_computed_dynamic": schema.DynamicAttribute{
				Computed: true,
				Default:  dynamicdefault.StaticValue(types.DynamicValue(types.StringValue("hello world!"))),
			},
			"test_computed_nested_list": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"string_attribute": types.StringType,
					},
				},
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"string_attribute": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"string_attribute": types.StringType,
								},
								map[string]attr.Value{
									"string_attribute": types.StringValue("default"),
								}),
						},
					),
				),
			},
			"test_computed_nested_list_attribute": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							Optional: true,
							Default:  stringdefault.StaticString("default-attribute"),
						},
					},
				},
			},
			"test_computed_nested_map": schema.MapAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"string_attribute": types.StringType,
					},
				},
				Default: mapdefault.StaticValue(
					types.MapValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"string_attribute": types.StringType,
							},
						},
						map[string]attr.Value{
							"test-key": types.ObjectValueMust(
								map[string]attr.Type{
									"string_attribute": types.StringType,
								},
								map[string]attr.Value{
									"string_attribute": types.StringValue("default"),
								}),
						},
					),
				),
			},
			"test_computed_nested_map_attribute": schema.MapNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							Optional: true,
							Default:  stringdefault.StaticString("default-attribute"),
						},
					},
				},
			},
			"test_computed_nested_set": schema.SetAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"string_attribute": types.StringType,
					},
				},
				Default: setdefault.StaticValue(
					types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"string_attribute": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"string_attribute": types.StringType,
								},
								map[string]attr.Value{
									"string_attribute": types.StringValue("default"),
								}),
						},
					),
				),
			},
			"test_computed_nested_set_attribute": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							Optional: true,
							Default:  stringdefault.StaticString("default-attribute"),
						},
					},
				},
			},
			"test_computed_nested_single": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Optional: true,
					},
				},
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						map[string]attr.Value{
							"string_attribute": types.StringValue("default"),
						}),
				),
			},
			"test_computed_nested_single_attribute": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Optional: true,
						Default:  stringdefault.StaticString("default-attribute"),
					},
				},
			},
			"test_configured_bool": schema.BoolAttribute{
				Optional: true,
				Default:  booldefault.StaticBool(true),
			},
			"test_configured_float32": schema.Float32Attribute{
				Optional: true,
				Computed: true,
				Default:  float32default.StaticFloat32(1.2345),
			},
			"test_configured_float64": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(1.2345),
			},
			"test_configured_int32": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				Default:  int32default.StaticInt32(12345),
			},
			"test_configured_int64": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(12345),
			},
			"test_configured_list": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default")})),
			},
			"test_configured_map": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{"a": types.StringValue("default")})),
			},
			"test_configured_number": schema.NumberAttribute{
				Optional: true,
				Computed: true,
				Default:  numberdefault.StaticBigFloat(big.NewFloat(1.2345)),
			},
			"test_configured_object": schema.ObjectAttribute{
				Optional:       true,
				Computed:       true,
				AttributeTypes: map[string]attr.Type{"a": types.StringType},
				Default:        objectdefault.StaticValue(types.ObjectValueMust(map[string]attr.Type{"a": types.StringType}, map[string]attr.Value{"a": types.StringValue("default")})),
			},
			"test_configured_set": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("default")})),
			},
			"test_configured_string": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("one"),
			},
			"test_configured_string_custom_type": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				CustomType: testtypes.StringTypeWithSemanticEquals{},
				Default:    stringdefault.StaticString("one"),
			},
			"test_configured_dynamic": schema.DynamicAttribute{
				Optional: true,
				Computed: true,
				Default:  dynamicdefault.StaticValue(types.DynamicValue(types.StringValue("hello world!"))),
			},
			"test_configured_nested_list": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"computed_attribute": schema.StringAttribute{
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
						"configurable_attribute": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
					},
				},
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"computed_attribute":     types.StringType,
								"configurable_attribute": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"computed_attribute":     types.StringType,
									"configurable_attribute": types.StringType,
								},
								map[string]attr.Value{
									"computed_attribute":     types.StringValue("list-default-value"),
									"configurable_attribute": types.StringValue("list-default-value"),
								}),
						},
					),
				),
			},
			"test_configured_nested_list_attribute": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"computed_attribute": schema.StringAttribute{
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
						"configurable_attribute": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
					},
				},
			},
			"test_configured_nested_map": schema.MapNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"computed_attribute": schema.StringAttribute{
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
						"configurable_attribute": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
					},
				},
				Default: mapdefault.StaticValue(
					types.MapValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"computed_attribute":     types.StringType,
								"configurable_attribute": types.StringType,
							},
						},
						map[string]attr.Value{
							"test-key": types.ObjectValueMust(
								map[string]attr.Type{
									"computed_attribute":     types.StringType,
									"configurable_attribute": types.StringType,
								},
								map[string]attr.Value{
									"computed_attribute":     types.StringValue("map-default-value"),
									"configurable_attribute": types.StringValue("map-default-value"),
								}),
						},
					),
				),
			},
			"test_configured_nested_map_attribute": schema.MapNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"computed_attribute": schema.StringAttribute{
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
						"configurable_attribute": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
					},
				},
			},
			"test_configured_nested_set": schema.SetNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"computed_attribute": schema.StringAttribute{
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
						"configurable_attribute": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
					},
				},
				Default: setdefault.StaticValue(
					types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"computed_attribute":     types.StringType,
								"configurable_attribute": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"computed_attribute":     types.StringType,
									"configurable_attribute": types.StringType,
								},
								map[string]attr.Value{
									"computed_attribute":     types.StringValue("set-default-value"),
									"configurable_attribute": types.StringValue("set-default-value"),
								}),
						},
					),
				),
			},
			"test_configured_nested_set_attribute": schema.SetNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"computed_attribute": schema.StringAttribute{
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
						"configurable_attribute": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("attribute-default-value"),
						},
					},
				},
			},
			"test_configured_nested_single": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"computed_attribute": schema.StringAttribute{
						Computed: true,
						Default:  stringdefault.StaticString("attribute-default-value"),
					},
					"configurable_attribute": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("attribute-default-value"),
					},
				},
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"computed_attribute":     types.StringType,
							"configurable_attribute": types.StringType,
						},
						map[string]attr.Value{
							"computed_attribute":     types.StringValue("object-default-value"),
							"configurable_attribute": types.StringValue("object-default-value"),
						}),
				),
			},
			"test_configured_nested_single_attribute": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"computed_attribute": schema.StringAttribute{
						Computed: true,
						Default:  stringdefault.StaticString("attribute-default-value"),
					},
					"configurable_attribute": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("attribute-default-value"),
					},
				},
			},
		},
	}

	testSchemaBlock := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"test_optional_block": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"test_optional_one": schema.StringAttribute{
						Optional: true,
					},
					"test_optional_two": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}

	testEmptyPlan := &tfsdk.Plan{
		Raw:    tftypes.NewValue(testSchemaType, nil),
		Schema: testSchema,
	}

	testEmptyState := &tfsdk.State{
		Raw:    tftypes.NewValue(testSchemaType, nil),
		Schema: testSchema,
	}

	testEmptyStateDefault := &tfsdk.State{
		Raw:    tftypes.NewValue(testSchemaTypeDefault, nil),
		Schema: testSchemaDefault,
	}

	type testSchemaData struct {
		TestComputed types.String `tfsdk:"test_computed"`
		TestRequired types.String `tfsdk:"test_required"`
	}

	type testSchemaDataBlock struct {
		TestRequired      types.String `tfsdk:"test_required"`
		TestOptionalBlock types.Object `tfsdk:"test_optional_block"`
	}

	testSchemaAttributePlanModifierAttributePlan := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.PlanValue = types.StringValue("test-attributeplanmodifier-value")
						},
					},
				},
			},
			"test_other_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testSchemaAttributePlanModifierAttributePlanCustomType := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed:   true,
				CustomType: testtypes.StringTypeWithSemanticEquals{},
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.PlanValue = types.StringValue("test-attributeplanmodifier-value")
						},
					},
				},
			},
			"test_other_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testSchemaAttributePlanModifierPrivatePlanRequest := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

							key := "providerKeyOne"
							got, diags := req.Private.GetKey(ctx, key)

							resp.Diagnostics.Append(diags...)

							if string(got) != expected {
								resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
							}
						},
					},
				},
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testSchemaAttributePlanModifierPrivatePlanResponse := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

							resp.Diagnostics.Append(diags...)
						},
					},
				},
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testSchemaAttributePlanModifierDiagnosticsError := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.Diagnostics.AddAttributeError(req.Path, "error summary", "error detail")
						},
					},
				},
			},
		},
	}

	testSchemaAttributePlanModifierRequiresReplace := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
		},
	}

	testProviderMetaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_provider_meta_attribute": tftypes.String,
		},
	}

	testProviderMetaValue := tftypes.NewValue(testProviderMetaType, map[string]tftypes.Value{
		"test_provider_meta_attribute": tftypes.NewValue(tftypes.String, "test-provider-meta-value"),
	})

	testProviderMetaSchema := metaschema.Schema{
		Attributes: map[string]metaschema.Attribute{
			"test_provider_meta_attribute": metaschema.StringAttribute{
				Optional: true,
			},
		},
	}

	testProviderMetaConfig := &tfsdk.Config{
		Raw:    testProviderMetaValue,
		Schema: testProviderMetaSchema,
	}

	type testProviderMetaData struct {
		TestProviderMetaAttribute types.String `tfsdk:"test_provider_meta_attribute"`
	}

	testPrivateFrameworkMap := map[string][]byte{
		".frameworkKey": []byte(`{"fk": "framework value"}`),
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testPrivateProvider := &privatestate.Data{
		Provider: testProviderData,
	}

	testPrivate := &privatestate.Data{
		Framework: testPrivateFrameworkMap,
		Provider:  testProviderData,
	}

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testEmptyPrivate := &privatestate.Data{
		Provider: testEmptyProviderData,
	}

	testDeferralAllowed := resource.ModifyPlanClientCapabilities{
		DeferralAllowed: true,
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.PlanResourceChangeRequest
		expectedResponse     *fwserver.PlanResourceChangeResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"resource-configure-data": {
			server: &fwserver.Server{
				Provider:              &testprovider.Provider{},
				ResourceConfigureData: "test-provider-configure-value",
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithConfigureAndModifyPlan{
					ConfigureMethod: func(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
						providerData, ok := req.ProviderData.(string)

						if !ok {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.ProviderData",
								fmt.Sprintf("Expected string, got: %T", req.ProviderData),
							)
							return
						}

						if providerData != "test-provider-configure-value" {
							resp.Diagnostics.AddError(
								"Unexpected ConfigureRequest.ProviderData",
								fmt.Sprintf("Expected test-provider-configure-value, got: %q", providerData),
							)
						}
					},
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						// In practice, the Configure method would save the
						// provider data to the Resource implementation and
						// use it here. The fact that Configure is able to
						// read the data proves this can work.
					},
					Resource: &testprovider.Resource{},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-mark-computed-config-nils-as-unknown": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-set-default-values": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":               tftypes.NewValue(tftypes.Bool, nil),
						"test_computed_float32":            tftypes.NewValue(tftypes.Number, nil),
						"test_computed_float64":            tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int32":              tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int64":              tftypes.NewValue(tftypes.Number, nil),
						"test_computed_list":               tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_computed_map":                tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"test_computed_number":             tftypes.NewValue(tftypes.Number, nil),
						"test_computed_object":             tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, nil),
						"test_computed_set":                tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
						"test_computed_string":             tftypes.NewValue(tftypes.String, nil),
						"test_computed_string_custom_type": tftypes.NewValue(tftypes.String, nil),
						"test_computed_dynamic":            tftypes.NewValue(tftypes.DynamicPseudoType, nil),
						"test_computed_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							nil,
						),
						"test_computed_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, nil),
							},
						),
						"test_configured_bool":    tftypes.NewValue(tftypes.Bool, true),
						"test_configured_float32": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_float64": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_int32":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_int64":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_list": tftypes.NewValue(
							tftypes.List{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_map": tftypes.NewValue(
							tftypes.Map{ElementType: tftypes.String},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_number": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_object": tftypes.NewValue(
							tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
							map[string]tftypes.Value{
								"a": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_set": tftypes.NewValue(
							tftypes.Set{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_string":             tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_string_custom_type": tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_dynamic":            tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
					}),
					Schema: testSchemaDefault,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":               tftypes.NewValue(tftypes.Bool, nil),
						"test_computed_float32":            tftypes.NewValue(tftypes.Number, nil),
						"test_computed_float64":            tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int32":              tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int64":              tftypes.NewValue(tftypes.Number, nil),
						"test_computed_list":               tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_computed_map":                tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"test_computed_number":             tftypes.NewValue(tftypes.Number, nil),
						"test_computed_object":             tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, nil),
						"test_computed_set":                tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
						"test_computed_string":             tftypes.NewValue(tftypes.String, nil),
						"test_computed_string_custom_type": tftypes.NewValue(tftypes.String, nil),
						"test_computed_dynamic":            tftypes.NewValue(tftypes.DynamicPseudoType, nil),
						"test_computed_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							nil,
						),
						"test_computed_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, nil),
							},
						),
						"test_configured_bool":    tftypes.NewValue(tftypes.Bool, true),
						"test_configured_float32": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_float64": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_int32":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_int64":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_list": tftypes.NewValue(
							tftypes.List{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_map": tftypes.NewValue(
							tftypes.Map{ElementType: tftypes.String},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_number": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_object": tftypes.NewValue(
							tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
							map[string]tftypes.Value{
								"a": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_set": tftypes.NewValue(
							tftypes.Set{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_string":             tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_string_custom_type": tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_dynamic":            tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
					}),
					Schema: testSchemaDefault,
				},
				PriorState:     testEmptyStateDefault,
				ResourceSchema: testSchemaDefault,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":               tftypes.NewValue(tftypes.Bool, true),
						"test_computed_float32":            tftypes.NewValue(tftypes.Number, 1.2345),
						"test_computed_float64":            tftypes.NewValue(tftypes.Number, 1.2345),
						"test_computed_int32":              tftypes.NewValue(tftypes.Number, 12345),
						"test_computed_int64":              tftypes.NewValue(tftypes.Number, 12345),
						"test_computed_list":               tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_map":                tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_number":             tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
						"test_computed_object":             tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_set":                tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_string":             tftypes.NewValue(tftypes.String, "one"),
						"test_computed_string_custom_type": tftypes.NewValue(tftypes.String, "one"),
						"test_computed_dynamic":            tftypes.NewValue(tftypes.String, "hello world!"),
						"test_computed_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default"),
									},
								),
							},
						),
						"test_computed_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default-attribute"),
									},
								),
							},
						),
						"test_computed_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default"),
									},
								),
							},
						),
						"test_computed_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default-attribute"),
									},
								),
							},
						),
						"test_computed_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default"),
									},
								),
							},
						),
						"test_computed_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default-attribute"),
									},
								),
							},
						),
						"test_computed_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "default"),
							},
						),
						"test_computed_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "default-attribute"),
							},
						),
						"test_configured_bool":    tftypes.NewValue(tftypes.Bool, true),
						"test_configured_float32": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_float64": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_int32":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_int64":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_list": tftypes.NewValue(
							tftypes.List{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_map": tftypes.NewValue(
							tftypes.Map{ElementType: tftypes.String},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_number": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_object": tftypes.NewValue(
							tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
							map[string]tftypes.Value{
								"a": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_set": tftypes.NewValue(
							tftypes.Set{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_string":             tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_string_custom_type": tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_dynamic":            tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
					}),
					Schema: testSchemaDefault,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-attributeplanmodifier-request-privateplan": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanRequest,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanRequest,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchemaAttributePlanModifierPrivatePlanRequest,
				Resource:       &testprovider.Resource{},
				PriorPrivate:   testPrivate,
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanRequest,
				},
				PlannedPrivate: testPrivate,
			},
		},
		"create-attributeplanmodifier-response-attributeplan": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, nil),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				ResourceSchema: testSchemaAttributePlanModifierAttributePlan,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, "test-attributeplanmodifier-value"),
						"test_other_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required":       tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-attributeplanmodifier-response-attributeplan-custom-type": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlanCustomType,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlanCustomType,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, nil),
					Schema: testSchemaAttributePlanModifierAttributePlanCustomType,
				},
				ResourceSchema: testSchemaAttributePlanModifierAttributePlanCustomType,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, "test-attributeplanmodifier-value"),
						"test_other_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required":       tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlanCustomType,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-attributeplanmodifier-response-privateplan": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchemaAttributePlanModifierPrivatePlanResponse,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				PlannedPrivate: testPrivateProvider,
			},
		},
		"create-attributeplanmodifier-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierDiagnosticsError,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierDiagnosticsError,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchemaAttributePlanModifierDiagnosticsError,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.WithPath(
						path.Root("test_required"),
						diag.NewErrorDiagnostic("error summary", "error detail"),
					),
				},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierDiagnosticsError,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-attributeplanmodifier-response-requiresreplace": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierRequiresReplace,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierRequiresReplace,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchemaAttributePlanModifierRequiresReplace,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierRequiresReplace,
				},
				// This is a strange thing to signal on creation, but the
				// framework does not prevent you from doing it and it might
				// be overly burdensome on provider developers to have the
				// framework raise an error if it is technically valid in the
				// protocol.
				RequiresReplace: path.Paths{
					path.Root("test_required"),
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-request-client-capabilities-deferral-allowed": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				ClientCapabilities: testDeferralAllowed,
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						if req.ClientCapabilities.DeferralAllowed != true {
							resp.Diagnostics.AddError("Unexpected req.ClientCapabilities.DeferralAllowed value",
								"expected: true but got: false")
						}

					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

						key := "providerKeyOne"
						got, diags := req.Private.GetKey(ctx, key)

						resp.Diagnostics.Append(diags...)

						if string(got) != expected {
							resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
						}
					},
				},
				PriorPrivate: testPrivate,
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testPrivate,
			},
		},
		"create-resourcewithmodifyplan-request-proposednewstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						if !data.TestComputed.IsUnknown() {
							resp.Diagnostics.AddError("Unexpected req.Plan Value", "Got: "+data.TestComputed.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ProviderMeta:   testProviderMetaConfig,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testProviderMetaData

						resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

						if data.TestProviderMetaAttribute.ValueString() != "test-provider-meta-value" {
							resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+data.TestProviderMetaAttribute.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-response-deferral-automatic-override-provider-deferral-reason": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
						resp.Deferred = &provider.Deferred{Reason: provider.DeferredReasonProviderConfigUnknown}
					},
				},
			},
			configureProviderReq: &provider.ConfigureRequest{
				ClientCapabilities: provider.ConfigureProviderClientCapabilities{
					DeferralAllowed: true,
				},
			},
			request: &fwserver.PlanResourceChangeRequest{
				ClientCapabilities: testDeferralAllowed,
				ResourceBehavior: resource.ResourceBehavior{
					ProviderDeferred: resource.ProviderDeferredBehavior{
						EnablePlanModification: true,
					},
				},
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						if req.ClientCapabilities.DeferralAllowed == true {
							resp.Deferred = &resource.Deferred{Reason: resource.DeferredReasonAbsentPrereq}
						}

						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						data.TestComputed = types.StringValue("test-plannedstate-value")

						resp.Diagnostics.Append(resp.Plan.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Deferred: &resource.Deferred{Reason: resource.DeferredReasonAbsentPrereq},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-response-deferral-automatic-plan-modification": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
						resp.Deferred = &provider.Deferred{Reason: provider.DeferredReasonProviderConfigUnknown}
					},
				},
			},
			configureProviderReq: &provider.ConfigureRequest{
				ClientCapabilities: provider.ConfigureProviderClientCapabilities{
					DeferralAllowed: true,
				},
			},
			request: &fwserver.PlanResourceChangeRequest{
				ResourceBehavior: resource.ResourceBehavior{
					ProviderDeferred: resource.ProviderDeferredBehavior{
						EnablePlanModification: true,
					},
				},
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						data.TestComputed = types.StringValue("test-plannedstate-value")

						resp.Diagnostics.Append(resp.Plan.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Deferred: &resource.Deferred{Reason: resource.DeferredReasonProviderConfigUnknown},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-response-deferral-automatic-skip-plan-modification": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{
					SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {},
					ConfigureMethod: func(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
						resp.Deferred = &provider.Deferred{Reason: provider.DeferredReasonProviderConfigUnknown}
					},
				},
			},
			configureProviderReq: &provider.ConfigureRequest{
				ClientCapabilities: provider.ConfigureProviderClientCapabilities{
					DeferralAllowed: true,
				},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						resp.Diagnostics.AddError("Test assertion failed: ", "modifyplan shouldn't be called")
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Deferred: &resource.Deferred{Reason: resource.DeferredReasonProviderConfigUnknown},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
			},
		},
		"create-resourcewithmodifyplan-response-deferral-manual": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				ClientCapabilities: testDeferralAllowed,
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						if req.ClientCapabilities.DeferralAllowed == true {
							resp.Deferred = &resource.Deferred{Reason: resource.DeferredReasonAbsentPrereq}
						}

					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Deferred: &resource.Deferred{Reason: resource.DeferredReasonAbsentPrereq},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("warning summary", "warning detail"),
					diag.NewErrorDiagnostic("error summary", "error detail"),
				},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-response-plannedstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						data.TestComputed = types.StringValue("test-plannedstate-value")

						resp.Diagnostics.Append(resp.Plan.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testPrivateProvider,
			},
		},
		"create-resourcewithmodifyplan-response-requiresreplace": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						// This is a strange thing to signal on creation,
						// but the framework does not prevent you from
						// doing it and it might be overly burdensome on
						// provider developers to have the framework raise
						// an error if it is technically valid in the
						// protocol.
						resp.RequiresReplace = path.Paths{
							path.Root("test_required"),
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				RequiresReplace: path.Paths{
					path.Root("test_required"),
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"create-resourcewithmodifyplan-attributeplanmodifier-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaTypeComputedRequired, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaTypeComputedRequired, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				PriorState:     testEmptyState,
				ResourceSchema: testSchemaAttributePlanModifierPrivatePlanResponse,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

						key := "providerKeyOne"
						got, diags := req.Private.GetKey(ctx, key)

						resp.Diagnostics.Append(diags...)

						if string(got) != expected {
							resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeComputedRequired, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				PlannedPrivate: testPrivateProvider,
			},
		},
		"delete-resourcewithmodifyplan-request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-state-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-config-value" {
							resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState:   testEmptyState,
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"delete-resourcewithmodifyplan-request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-state-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

						key := "providerKeyOne"
						got, diags := req.Private.GetKey(ctx, key)

						resp.Diagnostics.Append(diags...)

						if string(got) != expected {
							resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
						}
					},
				},
				PriorPrivate: testPrivateProvider,
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState:   testEmptyState,
				PlannedPrivate: testPrivateProvider,
			},
		},
		"delete-resourcewithmodifyplan-request-priorstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-state-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-state-value" {
							resp.Diagnostics.AddError("Unexpected req.State Value", "Got: "+data.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState:   testEmptyState,
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"delete-resourcewithmodifyplan-request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-state-value"),
					}),
					Schema: testSchema,
				},
				ProviderMeta:   testProviderMetaConfig,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testProviderMetaData

						resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

						if data.TestProviderMetaAttribute.ValueString() != "test-provider-meta-value" {
							resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+data.TestProviderMetaAttribute.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState:   testEmptyState,
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"delete-resourcewithmodifyplan-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-state-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("warning summary", "warning detail"),
					diag.NewErrorDiagnostic("error summary", "error detail"),
				},
				PlannedState:   testEmptyState,
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"delete-resourcewithmodifyplan-response-plannedstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-state-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						// This is invalid logic to run during deletion.
						resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("test_computed"), types.StringValue("test-plannedstate-value"))...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Planned Resource State on Destroy",
						"The Terraform Provider unexpectedly returned resource state data when the resource was planned for destruction. "+
							"This is always an issue in the Terraform Provider and should be reported to the provider developers.\n\n"+
							"Ensure all resource plan modifiers do not attempt to change resource plan data from being a null value if the request plan is a null value.",
					),
				},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"delete-resourcewithmodifyplan-response-requiresreplace": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-state-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						// This is a strange thing to signal on creation,
						// but the framework does not prevent you from
						// doing it and it might be overly burdensome on
						// provider developers to have the framework raise
						// an error if it is technically valid in the
						// protocol.
						resp.RequiresReplace = path.Paths{
							path.Root("test_required"),
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: testEmptyState,
				RequiresReplace: path.Paths{
					path.Root("test_required"),
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"delete-resourcewithmodifyplan-response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-state-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState:   testEmptyState,
				PlannedPrivate: testPrivateProvider,
			},
		},
		"delete-resourcewithmodifyplan-attributeplanmodifier-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaTypeComputedRequired, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				ProposedNewState: testEmptyPlan,
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeComputedRequired, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-state-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState:   testEmptyState,
				PlannedPrivate: testPrivateProvider,
			},
		},
		"update-mark-computed-config-nils-as-unknown": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-set-default-values": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":               tftypes.NewValue(tftypes.Bool, nil),
						"test_computed_float32":            tftypes.NewValue(tftypes.Number, nil),
						"test_computed_float64":            tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int32":              tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int64":              tftypes.NewValue(tftypes.Number, nil),
						"test_computed_list":               tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_computed_map":                tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"test_computed_number":             tftypes.NewValue(tftypes.Number, nil),
						"test_computed_object":             tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, nil),
						"test_computed_set":                tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
						"test_computed_string":             tftypes.NewValue(tftypes.String, nil),
						"test_computed_string_custom_type": tftypes.NewValue(tftypes.String, nil),
						"test_computed_dynamic":            tftypes.NewValue(tftypes.DynamicPseudoType, nil),
						"test_computed_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							nil,
						),
						"test_computed_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, nil),
							},
						),
						"test_configured_bool":    tftypes.NewValue(tftypes.Bool, true),
						"test_configured_float32": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_float64": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_int32":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_int64":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_list": tftypes.NewValue(
							tftypes.List{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_map": tftypes.NewValue(
							tftypes.Map{ElementType: tftypes.String},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_number": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_object": tftypes.NewValue(
							tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
							map[string]tftypes.Value{
								"a": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_set": tftypes.NewValue(
							tftypes.Set{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_string":             tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_string_custom_type": tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_dynamic":            tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								// TODO: Prevent random attributes map access from causing flaky test
								// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/783
								// tftypes.NewValue(
								// 	tftypes.Object{
								// 		AttributeTypes: map[string]tftypes.Type{
								// 			"computed_attribute":     tftypes.String,
								// 			"configurable_attribute": tftypes.String,
								// 		},
								// 	},
								// 	map[string]tftypes.Value{
								// 		"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
								// 		"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
								// 	},
								// ),
							},
						),
						"test_configured_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, nil),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
					}),
					Schema: testSchemaDefault,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":               tftypes.NewValue(tftypes.Bool, nil),
						"test_computed_float32":            tftypes.NewValue(tftypes.Number, nil),
						"test_computed_float64":            tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int32":              tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int64":              tftypes.NewValue(tftypes.Number, nil),
						"test_computed_list":               tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_computed_map":                tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"test_computed_number":             tftypes.NewValue(tftypes.Number, nil),
						"test_computed_object":             tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, nil),
						"test_computed_set":                tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
						"test_computed_string":             tftypes.NewValue(tftypes.String, nil),
						"test_computed_string_custom_type": tftypes.NewValue(tftypes.String, nil),
						"test_computed_dynamic":            tftypes.NewValue(tftypes.DynamicPseudoType, nil),
						"test_computed_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							nil,
						),
						"test_computed_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, nil),
									},
								),
							},
						),
						"test_computed_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							nil,
						),
						"test_computed_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, nil),
							},
						),
						"test_configured_bool":    tftypes.NewValue(tftypes.Bool, true),
						"test_configured_float32": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_float64": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_int64":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_int32":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_list": tftypes.NewValue(
							tftypes.List{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_map": tftypes.NewValue(
							tftypes.Map{ElementType: tftypes.String},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_number": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_object": tftypes.NewValue(
							tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
							map[string]tftypes.Value{
								"a": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_set": tftypes.NewValue(
							tftypes.Set{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_string":             tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_string_custom_type": tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_dynamic":            tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-1"), // Terraform core behavior
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-2"), // Terraform core behavior
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-1"), // Terraform core behavior
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-2"), // Terraform core behavior
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value"), // Terraform core behavior
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value"), // Terraform core behavior
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-1"), // Terraform core behavior
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								// TODO: Prevent random attributes map access from causing flaky test
								// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/783
								// tftypes.NewValue(
								// 	tftypes.Object{
								// 		AttributeTypes: map[string]tftypes.Type{
								// 			"computed_attribute":     tftypes.String,
								// 			"configurable_attribute": tftypes.String,
								// 		},
								// 	},
								// 	map[string]tftypes.Value{
								// 		"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-2"), // Terraform core behavior
								// 		"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
								// 	},
								// ),
							},
						),
						"test_configured_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-1"), // Terraform core behavior
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-2"), // Terraform core behavior
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value"), // Terraform core behavior
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value"), // Terraform core behavior
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
					}),
					Schema: testSchemaDefault,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":               tftypes.NewValue(tftypes.Bool, false),
						"test_computed_float32":            tftypes.NewValue(tftypes.Number, 5.4321),
						"test_computed_float64":            tftypes.NewValue(tftypes.Number, 5.4321),
						"test_computed_int32":              tftypes.NewValue(tftypes.Number, 54321),
						"test_computed_int64":              tftypes.NewValue(tftypes.Number, 54321),
						"test_computed_list":               tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "prior-state")}),
						"test_computed_map":                tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "prior-state")}),
						"test_computed_number":             tftypes.NewValue(tftypes.Number, big.NewFloat(5.4321)),
						"test_computed_object":             tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "prior-state")}),
						"test_computed_set":                tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "prior-state")}),
						"test_computed_string":             tftypes.NewValue(tftypes.String, "two"),
						"test_computed_string_custom_type": tftypes.NewValue(tftypes.String, "two"),
						"test_computed_dynamic":            tftypes.NewValue(tftypes.String, "two"),
						"test_computed_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "prior-state"),
									},
								),
							},
						),
						"test_computed_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "prior-state-attribute"),
									},
								),
							},
						),
						"test_computed_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "prior-state"),
									},
								),
							},
						),
						"test_computed_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "prior-state-attribute"),
									},
								),
							},
						),
						"test_computed_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "prior-state"),
									},
								),
							},
						),
						"test_computed_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "prior-state-attribute"),
									},
								),
							},
						),
						"test_computed_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "prior-state"),
							},
						),
						"test_computed_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "prior-state-attribute"),
							},
						),
						"test_configured_bool":    tftypes.NewValue(tftypes.Bool, false),
						"test_configured_float32": tftypes.NewValue(tftypes.Number, 2.4),
						"test_configured_float64": tftypes.NewValue(tftypes.Number, 2.4),
						"test_configured_int32":   tftypes.NewValue(tftypes.Number, 456),
						"test_configured_int64":   tftypes.NewValue(tftypes.Number, 456),
						"test_configured_list": tftypes.NewValue(
							tftypes.List{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "state-value"),
							},
						),
						"test_configured_map": tftypes.NewValue(
							tftypes.Map{ElementType: tftypes.String},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(tftypes.String, "state-value"),
							},
						),
						"test_configured_number": tftypes.NewValue(tftypes.Number, 2.4),
						"test_configured_object": tftypes.NewValue(
							tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
							map[string]tftypes.Value{
								"a": tftypes.NewValue(tftypes.String, "state-value"),
							},
						),
						"test_configured_set": tftypes.NewValue(
							tftypes.Set{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "state-value"),
							},
						),
						"test_configured_string":             tftypes.NewValue(tftypes.String, "state-value"),
						"test_configured_string_custom_type": tftypes.NewValue(tftypes.String, "state-value"),
						"test_configured_dynamic":            tftypes.NewValue(tftypes.String, "state-value"),
						"test_configured_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-1"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-2"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-1"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-2"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value"),
									},
								),
							},
						),
						"test_configured_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value"),
									},
								),
							},
						),
						"test_configured_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-1"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value-element-1"),
									},
								),
								// TODO: Prevent random attributes map access from causing flaky test
								// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/783
								// tftypes.NewValue(
								// 	tftypes.Object{
								// 		AttributeTypes: map[string]tftypes.Type{
								// 			"computed_attribute":     tftypes.String,
								// 			"configurable_attribute": tftypes.String,
								// 		},
								// 	},
								// 	map[string]tftypes.Value{
								// 		"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-2"),
								// 		"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value-element-2"),
								// 	},
								// ),
							},
						),
						"test_configured_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-1"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value-element-2"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value"),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value"),
							},
						),
						"test_configured_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, "state-value"),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "state-value"),
							},
						),
					}),
					Schema: testSchemaDefault,
				},
				ResourceSchema: testSchemaDefault,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/783
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test_configured_nested_set_attribute"),
						"Duplicate Set Element",
						"This attribute contains duplicate values of: tftypes.Object["+
							"\"computed_attribute\":tftypes.String, "+
							"\"configurable_attribute\":tftypes.String"+
							"]<"+
							"\"computed_attribute\":tftypes.String<\"attribute-default-value\">, "+
							"\"configurable_attribute\":tftypes.String<\"attribute-default-value\">"+
							">",
					),
				},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":               tftypes.NewValue(tftypes.Bool, true),
						"test_computed_float32":            tftypes.NewValue(tftypes.Number, 1.2345),
						"test_computed_float64":            tftypes.NewValue(tftypes.Number, 1.2345),
						"test_computed_int32":              tftypes.NewValue(tftypes.Number, 12345),
						"test_computed_int64":              tftypes.NewValue(tftypes.Number, 12345),
						"test_computed_list":               tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_map":                tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_number":             tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
						"test_computed_object":             tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_set":                tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_string":             tftypes.NewValue(tftypes.String, "one"),
						"test_computed_string_custom_type": tftypes.NewValue(tftypes.String, "one"),
						"test_computed_dynamic":            tftypes.NewValue(tftypes.String, "hello world!"),
						"test_computed_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default"),
									},
								),
							},
						),
						"test_computed_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default-attribute"),
									},
								),
							},
						),
						"test_computed_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default"),
									},
								),
							},
						),
						"test_computed_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default-attribute"),
									},
								),
							},
						),
						"test_computed_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default"),
									},
								),
							},
						),
						"test_computed_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "default-attribute"),
									},
								),
							},
						),
						"test_computed_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "default"),
							},
						),
						"test_computed_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "default-attribute"),
							},
						),
						"test_configured_bool":    tftypes.NewValue(tftypes.Bool, true),
						"test_configured_float32": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_float64": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_int32":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_int64":   tftypes.NewValue(tftypes.Number, 123),
						"test_configured_list": tftypes.NewValue(
							tftypes.List{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_map": tftypes.NewValue(
							tftypes.Map{ElementType: tftypes.String},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_number": tftypes.NewValue(tftypes.Number, 1.2),
						"test_configured_object": tftypes.NewValue(
							tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
							map[string]tftypes.Value{
								"a": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_set": tftypes.NewValue(
							tftypes.Set{ElementType: tftypes.String},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_string":             tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_string_custom_type": tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_dynamic":            tftypes.NewValue(tftypes.String, "config-value"),
						"test_configured_nested_list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
									},
								),
							},
						),
						"test_configured_nested_map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_map_attribute": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
									},
								),
							},
						),
						"test_configured_nested_set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute": tftypes.NewValue(tftypes.String, "attribute-default-value"),
										// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/783
										// "configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "attribute-default-value"),
									},
								),
								// TODO: Prevent random attributes map access from causing flaky test (also fix above data)
								// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/783
								// tftypes.NewValue(
								// 	tftypes.Object{
								// 		AttributeTypes: map[string]tftypes.Type{
								// 			"computed_attribute":     tftypes.String,
								// 			"configurable_attribute": tftypes.String,
								// 		},
								// 	},
								// 	map[string]tftypes.Value{
								// 		"computed_attribute": tftypes.NewValue(tftypes.String, "attribute-default-value"),
								// 		// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/783
								// 		// "configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
								// 		"configurable_attribute": tftypes.NewValue(tftypes.String, "attribute-default-value"),
								// 	},
								// ),
							},
						),
						"test_configured_nested_set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"computed_attribute":     tftypes.String,
										"configurable_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute": tftypes.NewValue(tftypes.String, "attribute-default-value"),
										// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/783
										// "configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-1"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "attribute-default-value"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"computed_attribute":     tftypes.String,
											"configurable_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"computed_attribute": tftypes.NewValue(tftypes.String, "attribute-default-value"),
										// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/783
										// "configurable_attribute": tftypes.NewValue(tftypes.String, "config-value-element-2"),
										"configurable_attribute": tftypes.NewValue(tftypes.String, "attribute-default-value"),
									},
								),
							},
						),
						"test_configured_nested_single": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
						"test_configured_nested_single_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"computed_attribute":     tftypes.String,
									"configurable_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"computed_attribute":     tftypes.NewValue(tftypes.String, "attribute-default-value"),
								"configurable_attribute": tftypes.NewValue(tftypes.String, "config-value"),
							},
						),
					}),
					Schema: testSchemaDefault,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-attributeplanmodifier-request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanRequest,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanRequest,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanRequest,
				},
				ResourceSchema: testSchemaAttributePlanModifierPrivatePlanRequest,
				Resource:       &testprovider.Resource{},
				PriorPrivate:   testPrivateProvider,
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanRequest,
				},
				PlannedPrivate: testPrivateProvider,
			},
		},
		"update-attributeplanmodifier-response-attributeplan-config-change": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				ResourceSchema: testSchemaAttributePlanModifierAttributePlan,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, "test-attributeplanmodifier-value"),
						"test_other_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required":       tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-attributeplanmodifier-response-attributeplan-no-config-change": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed":       tftypes.NewValue(tftypes.String, nil),
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				ResourceSchema: testSchemaAttributePlanModifierAttributePlan,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_computed":       tftypes.String,
							"test_other_computed": tftypes.String,
							"test_required":       tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-attributeplanmodifier-value"),
						// Ideally test_other_computed would be tftypes.UnknownValue, however
						// fixing the behavior without preventing provider developers from
						// leaving or setting plan values to null explicitly is non-trivial.
						// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/183
						// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/456
						"test_other_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required":       tftypes.NewValue(tftypes.String, "test-value"),
					}),
					Schema: testSchemaAttributePlanModifierAttributePlan,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-attributeplanmodifier-response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				ResourceSchema: testSchemaAttributePlanModifierPrivatePlanResponse,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				PlannedPrivate: testPrivateProvider,
			},
		},
		"update-attributeplanmodifier-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierDiagnosticsError,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierDiagnosticsError,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchemaAttributePlanModifierDiagnosticsError,
				},
				ResourceSchema: testSchemaAttributePlanModifierDiagnosticsError,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.WithPath(
						path.Root("test_required"),
						diag.NewErrorDiagnostic("error summary", "error detail"),
					),
				},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierDiagnosticsError,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-attributeplanmodifier-response-requiresreplace": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierRequiresReplace,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierRequiresReplace,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchemaAttributePlanModifierRequiresReplace,
				},
				ResourceSchema: testSchemaAttributePlanModifierRequiresReplace,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierRequiresReplace,
				},
				RequiresReplace: path.Paths{
					path.Root("test_required"),
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-resourcewithmodifyplan-request-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-new-value" {
							resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-resourcewithmodifyplan-request-config-nil-block": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaBlockType, map[string]tftypes.Value{
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
						"test_optional_block": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_optional_one": tftypes.String,
								"test_optional_two": tftypes.String,
							},
						}, nil),
					}),
					Schema: testSchemaBlock,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaBlockType, map[string]tftypes.Value{
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
						"test_optional_block": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_optional_one": tftypes.String,
								"test_optional_two": tftypes.String,
							},
						}, nil),
					}),
					Schema: testSchemaBlock,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaBlockType, map[string]tftypes.Value{
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
						"test_optional_block": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_optional_one": tftypes.String,
								"test_optional_two": tftypes.String,
							},
						}, nil),
					}),
					Schema: testSchemaBlock,
				},
				ResourceSchema: testSchemaBlock,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaDataBlock

						resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

						if data.TestRequired.ValueString() != "test-new-value" {
							resp.Diagnostics.AddError("Unexpected req.Config Value", "Got: "+data.TestRequired.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaBlockType, map[string]tftypes.Value{
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
						"test_optional_block": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_optional_one": tftypes.String,
								"test_optional_two": tftypes.String,
							},
						}, nil),
					}),
					Schema: testSchemaBlock,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-resourcewithmodifyplan-request-proposednewstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						if !data.TestComputed.IsUnknown() {
							resp.Diagnostics.AddError("Unexpected req.Plan Value", "Got: "+data.TestComputed.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-resourcewithmodifyplan-request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ProviderMeta:   testProviderMetaConfig,
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testProviderMetaData

						resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &data)...)

						if data.TestProviderMetaAttribute.ValueString() != "test-provider-meta-value" {
							resp.Diagnostics.AddError("Unexpected req.ProviderMeta Value", "Got: "+data.TestProviderMetaAttribute.ValueString())
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-resourcewithmodifyplan-request-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

						key := "providerKeyOne"
						got, diags := req.Private.GetKey(ctx, key)

						resp.Diagnostics.Append(diags...)

						if string(got) != expected {
							resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
						}
					},
				},
				PriorPrivate: testPrivate,
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testPrivate,
			},
		},
		"update-resourcewithmodifyplan-response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						resp.Diagnostics.AddWarning("warning summary", "warning detail")
						resp.Diagnostics.AddError("error summary", "error detail")
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("warning summary", "warning detail"),
					diag.NewErrorDiagnostic("error summary", "error detail"),
				},
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-resourcewithmodifyplan-response-plannedstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						var data testSchemaData

						resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

						data.TestComputed = types.StringValue("test-plannedstate-value")

						resp.Diagnostics.Append(resp.Plan.Set(ctx, &data)...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-plannedstate-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-resourcewithmodifyplan-response-requiresreplace": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						// This is a strange thing to signal on creation,
						// but the framework does not prevent you from
						// doing it and it might be overly burdensome on
						// provider developers to have the framework raise
						// an error if it is technically valid in the
						// protocol.
						resp.RequiresReplace = path.Paths{
							path.Root("test_required"),
						}
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				RequiresReplace: path.Paths{
					path.Root("test_required"),
				},
				PlannedPrivate: testEmptyPrivate,
			},
		},
		"update-resourcewithmodifyplan-response-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, nil),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchema,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaType, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchema,
				},
				PlannedPrivate: testPrivateProvider,
			},
		},
		"update-resourcewithmodifyplan-attributeplanmodifier-private": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw: tftypes.NewValue(testSchemaTypeComputedRequired, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-new-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaTypeComputedRequired, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-new-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeComputedRequired, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-old-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-old-value"),
					}),
					Schema: testSchema,
				},
				ResourceSchema: testSchemaAttributePlanModifierPrivatePlanResponse,
				Resource: &testprovider.ResourceWithModifyPlan{
					ModifyPlanMethod: func(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
						diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

						resp.Diagnostics.Append(diags...)
					},
				},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeComputedRequired, map[string]tftypes.Value{
						"test_computed": tftypes.NewValue(tftypes.String, "test-new-value"),
						"test_required": tftypes.NewValue(tftypes.String, "test-new-value"),
					}),
					Schema: testSchemaAttributePlanModifierPrivatePlanResponse,
				},
				PlannedPrivate: testPrivateProvider,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.configureProviderReq != nil {
				configureProviderResp := &provider.ConfigureResponse{}
				testCase.server.ConfigureProvider(context.Background(), testCase.configureProviderReq, configureProviderResp)
			}

			response := &fwserver.PlanResourceChangeResponse{}
			testCase.server.PlanResourceChange(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

// Ensure attribute values, especially nested attributes, are round-tripped
// through the server without modification if there is no plan modification
// defined in the schema or resource. This unit testing is not intended to
// include multiple attributes, as that unit testing is covered elsewhere.
//
// The create test cases cover a fully null PriorState.
//
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/993
func TestServerPlanResourceChange_AttributeRoundtrip(t *testing.T) {
	t.Parallel()

	type testCase struct {
		server   *fwserver.Server
		request  *fwserver.PlanResourceChangeRequest
		expected *fwserver.PlanResourceChangeResponse
	}

	type testCaseData struct {
		Schema           schema.Schema
		Config           tftypes.Value
		ProposedNewState tftypes.Value
		PriorState       tftypes.Value
		PlannedState     tftypes.Value
	}

	generateTestCase := func(in testCaseData) testCase {
		return testCase{
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw:    in.Config,
					Schema: in.Schema,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw:    in.ProposedNewState,
					Schema: in.Schema,
				},
				PriorState: &tfsdk.State{
					Raw:    in.PriorState,
					Schema: in.Schema,
				},
				ResourceSchema: in.Schema,
				Resource:       &testprovider.Resource{},
			},
			expected: &fwserver.PlanResourceChangeResponse{
				PlannedPrivate: &privatestate.Data{
					Provider: privatestate.EmptyProviderData(context.Background()),
				},
				PlannedState: &tfsdk.State{
					Raw:    in.PlannedState,
					Schema: in.Schema,
				},
			},
		}
	}

	testCases := map[string]testCase{
		"create-bool-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool_attribute": schema.BoolAttribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-bool-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool_attribute": schema.BoolAttribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				},
			),
		}),
		"create-bool-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool_attribute": schema.BoolAttribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
				},
			),
		}),
		"create-bool-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool_attribute": schema.BoolAttribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				},
			),
		}),
		"create-bool-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool_attribute": schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-bool-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool_attribute": schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool_attribute": tftypes.Bool,
					},
				},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				},
			),
		}),
		"create-float32-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float32_attribute": schema.Float32Attribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-float32-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float32_attribute": schema.Float32Attribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-float32-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float32_attribute": schema.Float32Attribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
		}),
		"create-float32-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float32_attribute": schema.Float32Attribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-float32-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float32_attribute": schema.Float32Attribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-float32-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float32_attribute": schema.Float32Attribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-float64-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float64_attribute": schema.Float64Attribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-float64-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float64_attribute": schema.Float64Attribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-float64-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float64_attribute": schema.Float64Attribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
		}),
		"create-float64-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float64_attribute": schema.Float64Attribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-float64-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float64_attribute": schema.Float64Attribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-float64-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float64_attribute": schema.Float64Attribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"float64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-int32-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int32_attribute": schema.Int32Attribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Compute nulls as unknown
				},
			),
		}),
		"create-int32-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int32_attribute": schema.Int32Attribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-int32-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int32_attribute": schema.Int32Attribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
		}),
		"create-int32-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int32_attribute": schema.Int32Attribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-int32-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int32_attribute": schema.Int32Attribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-int32-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int32_attribute": schema.Int32Attribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int32_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-int64-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int64_attribute": schema.Int64Attribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Compute nulls as unknown
				},
			),
		}),
		"create-int64-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int64_attribute": schema.Int64Attribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-int64-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int64_attribute": schema.Int64Attribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
		}),
		"create-int64-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int64_attribute": schema.Int64Attribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-int64-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int64_attribute": schema.Int64Attribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-int64-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int64_attribute": schema.Int64Attribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"int64_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-list-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-list-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-list-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
		}),
		"create-list-optional-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		}),
		"create-list-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-list-optional-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-list-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-list-optional-and-computed-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		}),
		"create-list-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-list-optional-and-computed-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-list-nested-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-list-nested-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-list-nested-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
		}),

		"create-list-nested-optional-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
		}),
		"create-list-nested-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-list-nested-optional-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
		}),
		"create-list-nested-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknowns
					),
				},
			),
		}),
		"create-list-nested-optional-and-computed-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil, // Computed nulls as unknowns does not apply to the nested object itself
							),
						},
					),
				},
			),
		}),
		"create-list-nested-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-list-nested-optional-and-computed-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested_attribute": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested_attribute": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
		}),
		"create-map-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue, // Computed nulls as unknowns
					),
				},
			),
		}),
		"create-map-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-map-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
		}),
		"create-map-optional-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		}),
		"create-map-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-map-optional-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-map-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue, // Computed nulls as unknowns
					),
				},
			),
		}),
		"create-map-optional-and-computed-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		}),
		"create-map-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-map-optional-and-computed-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-map-nested-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknowns
					),
				},
			),
		}),
		"create-map-nested-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-map-nested-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
		}),
		"create-map-nested-optional-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
		}),
		"create-map-nested-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-map-nested-optional-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
		}),
		"create-map-nested-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknowns
					),
				},
			),
		}),
		"create-map-nested-optional-and-computed-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil, // Computed nulls as unknowns does not apply to the nested object itself
							),
						},
					),
				},
			),
		}),
		"create-map-nested-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-map-nested-optional-and-computed-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested_attribute": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"null": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
		}),
		"create-number-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"number_attribute": schema.NumberAttribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Computed nulls as unknowns
				},
			),
		}),
		"create-number-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"number_attribute": schema.NumberAttribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-number-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"number_attribute": schema.NumberAttribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
		}),
		"create-number-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"number_attribute": schema.NumberAttribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-number-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"number_attribute": schema.NumberAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // Computed nulls as unknowns
				},
			),
		}),
		"create-number-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"number_attribute": schema.NumberAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"number_attribute": tftypes.Number,
					},
				},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				},
			),
		}),
		"create-object-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknowns
					),
				},
			),
		}),
		"create-object-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-object-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
		}),
		"create-object-optional-null-attribute": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		}),
		"create-object-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-object-optional-unknown-attribute": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-object-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-object-optional-and-computed-null-attribute": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		}),
		"create-object-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-object-optional-and-computed-unknown-attribute": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"string_attribute": types.StringType,
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-set-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-set-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-set-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
		}),
		"create-set-optional-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		}),
		"create-set-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-set-optional-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-set-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-set-optional-and-computed-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		}),
		"create-set-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-set-optional-and-computed-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-set-nested-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-set-nested-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-set-nested-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
		}),
		"create-set-nested-optional-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
		}),
		"create-set-nested-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-set-nested-optional-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
		}),
		"create-set-nested-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-set-nested-optional-and-computed-null-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								nil, // Computed nulls as unknowns does not apply to the nested object itself
							),
						},
					),
				},
			),
		}),
		"create-set-nested-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-set-nested-optional-and-computed-unknown-element": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested_attribute": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested_attribute": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
		}),
		"create-single-nested-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Computed: true,
							},
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-single-nested-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Computed: true,
							},
						},
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-single-nested-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Optional: true,
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
		}),
		"create-single-nested-optional-null-attribute": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Optional: true,
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		}),
		"create-single-nested-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Optional: true,
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-single-nested-optional-unknown-attribute": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Optional: true,
							},
						},
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-single-nested-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue, // Computed nulls as unknown
					),
				},
			),
		}),
		"create-single-nested-optional-and-computed-null-attribute": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue), // Computed nulls as unknown
						},
					),
				},
			),
		}),
		"create-single-nested-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
		}),
		"create-single-nested-optional-and-computed-unknown-attribute": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested_attribute": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						},
					),
				},
			),
		}),
		"create-string-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-string-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
		}),
		"create-string-optional-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil),
				},
			),
		}),
		"create-string-optional-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
		}),
		"create-string-optional-and-computed-null": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue), // Computed nulls as unknown
				},
			),
		}),
		"create-string-optional-and-computed-unknown": generateTestCase(testCaseData{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			Config: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
			ProposedNewState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
			PriorState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				nil,
			),
			PlannedState: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_attribute": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
		}),
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwserver.PlanResourceChangeResponse{}

			testCase.server.PlanResourceChange(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
