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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
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
		"string-value":                   tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":                     tftypes.NewValue(tftypes.String, nil),
		"string-nil-computed":            tftypes.NewValue(tftypes.String, nil),
		"string-nil-optional-computed":   tftypes.NewValue(tftypes.String, nil),
		"string-value-optional-computed": tftypes.NewValue(tftypes.String, "hello, world"),
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
		"string-value":                   tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":                     tftypes.NewValue(tftypes.String, nil),
		"string-nil-computed":            tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"string-nil-optional-computed":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"string-value-optional-computed": tftypes.NewValue(tftypes.String, "hello, world"),
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
			"test_computed_float64":                 tftypes.Number,
			"test_computed_int64":                   tftypes.Number,
			"test_computed_list":                    tftypes.List{ElementType: tftypes.String},
			"test_computed_map":                     tftypes.Map{ElementType: tftypes.String},
			"test_computed_number":                  tftypes.Number,
			"test_computed_object":                  tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}},
			"test_computed_set":                     tftypes.Set{ElementType: tftypes.String},
			"test_computed_string":                  tftypes.String,
			"test_computed_nested_list":             tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_list_attribute":   tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_map":              tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_map_attribute":    tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_set":              tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_set_attribute":    tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}}},
			"test_computed_nested_single":           tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}},
			"test_computed_nested_single_attribute": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"string_attribute": tftypes.String}},
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
			"test_computed_float64": schema.Float64Attribute{
				Computed: true,
				Default:  float64default.StaticFloat64(1.2345),
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

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.PlanResourceChangeRequest
		expectedResponse *fwserver.PlanResourceChangeResponse
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
						"test_computed_bool":    tftypes.NewValue(tftypes.Bool, nil),
						"test_computed_float64": tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int64":   tftypes.NewValue(tftypes.Number, nil),
						"test_computed_list":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_computed_map":     tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"test_computed_number":  tftypes.NewValue(tftypes.Number, nil),
						"test_computed_object":  tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, nil),
						"test_computed_set":     tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
						"test_computed_string":  tftypes.NewValue(tftypes.String, nil),
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
					}),
					Schema: testSchemaDefault,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":    tftypes.NewValue(tftypes.Bool, nil),
						"test_computed_float64": tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int64":   tftypes.NewValue(tftypes.Number, nil),
						"test_computed_list":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_computed_map":     tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"test_computed_number":  tftypes.NewValue(tftypes.Number, nil),
						"test_computed_object":  tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, nil),
						"test_computed_set":     tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
						"test_computed_string":  tftypes.NewValue(tftypes.String, nil),
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
						"test_computed_bool":    tftypes.NewValue(tftypes.Bool, true),
						"test_computed_float64": tftypes.NewValue(tftypes.Number, 1.2345),
						"test_computed_int64":   tftypes.NewValue(tftypes.Number, 12345),
						"test_computed_list":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_map":     tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_number":  tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
						"test_computed_object":  tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_set":     tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_string":  tftypes.NewValue(tftypes.String, "one"),
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
						"test_computed_bool":    tftypes.NewValue(tftypes.Bool, nil),
						"test_computed_float64": tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int64":   tftypes.NewValue(tftypes.Number, nil),
						"test_computed_list":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_computed_map":     tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"test_computed_number":  tftypes.NewValue(tftypes.Number, nil),
						"test_computed_object":  tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, nil),
						"test_computed_set":     tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
						"test_computed_string":  tftypes.NewValue(tftypes.String, nil),
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
					}),
					Schema: testSchemaDefault,
				},
				ProposedNewState: &tfsdk.Plan{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":    tftypes.NewValue(tftypes.Bool, nil),
						"test_computed_float64": tftypes.NewValue(tftypes.Number, nil),
						"test_computed_int64":   tftypes.NewValue(tftypes.Number, nil),
						"test_computed_list":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_computed_map":     tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"test_computed_number":  tftypes.NewValue(tftypes.Number, nil),
						"test_computed_object":  tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, nil),
						"test_computed_set":     tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
						"test_computed_string":  tftypes.NewValue(tftypes.String, nil),
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
					}),
					Schema: testSchemaDefault,
				},
				PriorState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":    tftypes.NewValue(tftypes.Bool, false),
						"test_computed_float64": tftypes.NewValue(tftypes.Number, 5.4321),
						"test_computed_int64":   tftypes.NewValue(tftypes.Number, 54321),
						"test_computed_list":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "prior-state")}),
						"test_computed_map":     tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "prior-state")}),
						"test_computed_number":  tftypes.NewValue(tftypes.Number, big.NewFloat(5.4321)),
						"test_computed_object":  tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "prior-state")}),
						"test_computed_set":     tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "prior-state")}),
						"test_computed_string":  tftypes.NewValue(tftypes.String, "two"),
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
					}),
					Schema: testSchemaDefault,
				},
				ResourceSchema: testSchemaDefault,
				Resource:       &testprovider.Resource{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{
				PlannedState: &tfsdk.State{
					Raw: tftypes.NewValue(testSchemaTypeDefault, map[string]tftypes.Value{
						"test_computed_bool":    tftypes.NewValue(tftypes.Bool, true),
						"test_computed_float64": tftypes.NewValue(tftypes.Number, 1.2345),
						"test_computed_int64":   tftypes.NewValue(tftypes.Number, 12345),
						"test_computed_list":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_map":     tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_number":  tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
						"test_computed_object":  tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String}}, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_set":     tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "default")}),
						"test_computed_string":  tftypes.NewValue(tftypes.String, "one"),
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

			response := &fwserver.PlanResourceChangeResponse{}
			testCase.server.PlanResourceChange(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
