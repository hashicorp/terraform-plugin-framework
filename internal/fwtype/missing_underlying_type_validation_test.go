// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwtype_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwtype"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestContainsMissingUnderlyingType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attrTyp  attr.Type
		expected bool
	}{
		"nil": {
			attrTyp:  nil,
			expected: true,
		},
		"bool": {
			attrTyp:  types.BoolType,
			expected: false,
		},
		"dynamic": {
			attrTyp:  types.DynamicType,
			expected: false,
		},
		"float64": {
			attrTyp:  types.Float64Type,
			expected: false,
		},
		"int64": {
			attrTyp:  types.Float64Type,
			expected: false,
		},
		"list-nil": {
			attrTyp:  types.ListType{},
			expected: true,
		},
		"list-bool": {
			attrTyp: types.ListType{
				ElemType: types.BoolType,
			},
			expected: false,
		},
		"list-dynamic": {
			attrTyp: types.ListType{
				ElemType: types.DynamicType,
			},
			expected: false,
		},
		"list-float64": {
			attrTyp: types.ListType{
				ElemType: types.Float64Type,
			},
			expected: false,
		},
		"list-int64": {
			attrTyp: types.ListType{
				ElemType: types.Int64Type,
			},
			expected: false,
		},
		"list-list-nil": {
			attrTyp: types.ListType{
				ElemType: types.ListType{},
			},
			expected: true,
		},
		"list-list-string": {
			attrTyp: types.ListType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"list-map-nil": {
			attrTyp: types.ListType{
				ElemType: types.MapType{},
			},
			expected: true,
		},
		"list-map-string": {
			attrTyp: types.ListType{
				ElemType: types.MapType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"list-object-nil": {
			attrTyp: types.ListType{
				ElemType: types.ObjectType{},
			},
			expected: false, // expected as objects can be empty
		},
		"list-object-attrtypes": {
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
		"list-object-attrtypes-nil": {
			attrTyp: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"bool": types.BoolType,
						"nil":  nil,
					},
				},
			},
			expected: true,
		},
		"list-number": {
			attrTyp: types.ListType{
				ElemType: types.NumberType,
			},
			expected: false,
		},
		"list-set-nil": {
			attrTyp: types.ListType{
				ElemType: types.SetType{},
			},
			expected: true,
		},
		"list-set-string": {
			attrTyp: types.ListType{
				ElemType: types.SetType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"list-string": {
			attrTyp: types.ListType{
				ElemType: types.StringType,
			},
			expected: false,
		},
		"list-tuple-nil": {
			attrTyp: types.ListType{
				ElemType: types.TupleType{},
			},
			expected: false, // expected as tuples can be empty
		},
		"list-tuple-elemtypes": {
			attrTyp: types.ListType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.StringType,
					},
				},
			},
			expected: false,
		},
		"list-tuple-elemtypes-nil": {
			attrTyp: types.ListType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.StringType,
						nil,
					},
				},
			},
			expected: true,
		},
		"map-nil": {
			attrTyp:  types.MapType{},
			expected: true,
		},
		"map-bool": {
			attrTyp: types.MapType{
				ElemType: types.BoolType,
			},
			expected: false,
		},
		"map-dynamic": {
			attrTyp: types.MapType{
				ElemType: types.DynamicType,
			},
			expected: false,
		},
		"map-float64": {
			attrTyp: types.MapType{
				ElemType: types.Float64Type,
			},
			expected: false,
		},
		"map-int64": {
			attrTyp: types.MapType{
				ElemType: types.Int64Type,
			},
			expected: false,
		},
		"map-list-nil": {
			attrTyp: types.MapType{
				ElemType: types.ListType{},
			},
			expected: true,
		},
		"map-list-string": {
			attrTyp: types.MapType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"map-map-nil": {
			attrTyp: types.MapType{
				ElemType: types.MapType{},
			},
			expected: true,
		},
		"map-map-string": {
			attrTyp: types.MapType{
				ElemType: types.MapType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"map-object-nil": {
			attrTyp: types.MapType{
				ElemType: types.ObjectType{},
			},
			expected: false, // expected as objects can be empty
		},
		"map-object-attrtypes": {
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
		"map-object-attrtypes-nil": {
			attrTyp: types.MapType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"bool": types.BoolType,
						"nil":  nil,
					},
				},
			},
			expected: true,
		},
		"map-number": {
			attrTyp: types.MapType{
				ElemType: types.NumberType,
			},
			expected: false,
		},
		"map-set-nil": {
			attrTyp: types.MapType{
				ElemType: types.SetType{},
			},
			expected: true,
		},
		"map-set-string": {
			attrTyp: types.MapType{
				ElemType: types.SetType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"map-string": {
			attrTyp: types.MapType{
				ElemType: types.StringType,
			},
			expected: false,
		},
		"map-tuple-nil": {
			attrTyp: types.MapType{
				ElemType: types.TupleType{},
			},
			expected: false, // expected as tuples can be empty
		},
		"map-tuple-elemtypes": {
			attrTyp: types.MapType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.StringType,
					},
				},
			},
			expected: false,
		},
		"map-tuple-elemtypes-nil": {
			attrTyp: types.MapType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.StringType,
						nil,
					},
				},
			},
			expected: true,
		},
		"number": {
			attrTyp:  types.NumberType,
			expected: false,
		},
		"object-nil": {
			attrTyp:  types.ObjectType{},
			expected: false, // expected as objects can be empty
		},
		"object-list-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{},
				},
			},
			expected: true,
		},
		"object-list-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"object-list-list-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.ListType{},
					},
				},
			},
			expected: true,
		},
		"object-list-list-string": {
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
		"object-list-map-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.MapType{},
					},
				},
			},
			expected: true,
		},
		"object-list-map-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.MapType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"object-list-object-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.ObjectType{},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"object-list-object-attrtypes": {
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
		"object-list-object-attrtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool": types.BoolType,
								"nil":  nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"object-list-set-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.SetType{},
					},
				},
			},
			expected: true,
		},
		"object-list-set-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"object-list-tuple-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.TupleType{},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"object-list-tuple-elemtypes": {
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
		"object-list-tuple-elemtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"object-map-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{},
				},
			},
			expected: true,
		},
		"object-map-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"object-map-list-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.ListType{},
					},
				},
			},
			expected: true,
		},
		"object-map-list-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"object-map-map-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.MapType{},
					},
				},
			},
			expected: true,
		},
		"object-map-map-string": {
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
		"object-map-object-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.ObjectType{},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"object-map-object-attrtypes": {
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
		"object-map-object-attrtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool": types.BoolType,
								"nil":  nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"object-map-set-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.SetType{},
					},
				},
			},
			expected: true,
		},
		"object-map-set-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"object-map-tuple-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.TupleType{},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"object-map-tuple-elemtypes": {
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
		"object-map-tuple-elemtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"object-object-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"object": types.ObjectType{},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"object-object-attrtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"object": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool":    types.BoolType,
							"float64": types.Float64Type,
						},
					},
				},
			},
			expected: false,
		},
		"object-object-attrtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"object": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool": types.BoolType,
							"nil":  nil,
						},
					},
				},
			},
			expected: true,
		},
		"object-set-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{},
				},
			},
			expected: true,
		},
		"object-set-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"object-set-list-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.ListType{},
					},
				},
			},
			expected: true,
		},
		"object-set-list-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"object-set-map-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.MapType{},
					},
				},
			},
			expected: true,
		},
		"object-set-map-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.MapType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"object-set-object-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.ObjectType{},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"object-set-object-attrtypes": {
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
		"object-set-object-attrtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool": types.BoolType,
								"nil":  nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"object-set-set-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.SetType{},
					},
				},
			},
			expected: true,
		},
		"object-set-set-string": {
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
		"object-set-tuple-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.TupleType{},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"object-set-tuple-elemtypes": {
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
		"object-set-tuple-elemtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"object-tuple-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"object-tuple-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{types.StringType},
					},
				},
			},
			expected: false,
		},
		"object-tuple-list-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{types.ListType{}},
					},
				},
			},
			expected: true,
		},
		"object-tuple-list-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{
							types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-tuple-map-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{types.MapType{}},
					},
				},
			},
			expected: true,
		},
		"object-tuple-map-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{
							types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-tuple-object-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{types.ObjectType{}},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"object-tuple-object-attrtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{
							types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"bool":    types.BoolType,
									"float64": types.Float64Type,
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-tuple-object-attrtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{
							types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"bool": types.BoolType,
									"nil":  nil,
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"object-tuple-set-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{types.SetType{}},
					},
				},
			},
			expected: true,
		},
		"object-tuple-set-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{
							types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-tuple-tuple-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{types.TupleType{}},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"object-tuple-tuple-elemtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{
							types.TupleType{
								ElemTypes: []attr.Type{
									types.BoolType,
									types.Float64Type,
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-tuple-tuple-elemtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"tuple": types.TupleType{
						ElemTypes: []attr.Type{
							types.TupleType{
								ElemTypes: []attr.Type{
									types.BoolType,
									nil,
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set-nil": {
			attrTyp:  types.SetType{},
			expected: true,
		},
		"set-bool": {
			attrTyp: types.SetType{
				ElemType: types.BoolType,
			},
			expected: false,
		},
		"set-dynamic": {
			attrTyp: types.SetType{
				ElemType: types.DynamicType,
			},
			expected: false,
		},
		"set-float64": {
			attrTyp: types.SetType{
				ElemType: types.Float64Type,
			},
			expected: false,
		},
		"set-int64": {
			attrTyp: types.SetType{
				ElemType: types.Int64Type,
			},
			expected: false,
		},
		"set-list-nil": {
			attrTyp: types.SetType{
				ElemType: types.ListType{},
			},
			expected: true,
		},
		"set-list-string": {
			attrTyp: types.SetType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"set-map-nil": {
			attrTyp: types.SetType{
				ElemType: types.MapType{},
			},
			expected: true,
		},
		"set-map-string": {
			attrTyp: types.SetType{
				ElemType: types.MapType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"set-object-nil": {
			attrTyp: types.SetType{
				ElemType: types.ObjectType{},
			},
			expected: false, // expected as objects can be empty
		},
		"set-object-attrtypes": {
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
		"set-object-attrtypes-nil": {
			attrTyp: types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"bool": types.BoolType,
						"nil":  nil,
					},
				},
			},
			expected: true,
		},
		"set-number": {
			attrTyp: types.SetType{
				ElemType: types.NumberType,
			},
			expected: false,
		},
		"set-set-nil": {
			attrTyp: types.SetType{
				ElemType: types.SetType{},
			},
			expected: true,
		},
		"set-set-string": {
			attrTyp: types.SetType{
				ElemType: types.SetType{
					ElemType: types.StringType,
				},
			},
			expected: false,
		},
		"set-string": {
			attrTyp: types.SetType{
				ElemType: types.StringType,
			},
			expected: false,
		},
		"set-tuple-nil": {
			attrTyp: types.SetType{
				ElemType: types.TupleType{},
			},
			expected: false, // expected as tuples can be empty
		},
		"set-tuple-elemtypes": {
			attrTyp: types.SetType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.StringType,
					},
				},
			},
			expected: false,
		},
		"set-tuple-elemtypes-nil": {
			attrTyp: types.SetType{
				ElemType: types.TupleType{
					ElemTypes: []attr.Type{
						types.StringType,
						nil,
					},
				},
			},
			expected: true,
		},
		"string": {
			attrTyp:  types.StringType,
			expected: false,
		},
		"tuple-nil": {
			attrTyp:  types.TupleType{},
			expected: false, // expected as tuples can be empty
		},
		"tuple-bool": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{types.BoolType},
			},
			expected: false,
		},
		"tuple-dynamic": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{types.DynamicType},
			},
			expected: false,
		},
		"tuple-float64": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{types.Float64Type},
			},
			expected: false,
		},
		"tuple-int64": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{types.Int64Type},
			},
			expected: false,
		},
		"tuple-list-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{},
				},
			},
			expected: true,
		},
		"tuple-list-string": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"tuple-list-list-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{
						ElemType: types.ListType{},
					},
				},
			},
			expected: true,
		},
		"tuple-list-list-string": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{
						ElemType: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"tuple-list-object-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{
						ElemType: types.ObjectType{},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"tuple-list-object-attrtypes": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{
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
		"tuple-list-object-attrtypes-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool": types.BoolType,
								"nil":  nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"tuple-list-tuple-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{
						ElemType: types.TupleType{},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"tuple-list-tuple-elemtypes": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{
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
		"tuple-list-tuple-elemtypes-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ListType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"tuple-map-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{},
				},
			},
			expected: true,
		},
		"tuple-map-string": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"tuple-map-map-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{
						ElemType: types.MapType{},
					},
				},
			},
			expected: true,
		},
		"tuple-map-map-string": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{
						ElemType: types.MapType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"tuple-map-object-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{
						ElemType: types.ObjectType{},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"tuple-map-object-attrtypes": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{
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
		"tuple-map-object-attrtypes-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool": types.BoolType,
								"nil":  nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"tuple-map-tuple-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{
						ElemType: types.TupleType{},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"tuple-map-tuple-elemtypes": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{
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
		"tuple-map-tuple-elemtypes-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.MapType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"tuple-number": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{types.NumberType},
			},
			expected: false,
		},
		"tuple-object-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ObjectType{},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"tuple-object-attrtypes": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool":    types.BoolType,
							"float64": types.Float64Type,
						},
					},
				},
			},
			expected: false,
		},
		"tuple-object-attrtypes-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool": types.BoolType,
							"nil":  nil,
						},
					},
				},
			},
			expected: true,
		},
		"tuple-set-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{},
				},
			},
			expected: true,
		},
		"tuple-set-object-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{
						ElemType: types.ObjectType{},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"tuple-set-object-attrtypes": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{
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
		"tuple-set-object-attrtypes-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool": types.BoolType,
								"nil":  nil,
							},
						},
					},
				},
			},
			expected: true,
		},
		"tuple-set-set-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{
						ElemType: types.SetType{},
					},
				},
			},
			expected: true,
		},
		"tuple-set-set-string": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{
						ElemType: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"tuple-set-string": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"tuple-set-tuple-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{
						ElemType: types.TupleType{},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"tuple-set-tuple-elemtypes": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{
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
		"tuple-set-tuple-elemtypes-nil": {
			attrTyp: types.TupleType{
				ElemTypes: []attr.Type{
					types.SetType{
						ElemType: types.TupleType{
							ElemTypes: []attr.Type{
								types.BoolType,
								nil,
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

			got := fwtype.ContainsMissingUnderlyingType(testCase.attrTyp)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
