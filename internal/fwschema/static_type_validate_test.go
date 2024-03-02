// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschema_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateStaticStructuralType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attrTyp  attr.Type
		expected diag.Diagnostic
	}{
		"nil": {
			attrTyp: nil,
		},
		"dynamic": {
			attrTyp: types.DynamicType,
		},
		"primitive": {
			attrTyp: types.StringType,
		},
		"obj-list-missing": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{},
				},
			},
		},
		"obj-list-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.StringType,
					},
				},
			},
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
		},
		"obj-list-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.DynamicType,
					},
				},
			},
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
		},
		"obj-map-missing": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{},
				},
			},
		},
		"obj-map-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.StringType,
					},
				},
			},
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
		},
		"obj-map-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.DynamicType,
					},
				},
			},
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
		},
		"obj-set-missing": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{},
				},
			},
		},
		"obj-set-static": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.StringType,
					},
				},
			},
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
		},
		"obj-set-dynamic": {
			attrTyp: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.DynamicType,
					},
				},
			},
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwschema.ValidateStaticStructuralType(path.Root("test"), testCase.attrTyp)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
func TestValidateStaticCollectionType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attrTyp  attr.Type
		expected diag.Diagnostic
	}{
		"nil": {
			attrTyp: nil,
		},
		"primitive": {
			attrTyp: types.StringType,
		},
		"list-missing": {
			attrTyp: types.ListType{},
		},
		"list-static": {
			attrTyp: types.ListType{
				ElemType: types.StringType,
			},
		},
		"list-list-static": {
			attrTyp: types.ListType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
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
		},
		"list-dynamic": {
			attrTyp: types.ListType{
				ElemType: types.DynamicType,
			},
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
		},
		"list-list-dynamic": {
			attrTyp: types.ListType{
				ElemType: types.ListType{
					ElemType: types.DynamicType,
				},
			},
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
		},
		"map-missing": {
			attrTyp: types.MapType{},
		},
		"map-static": {
			attrTyp: types.MapType{
				ElemType: types.StringType,
			},
		},
		"map-map-static": {
			attrTyp: types.MapType{
				ElemType: types.MapType{
					ElemType: types.StringType,
				},
			},
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
		},
		"map-dynamic": {
			attrTyp: types.MapType{
				ElemType: types.DynamicType,
			},
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
		},
		"map-map-dynamic": {
			attrTyp: types.MapType{
				ElemType: types.MapType{
					ElemType: types.DynamicType,
				},
			},
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
		},
		"set-missing": {
			attrTyp: types.SetType{},
		},
		"set-static": {
			attrTyp: types.SetType{
				ElemType: types.StringType,
			},
		},
		"set-set-static": {
			attrTyp: types.SetType{
				ElemType: types.SetType{
					ElemType: types.StringType,
				},
			},
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
		},
		"set-dynamic": {
			attrTyp: types.SetType{
				ElemType: types.DynamicType,
			},
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
		},
		"set-set-dynamic": {
			attrTyp: types.SetType{
				ElemType: types.SetType{
					ElemType: types.DynamicType,
				},
			},
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
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
			expected: diag.NewErrorDiagnostic(
				"Invalid Schema Implementation",
				"When validating the schema, an implementation issue was found. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"\"test\" is an attribute that contains a collection type with a nested dynamic type. "+
					"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
			),
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwschema.ValidateStaticCollectionType(path.Root("test"), testCase.attrTyp)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
