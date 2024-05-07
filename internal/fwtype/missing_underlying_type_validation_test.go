// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwtype_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwtype"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
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
		"custom-bool": {
			attrTyp:  testtypes.BoolType{},
			expected: false,
		},
		"custom-dynamic": {
			attrTyp:  testtypes.DynamicType{},
			expected: false,
		},
		"custom-float64": {
			attrTyp:  testtypes.Float64Type{},
			expected: false,
		},
		"custom-int64": {
			attrTyp:  testtypes.Int64Type{},
			expected: false,
		},
		"custom-list-nil": {
			attrTyp: testtypes.ListType{},
			// While testtypes.ListType embeds basetypes.ListType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"custom-map-nil": {
			attrTyp: testtypes.MapType{},
			// While testtypes.MapType embeds basetypes.MapType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"custom-object-nil": {
			attrTyp:  testtypes.ObjectType{},
			expected: false, // expected as objects can be empty
		},
		"custom-number": {
			attrTyp:  testtypes.NumberType{},
			expected: false,
		},
		"custom-set-nil": {
			attrTyp: testtypes.SetType{},
			// While testtypes.SetType embeds basetypes.SetType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"custom-string": {
			attrTyp:  testtypes.StringType{},
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
		"list-custom-bool": {
			attrTyp: types.ListType{
				ElemType: testtypes.BoolType{},
			},
			expected: false,
		},
		"list-custom-dynamic": {
			attrTyp: types.ListType{
				ElemType: testtypes.DynamicType{},
			},
			expected: false,
		},
		"list-custom-float64": {
			attrTyp: types.ListType{
				ElemType: testtypes.Float64Type{},
			},
			expected: false,
		},
		"list-custom-int64": {
			attrTyp: types.ListType{
				ElemType: testtypes.Int64Type{},
			},
			expected: false,
		},
		"list-custom-list-nil": {
			attrTyp: types.ListType{
				ElemType: testtypes.ListType{},
			},
			// While testtypes.ListType embeds basetypes.ListType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"list-custom-list-string": {
			attrTyp: types.ListType{
				ElemType: testtypes.ListType{
					ListType: types.ListType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"list-custom-map-nil": {
			attrTyp: types.ListType{
				ElemType: testtypes.MapType{},
			},
			// While testtypes.MapType embeds basetypes.MapType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"list-custom-map-string": {
			attrTyp: types.ListType{
				ElemType: testtypes.MapType{
					MapType: types.MapType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"list-custom-number": {
			attrTyp: types.ListType{
				ElemType: testtypes.NumberType{},
			},
			expected: false,
		},
		"list-custom-object-nil": {
			attrTyp: types.ListType{
				ElemType: testtypes.ObjectType{},
			},
			expected: false, // expected as objects can be empty
		},
		"list-custom-object-attrtypes": {
			attrTyp: types.ListType{
				ElemType: testtypes.ObjectType{
					ObjectType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool":    types.BoolType,
							"float64": types.Float64Type,
						},
					},
				},
			},
			expected: false,
		},
		"list-custom-object-attrtypes-nil": {
			attrTyp: types.ListType{
				ElemType: testtypes.ObjectType{
					ObjectType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool":    types.BoolType,
							"float64": nil,
						},
					},
				},
			},
			// While testtypes.ObjectType embeds basetypes.ObjectType and this
			// test case specifies a nil AttrTypes value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// AttributeTypes() method which would be used for custom types.
			expected: false,
		},
		"list-custom-set-nil": {
			attrTyp: types.ListType{
				ElemType: testtypes.SetType{},
			},
			// While testtypes.SetType embeds basetypes.SetType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"list-custom-set-string": {
			attrTyp: types.ListType{
				ElemType: testtypes.SetType{
					SetType: types.SetType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"list-custom-string": {
			attrTyp: types.ListType{
				ElemType: testtypes.StringType{},
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
		"map-custom-bool": {
			attrTyp: types.MapType{
				ElemType: testtypes.BoolType{},
			},
			expected: false,
		},
		"map-custom-dynamic": {
			attrTyp: types.MapType{
				ElemType: testtypes.DynamicType{},
			},
			expected: false,
		},
		"map-custom-float64": {
			attrTyp: types.MapType{
				ElemType: testtypes.Float64Type{},
			},
			expected: false,
		},
		"map-custom-int64": {
			attrTyp: types.MapType{
				ElemType: testtypes.Int64Type{},
			},
			expected: false,
		},
		"map-custom-list-nil": {
			attrTyp: types.MapType{
				ElemType: testtypes.ListType{},
			},
			// While testtypes.ListType embeds basetypes.ListType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"map-custom-list-string": {
			attrTyp: types.MapType{
				ElemType: testtypes.ListType{
					ListType: types.ListType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"map-custom-map-nil": {
			attrTyp: types.MapType{
				ElemType: testtypes.MapType{},
			},
			// While testtypes.MapType embeds basetypes.MapType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"map-custom-map-string": {
			attrTyp: types.MapType{
				ElemType: testtypes.MapType{
					MapType: types.MapType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"map-custom-number": {
			attrTyp: types.MapType{
				ElemType: testtypes.NumberType{},
			},
			expected: false,
		},
		"map-custom-object-nil": {
			attrTyp: types.MapType{
				ElemType: testtypes.ObjectType{},
			},
			expected: false, // expected as objects can be empty
		},
		"map-custom-object-attrtypes": {
			attrTyp: types.MapType{
				ElemType: testtypes.ObjectType{
					ObjectType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool":    types.BoolType,
							"float64": types.Float64Type,
						},
					},
				},
			},
			expected: false,
		},
		"map-custom-object-attrtypes-nil": {
			attrTyp: types.MapType{
				ElemType: testtypes.ObjectType{
					ObjectType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool":    types.BoolType,
							"float64": nil,
						},
					},
				},
			},
			// While testtypes.ObjectType embeds basetypes.ObjectType and this
			// test case specifies a nil AttrTypes value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// AttributeTypes() method which would be used for custom types.
			expected: false,
		},
		"map-custom-set-nil": {
			attrTyp: types.MapType{
				ElemType: testtypes.SetType{},
			},
			// While testtypes.SetType embeds basetypes.SetType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"map-custom-set-string": {
			attrTyp: types.MapType{
				ElemType: testtypes.SetType{
					SetType: types.SetType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"map-custom-string": {
			attrTyp: types.MapType{
				ElemType: testtypes.StringType{},
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
		"object-custom-list-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{},
				},
			},
			// While testtypes.ListType embeds basetypes.ListType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"object-custom-list-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-list-list-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.ListType{},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-list-list-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-list-map-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.MapType{},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-list-map-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-list-object-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.ObjectType{},
						},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"object-custom-list-object-attrtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.ObjectType{
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
		"object-custom-list-object-attrtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"bool": types.BoolType,
									"nil":  nil,
								},
							},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-list-set-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.SetType{},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-list-set-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-list-tuple-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.TupleType{},
						},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"object-custom-list-tuple-elemtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.TupleType{
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
		"object-custom-list-tuple-elemtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": testtypes.ListType{
						ListType: types.ListType{
							ElemType: types.TupleType{
								ElemTypes: []attr.Type{
									types.BoolType,
									nil,
								},
							},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-map-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{},
				},
			},
			// While testtypes.MapType embeds basetypes.MapType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"object-custom-map-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-map-list-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.ListType{},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-map-list-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-map-map-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.MapType{},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-map-map-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-map-object-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.ObjectType{},
						},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"object-custom-map-object-attrtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.ObjectType{
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
		"object-custom-map-object-attrtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"bool": types.BoolType,
									"nil":  nil,
								},
							},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-map-set-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.SetType{},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-map-set-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-map-tuple-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.TupleType{},
						},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"object-custom-map-tuple-elemtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.TupleType{
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
		"object-custom-map-tuple-elemtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": testtypes.MapType{
						MapType: types.MapType{
							ElemType: types.TupleType{
								ElemTypes: []attr.Type{
									types.BoolType,
									nil,
								},
							},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-object-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"object": testtypes.ObjectType{},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"object-custom-object-attrtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"object": testtypes.ObjectType{
						ObjectType: types.ObjectType{
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
		"object-custom-object-attrtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"object": testtypes.ObjectType{
						ObjectType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bool": types.BoolType,
								"nil":  nil,
							},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-set-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{},
				},
			},
			// While testtypes.SetType embeds basetypes.SetType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"object-custom-set-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-set-list-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.ListType{},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-set-list-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-set-map-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.MapType{},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-set-map-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-set-object-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.ObjectType{},
						},
					},
				},
			},
			expected: false, // expected as objects can be empty
		},
		"object-custom-set-object-attrtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.ObjectType{
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
		"object-custom-set-object-attrtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"bool": types.BoolType,
									"nil":  nil,
								},
							},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
		},
		"object-custom-set-set-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.SetType{},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-set-set-string": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: false,
		},
		"object-custom-set-tuple-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.TupleType{},
						},
					},
				},
			},
			expected: false, // expected as tuples can be empty
		},
		"object-custom-set-tuple-elemtypes": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.TupleType{
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
		"object-custom-set-tuple-elemtypes-nil": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": testtypes.SetType{
						SetType: types.SetType{
							ElemType: types.TupleType{
								ElemTypes: []attr.Type{
									types.BoolType,
									nil,
								},
							},
						},
					},
				},
			},
			// Similar to other custom type test cases, the custom type will
			// prevent the further checking of the element type.
			expected: false,
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
		"set-custom-bool": {
			attrTyp: types.SetType{
				ElemType: testtypes.BoolType{},
			},
			expected: false,
		},
		"set-custom-dynamic": {
			attrTyp: types.SetType{
				ElemType: testtypes.DynamicType{},
			},
			expected: false,
		},
		"set-custom-float64": {
			attrTyp: types.SetType{
				ElemType: testtypes.Float64Type{},
			},
			expected: false,
		},
		"set-custom-int64": {
			attrTyp: types.SetType{
				ElemType: testtypes.Int64Type{},
			},
			expected: false,
		},
		"set-custom-list-nil": {
			attrTyp: types.SetType{
				ElemType: testtypes.ListType{},
			},
			// While testtypes.ListType embeds basetypes.ListType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"set-custom-list-string": {
			attrTyp: types.SetType{
				ElemType: testtypes.ListType{
					ListType: types.ListType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"set-custom-map-nil": {
			attrTyp: types.SetType{
				ElemType: testtypes.MapType{},
			},
			// While testtypes.MapType embeds basetypes.MapType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"set-custom-map-string": {
			attrTyp: types.SetType{
				ElemType: testtypes.MapType{
					MapType: types.MapType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"set-custom-number": {
			attrTyp: types.SetType{
				ElemType: testtypes.NumberType{},
			},
			expected: false,
		},
		"set-custom-object-nil": {
			attrTyp: types.SetType{
				ElemType: testtypes.ObjectType{},
			},
			expected: false, // expected as objects can be empty
		},
		"set-custom-object-attrtypes": {
			attrTyp: types.SetType{
				ElemType: testtypes.ObjectType{
					ObjectType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool":    types.BoolType,
							"float64": types.Float64Type,
						},
					},
				},
			},
			expected: false,
		},
		"set-custom-object-attrtypes-nil": {
			attrTyp: types.SetType{
				ElemType: testtypes.ObjectType{
					ObjectType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bool":    types.BoolType,
							"float64": nil,
						},
					},
				},
			},
			// While testtypes.ObjectType embeds basetypes.ObjectType and this
			// test case specifies a nil AttrTypes value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// AttributeTypes() method which would be used for custom types.
			expected: false,
		},
		"set-custom-set-nil": {
			attrTyp: types.SetType{
				ElemType: testtypes.SetType{},
			},
			// While testtypes.SetType embeds basetypes.SetType and this test
			// case does not specify an ElemType value, the function logic is
			// coded to only handle basetypes implementations due to the
			// unexported missingType that would be returned from the
			// ElementType() method which would be used for custom types.
			expected: false,
		},
		"set-custom-set-string": {
			attrTyp: types.SetType{
				ElemType: testtypes.SetType{
					SetType: types.SetType{
						ElemType: types.StringType,
					},
				},
			},
			expected: false,
		},
		"set-custom-string": {
			attrTyp: types.SetType{
				ElemType: testtypes.StringType{},
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
