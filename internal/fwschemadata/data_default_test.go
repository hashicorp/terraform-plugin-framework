// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDataDefault(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		data          *fwschemadata.Data
		rawConfig     tftypes.Value
		expected      *fwschemadata.Data
		expectedDiags diag.Diagnostics
	}{
		"bool-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, true), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
		},
		"bool-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
		},
		"bool-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, true),
					},
				),
			},
		},
		"bool-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
		},
		"float64-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  float64default.StaticFloat64(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, 5.4321), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  float64default.StaticFloat64(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
		},
		"float64-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
		},
		"float64-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  float64default.StaticFloat64(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  float64default.StaticFloat64(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 5.4321),
					},
				),
			},
		},
		"float64-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
		},
		"int64-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  int64default.StaticInt64(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, 54321), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  int64default.StaticInt64(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
		},
		"int64-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
		},
		"int64-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  int64default.StaticInt64(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  int64default.StaticInt64(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 54321),
					},
				),
			},
		},
		"int64-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
		},
		"list-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "one"),
					}),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"list-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"list-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "two"),
						}),
					},
				),
			},
		},
		"list-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"map-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.StringType,
									map[string]attr.Value{
										"b": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.String, "one"),
					}),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.StringType,
									map[string]attr.Value{
										"b": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"map-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.MapType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.MapType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"map-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.StringType,
									map[string]attr.Value{
										"b": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.StringType,
									map[string]attr.Value{
										"b": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"b": tftypes.NewValue(tftypes.String, "two"),
						}),
					},
				),
			},
		},
		"map-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"number-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  numberdefault.StaticBigFloat(big.NewFloat(5.4321)),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(5.4321)), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  numberdefault.StaticBigFloat(big.NewFloat(5.4321)),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
		},
		"number-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.NumberType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.NumberType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
		},
		"number-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  numberdefault.StaticBigFloat(big.NewFloat(5.4321)),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  numberdefault.StaticBigFloat(big.NewFloat(5.4321)),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(5.4321)),
					},
				),
			},
		},
		"number-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
		},
		"object-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{"a": types.StringType},
									map[string]attr.Value{
										"a": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
					}, map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.String, "one"),
					}),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{"a": types.StringType},
									map[string]attr.Value{
										"a": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"object-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.Attribute{
							Computed: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"a": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.Attribute{
							Computed: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"a": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"object-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{"a": types.StringType},
									map[string]attr.Value{
										"a": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{"a": types.StringType},
									map[string]attr.Value{
										"a": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "two"),
						}),
					},
				),
			},
		},
		"object-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default:        nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default:        nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"set-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "one"),
					}),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"set-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.SetType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.SetType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"set-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "two"),
						}),
					},
				),
			},
		},
		"set-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  stringdefault.StaticString("two"),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, "two"), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  stringdefault.StaticString("two"),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  stringdefault.StaticString("two"),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  stringdefault.StaticString("two"),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "two"),
					},
				),
			},
		},
		"string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"list-nested-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
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
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
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
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
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
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
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
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
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
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
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
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
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
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
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
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
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
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
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
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
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
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
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
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
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
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
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
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
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
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
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
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
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
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{
										"string_attribute": types.StringType,
									},
									map[string]attr.Value{
										"string_attribute": types.StringValue("two"),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, "one"),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{
										"string_attribute": types.StringType,
									},
									map[string]attr.Value{
										"string_attribute": types.StringValue("two"),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{
										"string_attribute": types.StringType,
									},
									map[string]attr.Value{
										"string_attribute": types.StringValue("two"),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{
										"string_attribute": types.StringType,
									},
									map[string]attr.Value{
										"string_attribute": types.StringValue("two"),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "two"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  stringdefault.StaticString("two"),
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, "one"),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  stringdefault.StaticString("two"),
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  stringdefault.StaticString("two"),
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  stringdefault.StaticString("two"),
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "two"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  nil,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
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
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  nil,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.data.TransformDefaults(context.Background(), testCase.rawConfig)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.data, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
