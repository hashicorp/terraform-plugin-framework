// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschema_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStructuralTypeContainsDynamic(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attrTyp  attr.Type
		expected bool
	}{
		"nil": {
			attrTyp:  nil,
			expected: false,
		},
		"dynamic": {
			attrTyp:  types.DynamicType,
			expected: false,
		},
		"primitive": {
			attrTyp:  types.StringType,
			expected: false,
		},
		"obj-list-missing": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{},
				},
			},
			expected: false,
		},
		"obj-list-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"obj-list-list-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"obj-list-obj-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool":    types.BoolType,
								"float64": types.Float64Type,
							},
						},
					},
				},
			},
			expected: false,
		},
		"obj-list-tuple-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								types.Float64Type,
							},
						},
					},
				},
			},
			expected: false,
		},
		"obj-list-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.DynamicType,
					},
				},
			},
			expected: true,
		},
		"obj-list-list-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.ListType{
							ElemType: types.DynamicType,
						},
					},
				},
			},
			expected: true,
		},
		"obj-list-obj-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool":    types.BoolType,
								"dynamic": types.DynamicType,
							},
						},
					},
				},
			},
			expected: true,
		},
		"obj-list-tuple-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								types.DynamicType,
							},
						},
					},
				},
			},
			expected: true,
		},
		"obj-map-missing": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{},
				},
			},
			expected: false,
		},
		"obj-map-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"obj-map-map-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.MapType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"obj-map-obj-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool":    types.BoolType,
								"float64": types.Float64Type,
							},
						},
					},
				},
			},
			expected: false,
		},
		"obj-map-tuple-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								types.Float64Type,
							},
						},
					},
				},
			},
			expected: false,
		},
		"obj-map-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.DynamicType,
					},
				},
			},
			expected: true,
		},
		"obj-map-map-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.MapType{
							ElemType: types.DynamicType,
						},
					},
				},
			},
			expected: true,
		},
		"obj-map-obj-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool":    types.BoolType,
								"dynamic": types.DynamicType,
							},
						},
					},
				},
			},
			expected: true,
		},
		"obj-map-tuple-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								types.DynamicType,
							},
						},
					},
				},
			},
			expected: true,
		},
		"obj-set-missing": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{},
				},
			},
			expected: false,
		},
		"obj-set-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"obj-set-set-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"obj-set-obj-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool":    types.BoolType,
								"float64": types.Float64Type,
							},
						},
					},
				},
			},
			expected: false,
		},
		"obj-set-tuple-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								types.Float64Type,
							},
						},
					},
				},
			},
			expected: false,
		},
		"obj-set-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.DynamicType,
					},
				},
			},
			expected: true,
		},
		"obj-set-set-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.SetType{
							ElemType: types.DynamicType,
						},
					},
				},
			},
			expected: true,
		},
		"obj-set-obj-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool":    types.BoolType,
								"dynamic": types.DynamicType,
							},
						},
					},
				},
			},
			expected: true,
		},
		"obj-set-tuple-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								types.DynamicType,
							},
						},
					},
				},
			},
			expected: true,
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwschema.StructuralTypeContainsDynamic(testCase.attrTyp)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
func TestCollectionTypeContainsDynamic(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attrTyp  attr.Type
		expected bool
	}{
		"nil": {
			attrTyp:  nil,
			expected: false,
		},
		"primitive": {
			attrTyp:  types.StringType,
			expected: false,
		},
		"list-missing": {
			attrTyp:  types.ListType{},
			expected: false,
		},
		"list-static": {
			attrTyp: types.ListType{
				ElemType: types.StringType,
			},
			expected: false,
		},
		"list-list-static": {
			attrTyp: types.ListType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"list-obj-static": {
			attrTyp: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"bool":    types.BoolType,
						"float64": types.Float64Type,
					},
				},
			},
			expected: false,
		},
		"list-tuple-static": {
			attrTyp: types.ListType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.BoolType,
						types.Float64Type,
					},
				},
			},
			expected: false,
		},
		"list-dynamic": {
			attrTyp: types.ListType{
				ElemType: types.DynamicType,
			},
			expected: true,
		},
		"list-list-dynamic": {
			attrTyp: types.ListType{
				ElemType: types.ListType{
					ElemType: types.DynamicType,
				},
			},
			expected: true,
		},
		"list-obj-dynamic": {
			attrTyp: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"bool":    types.BoolType,
						"dynamic": types.DynamicType,
					},
				},
			},
			expected: true,
		},
		"list-tuple-dynamic": {
			attrTyp: types.ListType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.BoolType,
						types.DynamicType,
					},
				},
			},
			expected: true,
		},
		"map-missing": {
			attrTyp:  types.MapType{},
			expected: false,
		},
		"map-static": {
			attrTyp: types.MapType{
				ElemType: types.StringType,
			},
			expected: false,
		},
		"map-map-static": {
			attrTyp: types.MapType{
				ElemType: types.MapType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"map-obj-static": {
			attrTyp: types.MapType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"bool":    types.BoolType,
						"float64": types.Float64Type,
					},
				},
			},
			expected: false,
		},
		"map-tuple-static": {
			attrTyp: types.MapType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.BoolType,
						types.Float64Type,
					},
				},
			},
			expected: false,
		},
		"map-dynamic": {
			attrTyp: types.MapType{
				ElemType: types.DynamicType,
			},
			expected: true,
		},
		"map-map-dynamic": {
			attrTyp: types.MapType{
				ElemType: types.MapType{
					ElemType: types.DynamicType,
				},
			},
			expected: true,
		},
		"map-obj-dynamic": {
			attrTyp: types.MapType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"bool":    types.BoolType,
						"dynamic": types.DynamicType,
					},
				},
			},
			expected: true,
		},
		"map-tuple-dynamic": {
			attrTyp: types.MapType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.BoolType,
						types.DynamicType,
					},
				},
			},
			expected: true,
		},
		"set-missing": {
			attrTyp:  types.SetType{},
			expected: false,
		},
		"set-static": {
			attrTyp: types.SetType{
				ElemType: types.StringType,
			},
			expected: false,
		},
		"set-set-static": {
			attrTyp: types.SetType{
				ElemType: types.SetType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"set-obj-static": {
			attrTyp: types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"bool":    types.BoolType,
						"float64": types.Float64Type,
					},
				},
			},
			expected: false,
		},
		"set-tuple-static": {
			attrTyp: types.SetType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.BoolType,
						types.Float64Type,
					},
				},
			},
			expected: false,
		},
		"set-dynamic": {
			attrTyp: types.SetType{
				ElemType: types.DynamicType,
			},
			expected: true,
		},
		"set-set-dynamic": {
			attrTyp: types.SetType{
				ElemType: types.SetType{
					ElemType: types.DynamicType,
				},
			},
			expected: true,
		},
		"set-obj-dynamic": {
			attrTyp: types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"bool":    types.BoolType,
						"dynamic": types.DynamicType,
					},
				},
			},
			expected: true,
		},
		"set-tuple-dynamic": {
			attrTyp: types.SetType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.BoolType,
						types.DynamicType,
					},
				},
			},
			expected: true,
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwschema.CollectionTypeContainsDynamic(testCase.attrTyp)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
