// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type EmbedSingleField struct {
	Attr1 types.String `tfsdk:"attr_1"`
}

type embedSingleField struct {
	Attr1 types.String `tfsdk:"attr_1"`
}

type embedNoFields struct{}

type embedWithUntaggedExportedField struct {
	ExportedAndUntagged types.String
}

type embedWithInvalidTag struct {
	A types.String `tfsdk:"invalidTag"`
}

func TestNewStruct_errors(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.Type
		objVal        tftypes.Value
		targetVal     reflect.Value
		expectedError error
		expectedDiags diag.Diagnostics
	}{
		"not-an-object": {
			typ:           types.StringType,
			objVal:        tftypes.NewValue(tftypes.String, "hello"),
			targetVal:     reflect.ValueOf(struct{}{}),
			expectedError: fmt.Errorf("cannot reflect %s into a struct, must be an object", tftypes.String),
		},
		"not-a-struct": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal:     reflect.ValueOf(""),
			expectedError: errors.New("expected a struct type, got string"),
		},
		"object-missing-fields": {
			typ:    types.ObjectType{},
			objVal: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
			targetVal: reflect.ValueOf(struct {
				A string `tfsdk:"a"`
			}{}),
			expectedError: errors.New("mismatch between struct and object: Struct defines fields not found in object: a."),
		},
		"embedded-object-missing-fields": {
			typ:    types.ObjectType{},
			objVal: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
			targetVal: reflect.ValueOf(struct {
				embedSingleField
			}{}),
			expectedError: errors.New("mismatch between struct and object: Struct defines fields not found in object: attr_1."),
		},
		"struct-missing-fields": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal:     reflect.ValueOf(struct{}{}),
			expectedError: errors.New("mismatch between struct and object: Object defines fields not found in struct: a."),
		},
		"embedded-struct-missing-fields": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				embedNoFields
			}{}),
			expectedError: errors.New("mismatch between struct and object: Object defines fields not found in struct: a."),
		},
		"object-and-struct-missing-fields": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"b": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"b": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"b": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				A string `tfsdk:"a"`
			}{}),
			expectedError: errors.New("mismatch between struct and object: Struct defines fields not found in object: a. Object defines fields not found in struct: b."),
		},
		"embedded-object-and-struct-missing-fields": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"b": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"b": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"b": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				embedSingleField
			}{}),
			expectedError: errors.New("mismatch between struct and object: Struct defines fields not found in object: attr_1. Object defines fields not found in struct: b."),
		},
		"struct-has-untagged-fields": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				A                   string `tfsdk:"a"`
				ExportedAndUntagged string
			}{}),
			expectedError: fmt.Errorf(
				"error retrieving field names from struct tags: %w",
				errors.New(`: need a struct tag for "tfsdk" on ExportedAndUntagged`)),
		},
		"embedded-struct-has-untagged-fields": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				A string `tfsdk:"a"`
				embedWithUntaggedExportedField
			}{}),
			expectedError: fmt.Errorf(
				"error retrieving field names from struct tags: %w",
				fmt.Errorf(`error retrieving embedded struct "embedWithUntaggedExportedField" field tags: %w`,
					errors.New(`: need a struct tag for "tfsdk" on ExportedAndUntagged`),
				),
			),
		},
		"struct-has-invalid-tags": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				A string `tfsdk:"invalidTag"`
			}{}),
			expectedError: fmt.Errorf(
				"error retrieving field names from struct tags: %w",
				errors.New("invalidTag: invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter")),
		},
		"struct-has-empty-tag": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				A string `tfsdk:""`
			}{}),
			expectedError: fmt.Errorf(
				"error retrieving field names from struct tags: %w",
				errors.New(": invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter")),
		},
		"embedded-struct-has-invalid-tags": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				embedWithInvalidTag
			}{}),
			expectedError: fmt.Errorf(
				"error retrieving field names from struct tags: %w",
				fmt.Errorf(`error retrieving embedded struct "embedWithInvalidTag" field tags: %w`,
					errors.New("invalidTag: invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter"),
				),
			),
		},
		"struct-has-duplicate-tags": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				A string `tfsdk:"a"`
				B string `tfsdk:"a"`
			}{}),
			expectedError: fmt.Errorf(
				"error retrieving field names from struct tags: %w",
				errors.New(`a: can't use tfsdk tag "a" for both A and B fields`)),
		},
		"embedded-struct-has-duplicate-tags": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				FirstAttr1 types.String `tfsdk:"attr_1"`
				// Also contains an attr_1 tfsdk tag
				embedSingleField
			}{}),
			expectedError: fmt.Errorf(
				"error retrieving field names from struct tags: %w",
				errors.New(`embedded struct "embedSingleField" promotes a field with a duplicate tfsdk tag "attr_1", conflicts with "FirstAttr1" tfsdk tag`),
			),
		},
		"embedded-exported-struct-with-pointer-not-supported": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attr_1": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attr_1": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				*EmbedSingleField
			}{}),
			expectedError: errors.New(
				"struct { *reflect_test.EmbedSingleField } contains a struct embedded by a pointer which is not supported. Switch any embedded structs to be embedded by value.\n\n" +
					"Error: reflect: indirection through nil pointer to embedded struct field EmbedSingleField",
			),
		},
		"embedded-unexported-struct-with-pointer-not-supported": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attr_1": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attr_1": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				*embedSingleField
			}{}),
			expectedError: errors.New(
				"struct { *reflect_test.embedSingleField } contains a struct embedded by a pointer which is not supported. Switch any embedded structs to be embedded by value.\n\n" +
					"Error: reflect: indirection through nil pointer to embedded struct field embedSingleField",
			),
		},
		"embedded-struct-has-empty-tfsdk-tag": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attr_1": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attr_1": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				embedSingleField `tfsdk:""`
			}{}),
			expectedError: errors.New(
				"error retrieving field names from struct tags: : embedded struct field embedSingleField cannot have tfsdk tag",
			),
		},
		"embedded-struct-has-tfsdk-tag": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.StringType,
				},
			},
			objVal: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attr_1": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attr_1": tftypes.NewValue(tftypes.String, "hello"),
			}),
			targetVal: reflect.ValueOf(struct {
				embedSingleField `tfsdk:"attr_2"`
			}{}),
			expectedError: errors.New(
				"error retrieving field names from struct tags: attr_2: embedded struct field embedSingleField cannot have tfsdk tag",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			expectedDiags := diag.Diagnostics{
				diag.WithPath(path.Empty(), refl.DiagIntoIncompatibleType{
					Err:        testCase.expectedError,
					TargetType: testCase.targetVal.Type(),
					Val:        testCase.objVal,
				}),
			}

			_, diags := refl.Struct(context.Background(), testCase.typ, testCase.objVal, testCase.targetVal, refl.Options{}, path.Empty())

			if diff := cmp.Diff(diags, expectedDiags); diff != "" {
				for _, d := range diags {
					t.Logf("%s: %s\n%s\n", d.Severity(), d.Summary(), d.Detail())
				}
				t.Errorf("unexpected diagnostics: %s", diff)
			}
		})
	}
}

func TestNewStruct_primitives(t *testing.T) {
	t.Parallel()

	var s struct {
		A string     `tfsdk:"a"`
		B *big.Float `tfsdk:"b"`
		C bool       `tfsdk:"c"`
	}
	result, diags := refl.Struct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"a": types.StringType,
			"b": types.NumberType,
			"c": types.BoolType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
			"b": tftypes.Number,
			"c": tftypes.Bool,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
		"b": tftypes.NewValue(tftypes.Number, 123),
		"c": tftypes.NewValue(tftypes.Bool, true),
	}), reflect.ValueOf(s), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&s).Elem().Set(result)
	if s.A != "hello" {
		t.Errorf("Expected s.A to be %q, was %q", "hello", s.A)
	}
	if s.B.Cmp(big.NewFloat(123)) != 0 {
		t.Errorf("Expected s.B to be %v, was %v", big.NewFloat(123), s.B)
	}
	if s.C != true {
		t.Errorf("Expected s.C to be %v, was %v", true, s.C)
	}
}

func TestNewStruct_embedded_primitives(t *testing.T) {
	t.Parallel()

	type S3 struct {
		C bool `tfsdk:"c"`
	}
	type s2 struct {
		B *big.Float `tfsdk:"b"`
		S3
	}
	type s1 struct {
		A string `tfsdk:"a"`
		s2
	}
	var s s1

	result, diags := refl.Struct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"a": types.StringType,
			"b": types.NumberType,
			"c": types.BoolType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
			"b": tftypes.Number,
			"c": tftypes.Bool,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
		"b": tftypes.NewValue(tftypes.Number, 123),
		"c": tftypes.NewValue(tftypes.Bool, true),
	}), reflect.ValueOf(s), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&s).Elem().Set(result)
	if s.A != "hello" {
		t.Errorf("Expected s.A to be %q, was %q", "hello", s.A)
	}
	if s.B.Cmp(big.NewFloat(123)) != 0 {
		t.Errorf("Expected s.B to be %v, was %v", big.NewFloat(123), s.B)
	}
	if s.C != true {
		t.Errorf("Expected s.C to be %v, was %v", true, s.C)
	}
}

func TestNewStruct_complex(t *testing.T) {
	t.Parallel()

	type myStruct struct {
		ListSlice          []string `tfsdk:"list_slice"`
		ListSliceOfStructs []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		} `tfsdk:"list_slice_of_structs"`
		SetSlice          []string `tfsdk:"set_slice"`
		SetSliceOfStructs []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		} `tfsdk:"set_slice_of_structs"`
		Struct struct {
			A     bool      `tfsdk:"a"`
			Slice []float64 `tfsdk:"slice"`
		} `tfsdk:"struct"`
		Map              map[string][]string `tfsdk:"map"`
		Pointer          *string             `tfsdk:"pointer"`
		Unknownable      *unknownableString  `tfsdk:"unknownable"`
		Nullable         *nullableString     `tfsdk:"nullable"`
		AttributeValue   types.String        `tfsdk:"attribute_value"`
		ValueConverter   *valueConverter     `tfsdk:"value_converter"`
		UnhandledNull    string              `tfsdk:"unhandled_null"`
		UnhandledUnknown string              `tfsdk:"unhandled_unknown"`
	}
	var s myStruct
	result, diags := refl.Struct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"list_slice": types.ListType{
				ElemType: types.StringType,
			},
			"list_slice_of_structs": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"a": types.StringType,
						"b": types.NumberType,
					},
				},
			},
			"set_slice": types.SetType{
				ElemType: types.StringType,
			},
			"set_slice_of_structs": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"a": types.StringType,
						"b": types.NumberType,
					},
				},
			},
			"struct": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.BoolType,
					"slice": types.ListType{
						ElemType: types.NumberType,
					},
				},
			},
			"map": types.MapType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
			"pointer":           types.StringType,
			"unknownable":       types.StringType,
			"nullable":          types.StringType,
			"attribute_value":   types.StringType,
			"value_converter":   types.StringType,
			"unhandled_null":    types.StringType,
			"unhandled_unknown": types.StringType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"list_slice": tftypes.List{
				ElementType: tftypes.String,
			},
			"list_slice_of_structs": tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"a": tftypes.String,
						"b": tftypes.Number,
					},
				},
			},
			"set_slice": tftypes.Set{
				ElementType: tftypes.String,
			},
			"set_slice_of_structs": tftypes.Set{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"a": tftypes.String,
						"b": tftypes.Number,
					},
				},
			},
			"struct": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.Bool,
					"slice": tftypes.List{
						ElementType: tftypes.Number,
					},
				},
			},
			"map": tftypes.Map{
				ElementType: tftypes.List{
					ElementType: tftypes.String,
				},
			},
			"pointer":           tftypes.String,
			"unknownable":       tftypes.String,
			"nullable":          tftypes.String,
			"attribute_value":   tftypes.String,
			"value_converter":   tftypes.String,
			"unhandled_null":    tftypes.String,
			"unhandled_unknown": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"list_slice": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.String,
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "red"),
			tftypes.NewValue(tftypes.String, "blue"),
			tftypes.NewValue(tftypes.String, "green"),
		}),
		"list_slice_of_structs": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Number,
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello, world"),
				"b": tftypes.NewValue(tftypes.Number, 123),
			}),
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "goodnight, moon"),
				"b": tftypes.NewValue(tftypes.Number, 456),
			}),
		}),
		"set_slice": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.String,
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "red"),
			tftypes.NewValue(tftypes.String, "blue"),
			tftypes.NewValue(tftypes.String, "green"),
		}),
		"set_slice_of_structs": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Number,
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "hello, world"),
				"b": tftypes.NewValue(tftypes.Number, 123),
			}),
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "goodnight, moon"),
				"b": tftypes.NewValue(tftypes.Number, 456),
			}),
		}),
		"struct": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"a": tftypes.Bool,
				"slice": tftypes.List{
					ElementType: tftypes.Number,
				},
			},
		}, map[string]tftypes.Value{
			"a": tftypes.NewValue(tftypes.Bool, true),
			"slice": tftypes.NewValue(tftypes.List{
				ElementType: tftypes.Number,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Number, 123),
				tftypes.NewValue(tftypes.Number, 456),
				tftypes.NewValue(tftypes.Number, 789),
			}),
		}),
		"map": tftypes.NewValue(tftypes.Map{
			ElementType: tftypes.List{
				ElementType: tftypes.String,
			},
		}, map[string]tftypes.Value{
			"colors": tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "red"),
				tftypes.NewValue(tftypes.String, "orange"),
				tftypes.NewValue(tftypes.String, "yellow"),
			}),
			"fruits": tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "apple"),
				tftypes.NewValue(tftypes.String, "banana"),
			}),
		}),
		"pointer":           tftypes.NewValue(tftypes.String, "pointed"),
		"unknownable":       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"nullable":          tftypes.NewValue(tftypes.String, nil),
		"attribute_value":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_converter":   tftypes.NewValue(tftypes.String, nil),
		"unhandled_null":    tftypes.NewValue(tftypes.String, nil),
		"unhandled_unknown": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	}), reflect.ValueOf(s), refl.Options{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	}, path.Empty())
	reflect.ValueOf(&s).Elem().Set(result)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	str := "pointed"
	expected := myStruct{
		ListSlice: []string{"red", "blue", "green"},
		ListSliceOfStructs: []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		}{
			{
				A: "hello, world",
				B: 123,
			},
			{
				A: "goodnight, moon",
				B: 456,
			},
		},
		SetSlice: []string{"red", "blue", "green"},
		SetSliceOfStructs: []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		}{
			{
				A: "hello, world",
				B: 123,
			},
			{
				A: "goodnight, moon",
				B: 456,
			},
		},
		Struct: struct {
			A     bool      `tfsdk:"a"`
			Slice []float64 `tfsdk:"slice"`
		}{
			A:     true,
			Slice: []float64{123, 456, 789},
		},
		Map: map[string][]string{
			"colors": {"red", "orange", "yellow"},
			"fruits": {"apple", "banana"},
		},
		Pointer: &str,
		Unknownable: &unknownableString{
			Unknown: true,
		},
		Nullable: &nullableString{
			Null: true,
		},
		AttributeValue: types.StringUnknown(),
		ValueConverter: &valueConverter{
			null: true,
		},
		UnhandledNull:    "",
		UnhandledUnknown: "",
	}
	if diff := cmp.Diff(s, expected); diff != "" {
		t.Errorf("Didn't get expected value. Diff (+ is expected, - is result): %s", diff)
	}
}

func TestNewStruct_structtags_ignores(t *testing.T) {
	t.Parallel()

	var s struct {
		ExportedAndTagged   string `tfsdk:"exported_and_tagged"`
		unexported          string //nolint:structcheck,unused
		unexportedAndTagged string `tfsdk:"unexported_and_tagged"`
		ExportedAndExcluded string `tfsdk:"-"`
	}
	result, diags := refl.Struct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"exported_and_tagged": types.StringType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"exported_and_tagged": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"exported_and_tagged": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(s), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&s).Elem().Set(result)
	if s.ExportedAndTagged != "hello" {
		t.Errorf("Expected s.ExportedAndTagged to be %q, was %q", "hello", s.ExportedAndTagged)
	}

	if s.unexported != "" {
		t.Errorf("Expected s.unexported to be empty, was %q", s.unexported)
	}

	if s.unexportedAndTagged != "" {
		t.Errorf("Expected s.unexportedAndTagged to be empty, was %q", s.unexportedAndTagged)
	}

	if s.ExportedAndExcluded != "" {
		t.Errorf("Expected s.ExportedAndExcluded to be empty, was %q", s.ExportedAndExcluded)
	}
}

func TestNewStruct_embedded_structtags_ignores(t *testing.T) {
	t.Parallel()

	type s1 struct {
		ExportedAndTagged   string `tfsdk:"exported_and_tagged"`
		unexported          string //nolint:structcheck,unused
		unexportedAndTagged string `tfsdk:"unexported_and_tagged"`
		ExportedAndExcluded string `tfsdk:"-"`
	}

	type s2 struct {
		unexportedField string //nolint:structcheck,unused
	}

	var s struct {
		s1
		*s2 `tfsdk:"-"`
	}
	result, diags := refl.Struct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"exported_and_tagged": types.StringType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"exported_and_tagged": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"exported_and_tagged": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(s), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&s).Elem().Set(result)
	if s.ExportedAndTagged != "hello" {
		t.Errorf("Expected s.ExportedAndTagged to be %q, was %q", "hello", s.ExportedAndTagged)
	}

	if s.unexported != "" {
		t.Errorf("Expected s.unexported to be empty, was %q", s.unexported)
	}

	if s.unexportedAndTagged != "" {
		t.Errorf("Expected s.unexportedAndTagged to be empty, was %q", s.unexportedAndTagged)
	}

	if s.ExportedAndExcluded != "" {
		t.Errorf("Expected s.ExportedAndExcluded to be empty, was %q", s.ExportedAndExcluded)
	}
}

func TestFromStruct_primitives(t *testing.T) {
	t.Parallel()

	type disk struct {
		Name    string `tfsdk:"name"`
		Age     int    `tfsdk:"age"`
		OptedIn bool   `tfsdk:"opted_in"`
	}
	disk1 := disk{
		Name:    "myfirstdisk",
		Age:     30,
		OptedIn: true,
	}

	actualVal, diags := refl.FromStruct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":     types.StringType,
			"age":      types.NumberType,
			"opted_in": types.BoolType,
		},
	}, reflect.ValueOf(disk1), path.Empty())
	if diags.HasError() {
		t.Fatalf("Unexpected error: %v", diags)
	}

	expectedVal := types.ObjectValueMust(
		map[string]attr.Type{
			"name":     types.StringType,
			"age":      types.NumberType,
			"opted_in": types.BoolType,
		},
		map[string]attr.Value{
			"name":     types.StringValue("myfirstdisk"),
			"age":      types.NumberValue(big.NewFloat(30)),
			"opted_in": types.BoolValue(true),
		},
	)

	if diff := cmp.Diff(expectedVal, actualVal); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromStruct_embedded_primitives(t *testing.T) {
	t.Parallel()

	type embedFields struct {
		Name string `tfsdk:"name"`
		Age  int    `tfsdk:"age"`
	}

	type EmbedFields struct {
		OptedIn bool `tfsdk:"opted_in"`
	}

	type disk struct {
		embedFields
		EmbedFields
	}
	disk1 := disk{
		embedFields{
			Name: "myfirstdisk",
			Age:  30,
		},
		EmbedFields{
			OptedIn: true,
		},
	}

	actualVal, diags := refl.FromStruct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":     types.StringType,
			"age":      types.NumberType,
			"opted_in": types.BoolType,
		},
	}, reflect.ValueOf(disk1), path.Empty())
	if diags.HasError() {
		t.Fatalf("Unexpected error: %v", diags)
	}

	expectedVal := types.ObjectValueMust(
		map[string]attr.Type{
			"name":     types.StringType,
			"age":      types.NumberType,
			"opted_in": types.BoolType,
		},
		map[string]attr.Value{
			"name":     types.StringValue("myfirstdisk"),
			"age":      types.NumberValue(big.NewFloat(30)),
			"opted_in": types.BoolValue(true),
		},
	)

	if diff := cmp.Diff(expectedVal, actualVal); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromStruct_complex(t *testing.T) {
	t.Parallel()

	type myStruct struct {
		ListSlice          []string `tfsdk:"list_slice"`
		ListSliceOfStructs []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		} `tfsdk:"list_slice_of_structs"`
		SetSlice          []string `tfsdk:"set_slice"`
		SetSliceOfStructs []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		} `tfsdk:"set_slice_of_structs"`
		Struct struct {
			A     bool      `tfsdk:"a"`
			Slice []float64 `tfsdk:"slice"`
		} `tfsdk:"struct"`
		Map            map[string][]string `tfsdk:"map"`
		Pointer        *string             `tfsdk:"pointer"`
		Unknownable    *unknownableString  `tfsdk:"unknownable"`
		Nullable       *nullableString     `tfsdk:"nullable"`
		AttributeValue types.String        `tfsdk:"attribute_value"`
		ValueCreator   *valueCreator       `tfsdk:"value_creator"`
		BigFloat       *big.Float          `tfsdk:"big_float"`
		BigInt         *big.Int            `tfsdk:"big_int"`
		Uint           uint64              `tfsdk:"uint"`
	}
	str := "pointed"
	s := myStruct{
		ListSlice: []string{"red", "blue", "green"},
		ListSliceOfStructs: []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		}{
			{
				A: "hello, world",
				B: 123,
			},
			{
				A: "goodnight, moon",
				B: 456,
			},
		},
		SetSlice: []string{"red", "blue", "green"},
		SetSliceOfStructs: []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		}{
			{
				A: "hello, world",
				B: 123,
			},
			{
				A: "goodnight, moon",
				B: 456,
			},
		},
		Struct: struct {
			A     bool      `tfsdk:"a"`
			Slice []float64 `tfsdk:"slice"`
		}{
			A:     true,
			Slice: []float64{123, 456, 789},
		},
		Map: map[string][]string{
			"colors": {"red", "orange", "yellow"},
			"fruits": {"apple", "banana"},
		},
		Pointer: &str,
		Unknownable: &unknownableString{
			Unknown: true,
		},
		Nullable: &nullableString{
			Null: true,
		},
		AttributeValue: types.StringUnknown(),
		ValueCreator: &valueCreator{
			null: true,
		},
		BigFloat: big.NewFloat(123.456),
		BigInt:   big.NewInt(123456),
		Uint:     123456,
	}
	result, diags := refl.FromStruct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"list_slice": types.ListType{
				ElemType: types.StringType,
			},
			"list_slice_of_structs": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"a": types.StringType,
						"b": types.NumberType,
					},
				},
			},
			"set_slice": types.SetType{
				ElemType: types.StringType,
			},
			"set_slice_of_structs": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"a": types.StringType,
						"b": types.NumberType,
					},
				},
			},
			"struct": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.BoolType,
					"slice": types.ListType{
						ElemType: types.NumberType,
					},
				},
			},
			"map": types.MapType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
			"pointer":         types.StringType,
			"unknownable":     types.StringType,
			"nullable":        types.StringType,
			"attribute_value": types.StringType,
			"value_creator":   types.StringType,
			"big_float":       types.NumberType,
			"big_int":         types.NumberType,
			"uint":            types.NumberType,
		},
	}, reflect.ValueOf(s), path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	expected := types.ObjectValueMust(
		map[string]attr.Type{
			"list_slice": types.ListType{
				ElemType: types.StringType,
			},
			"list_slice_of_structs": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"a": types.StringType,
						"b": types.NumberType,
					},
				},
			},
			"set_slice": types.SetType{
				ElemType: types.StringType,
			},
			"set_slice_of_structs": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"a": types.StringType,
						"b": types.NumberType,
					},
				},
			},
			"struct": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": types.BoolType,
					"slice": types.ListType{
						ElemType: types.NumberType,
					},
				},
			},
			"map": types.MapType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
			"pointer":         types.StringType,
			"unknownable":     types.StringType,
			"nullable":        types.StringType,
			"attribute_value": types.StringType,
			"value_creator":   types.StringType,
			"big_float":       types.NumberType,
			"big_int":         types.NumberType,
			"uint":            types.NumberType,
		},
		map[string]attr.Value{
			"list_slice": types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("red"),
					types.StringValue("blue"),
					types.StringValue("green"),
				},
			),
			"list_slice_of_structs": types.ListValueMust(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"a": types.StringType,
						"b": types.NumberType,
					},
				},
				[]attr.Value{
					types.ObjectValueMust(
						map[string]attr.Type{
							"a": types.StringType,
							"b": types.NumberType,
						},
						map[string]attr.Value{
							"a": types.StringValue("hello, world"),
							"b": types.NumberValue(big.NewFloat(123)),
						},
					),
					types.ObjectValueMust(
						map[string]attr.Type{
							"a": types.StringType,
							"b": types.NumberType,
						},
						map[string]attr.Value{
							"a": types.StringValue("goodnight, moon"),
							"b": types.NumberValue(big.NewFloat(456)),
						},
					),
				},
			),
			"set_slice": types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("red"),
					types.StringValue("blue"),
					types.StringValue("green"),
				},
			),
			"set_slice_of_structs": types.SetValueMust(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"a": types.StringType,
						"b": types.NumberType,
					},
				},
				[]attr.Value{
					types.ObjectValueMust(
						map[string]attr.Type{
							"a": types.StringType,
							"b": types.NumberType,
						},
						map[string]attr.Value{
							"a": types.StringValue("hello, world"),
							"b": types.NumberValue(big.NewFloat(123)),
						},
					),
					types.ObjectValueMust(
						map[string]attr.Type{
							"a": types.StringType,
							"b": types.NumberType,
						},
						map[string]attr.Value{
							"a": types.StringValue("goodnight, moon"),
							"b": types.NumberValue(big.NewFloat(456)),
						},
					),
				},
			),
			"struct": types.ObjectValueMust(
				map[string]attr.Type{
					"a": types.BoolType,
					"slice": types.ListType{
						ElemType: types.NumberType,
					},
				},
				map[string]attr.Value{
					"a": types.BoolValue(true),
					"slice": types.ListValueMust(
						types.NumberType,
						[]attr.Value{
							types.NumberValue(big.NewFloat(123)),
							types.NumberValue(big.NewFloat(456)),
							types.NumberValue(big.NewFloat(789)),
						},
					),
				},
			),
			"map": types.MapValueMust(
				types.ListType{
					ElemType: types.StringType,
				},
				map[string]attr.Value{
					"colors": types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("red"),
							types.StringValue("orange"),
							types.StringValue("yellow"),
						},
					),
					"fruits": types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("apple"),
							types.StringValue("banana"),
						},
					),
				},
			),
			"pointer":         types.StringValue("pointed"),
			"unknownable":     types.StringUnknown(),
			"nullable":        types.StringNull(),
			"attribute_value": types.StringUnknown(),
			"value_creator":   types.StringNull(),
			"big_float":       types.NumberValue(big.NewFloat(123.456)),
			"big_int":         types.NumberValue(big.NewFloat(123456)),
			"uint":            types.NumberValue(big.NewFloat(123456)),
		},
	)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Didn't get expected value. Diff (+ is expected, - is result): %s", diff)
	}
}

func TestFromStruct_errors(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.TypeWithAttributeTypes
		val           reflect.Value
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"not-a-struct": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": types.StringType,
				},
			},
			val: reflect.ValueOf("not-a-struct"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: test: can't get struct tags of string, is not a struct",
				),
			},
		},
		"struct-field-mismatch": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					NotTest types.String `tfsdk:"not_test"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct into an object. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Mismatch between struct and object type: Struct defines fields not found in object: not_test. Object defines fields not found in struct: test.\n"+
						`Struct: struct { NotTest basetypes.StringValue "tfsdk:\"not_test\"" }`+"\n"+
						`Object type: types.ObjectType["test":basetypes.StringType]`,
				),
			},
		},
		"embedded-struct-field-mismatch": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					embedSingleField
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct into an object. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Mismatch between struct and object type: Struct defines fields not found in object: attr_1. Object defines fields not found in struct: test.\n"+
						`Struct: struct { reflect_test.embedSingleField }`+"\n"+
						`Object type: types.ObjectType["test":basetypes.StringType]`,
				),
			},
		},
		"struct-type-mismatch": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"string": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.Bool `tfsdk:"string"` // intentionally not types.String
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("string"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: basetypes.StringType / underlying type: tftypes.String\n"+
						"Received framework type from provider logic: basetypes.BoolType / underlying type: tftypes.Bool\n"+
						"Path: test.string",
				),
			},
		},
		"embedded-struct-type-mismatch": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.BoolType,
				},
			},
			val: reflect.ValueOf(
				struct {
					embedSingleField // intentionally not types.Bool
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("attr_1"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: basetypes.BoolType / underlying type: tftypes.Bool\n"+
						"Received framework type from provider logic: basetypes.StringType / underlying type: tftypes.String\n"+
						"Path: test.attr_1",
				),
			},
		},
		"struct-validate-error": {
			typ: testtypes.ObjectTypeWithValidateError{
				ObjectType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test": types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.String `tfsdk:"test"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Error Diagnostic",
					"This is an error.",
				),
			},
		},
		"embedded-struct-validate-error": {
			typ: testtypes.ObjectTypeWithValidateError{
				ObjectType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attr_1": types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					embedSingleField
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Error Diagnostic",
					"This is an error.",
				),
			},
		},
		"struct-validate-attribute-error": {
			typ: testtypes.ObjectTypeWithValidateAttributeError{
				ObjectType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test": types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.String `tfsdk:"test"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Error Diagnostic",
					"This is an error.",
				),
			},
		},
		"embedded-struct-validate-attribute-error": {
			typ: testtypes.ObjectTypeWithValidateAttributeError{
				ObjectType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attr_1": types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					embedSingleField
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Error Diagnostic",
					"This is an error.",
				),
			},
		},
		"struct-validate-warning": {
			typ: testtypes.ObjectTypeWithValidateWarning{
				ObjectType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test": types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.String `tfsdk:"test"`
				}{
					Test: types.StringValue("test"),
				},
			),
			expected: types.ObjectValueMust(
				map[string]attr.Type{
					"test": types.StringType,
				},
				map[string]attr.Value{
					"test": types.StringValue("test"),
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(
					path.Root("test"),
					"Warning Diagnostic",
					"This is a warning.",
				),
			},
		},
		"embedded-struct-validate-warning": {
			typ: testtypes.ObjectTypeWithValidateWarning{
				ObjectType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attr_1": types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					embedSingleField
				}{
					embedSingleField{
						Attr1: types.StringValue("test"),
					},
				},
			),
			expected: types.ObjectValueMust(
				map[string]attr.Type{
					"attr_1": types.StringType,
				},
				map[string]attr.Value{
					"attr_1": types.StringValue("test"),
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(
					path.Root("test"),
					"Warning Diagnostic",
					"This is a warning.",
				),
			},
		},
		"struct-validate-attribute-warning": {
			typ: testtypes.ObjectTypeWithValidateAttributeWarning{
				ObjectType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test": types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.String `tfsdk:"test"`
				}{
					Test: types.StringValue("test"),
				},
			),
			expected: testtypes.ObjectValueWithValidateAttributeWarning{
				Object: types.ObjectValueMust(
					map[string]attr.Type{
						"test": types.StringType,
					},
					map[string]attr.Value{
						"test": types.StringValue("test"),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(
					path.Root("test"),
					"Warning Diagnostic",
					"This is a warning.",
				),
			},
		},
		"embedded-struct-validate-attribute-warning": {
			typ: testtypes.ObjectTypeWithValidateAttributeWarning{
				ObjectType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attr_1": types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					embedSingleField
				}{
					embedSingleField: embedSingleField{
						Attr1: types.StringValue("test"),
					},
				},
			),
			expected: testtypes.ObjectValueWithValidateAttributeWarning{
				Object: types.ObjectValueMust(
					map[string]attr.Type{
						"attr_1": types.StringType,
					},
					map[string]attr.Value{
						"attr_1": types.StringValue("test"),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(
					path.Root("test"),
					"Warning Diagnostic",
					"This is a warning.",
				),
			},
		},
		"struct-field-validate-error": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": testtypes.StringTypeWithValidateError{},
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.String `tfsdk:"test"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("test"),
					"Error Diagnostic",
					"This is an error.",
				),
			},
		},
		"embedded-struct-field-validate-error": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": testtypes.StringTypeWithValidateError{},
				},
			},
			val: reflect.ValueOf(
				struct {
					embedSingleField
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("attr_1"),
					"Error Diagnostic",
					"This is an error.",
				),
			},
		},
		"struct-field-validate-attribute-error": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": testtypes.StringTypeWithValidateAttributeWarning{},
				},
			},
			val: reflect.ValueOf(
				struct {
					Test testtypes.StringValueWithValidateAttributeError `tfsdk:"test"`
				}{
					Test: testtypes.StringValueWithValidateAttributeError{
						InternalString: testtypes.String{
							InternalString: types.String{},
							CreatedBy:      testtypes.StringTypeWithValidateAttributeError{},
						},
					},
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("test"),
					"Error Diagnostic",
					"This is an error.",
				),
			},
		},
		"struct-field-validate-warning": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": testtypes.StringTypeWithValidateWarning{},
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.String `tfsdk:"test"`
				}{
					Test: types.StringValue("test"),
				},
			),
			expected: types.ObjectValueMust(
				map[string]attr.Type{
					"test": testtypes.StringTypeWithValidateWarning{},
				},
				map[string]attr.Value{
					"test": testtypes.String{
						InternalString: types.StringValue("test"),
						CreatedBy:      testtypes.StringTypeWithValidateWarning{},
					},
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(
					path.Root("test").AtName("test"),
					"Warning Diagnostic",
					"This is a warning.",
				),
			},
		},
		"struct-field-validate-attribute-warning": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": testtypes.StringTypeWithValidateAttributeWarning{},
				},
			},
			val: reflect.ValueOf(
				struct {
					Test testtypes.StringValueWithValidateAttributeWarning `tfsdk:"test"`
				}{
					Test: testtypes.StringValueWithValidateAttributeWarning{
						InternalString: testtypes.String{
							InternalString: types.StringValue("test"),
							CreatedBy:      testtypes.StringTypeWithValidateAttributeWarning{},
						},
					},
				},
			),
			expected: types.ObjectValueMust(
				map[string]attr.Type{
					"test": testtypes.StringTypeWithValidateAttributeWarning{},
				},
				map[string]attr.Value{
					"test": testtypes.StringValueWithValidateAttributeWarning{
						InternalString: testtypes.String{
							InternalString: types.StringValue("test"),
							CreatedBy:      testtypes.StringTypeWithValidateAttributeWarning{},
						},
					},
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(
					path.Root("test").AtName("test"),
					"Warning Diagnostic",
					"This is a warning.",
				),
			},
		},
		"struct-has-untagged-fields": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					Test                types.String `tfsdk:"test"`
					ExportedAndUntagged string
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: test: need a struct tag for \"tfsdk\" on ExportedAndUntagged",
				),
			},
		},
		"embedded-struct-has-untagged-fields": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.String `tfsdk:"test"`
					embedWithUntaggedExportedField
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: error retrieving embedded struct \"embedWithUntaggedExportedField\" field tags: "+
						"test: need a struct tag for \"tfsdk\" on ExportedAndUntagged",
				),
			},
		},
		"struct-has-invalid-tags": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.String `tfsdk:"invalidTag"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: test.invalidTag: invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter",
				),
			},
		},
		"struct-has-empty-tag": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					Test types.String `tfsdk:""`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: test.: invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter",
				),
			},
		},
		"embedded-struct-has-invalid-tags": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					embedWithInvalidTag
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: error retrieving embedded struct \"embedWithInvalidTag\" field tags: "+
						"test.invalidTag: invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter",
				),
			},
		},
		"struct-has-duplicate-tags": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					Test  types.String `tfsdk:"test"`
					Test2 types.String `tfsdk:"test"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: test.test: can't use tfsdk tag \"test\" for both Test and Test2 fields",
				),
			},
		},
		"embedded-struct-has-duplicate-tags": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					Attr1 types.String `tfsdk:"attr_1"`
					embedSingleField
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: embedded struct \"embedSingleField\" promotes a field with a duplicate tfsdk tag \"attr_1\", conflicts with \"Attr1\" tfsdk tag",
				),
			},
		},
		"list-zero-value": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"list": types.ListType{
						ElemType: types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					List types.List `tfsdk:"list"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("list"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.ListType[basetypes.StringType] / underlying type: tftypes.List[tftypes.String]\n"+
						"Received framework type from provider logic: types.ListType[!!! MISSING TYPE !!!] / underlying type: tftypes.List[tftypes.DynamicPseudoType]\n"+
						"Path: test.list",
				),
			},
		},
		"map-zero-value": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"map": types.MapType{
						ElemType: types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					Map types.Map `tfsdk:"map"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("map"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.MapType[basetypes.StringType] / underlying type: tftypes.Map[tftypes.String]\n"+
						"Received framework type from provider logic: types.MapType[!!! MISSING TYPE !!!] / underlying type: tftypes.Map[tftypes.DynamicPseudoType]\n"+
						"Path: test.map",
				),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/566
		"object-zero-value": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"object": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"test": types.StringType,
						},
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					Object types.Object `tfsdk:"object"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("object"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.ObjectType[\"test\":basetypes.StringType] / underlying type: tftypes.Object[\"test\":tftypes.String]\n"+
						"Received framework type from provider logic: types.ObjectType[] / underlying type: tftypes.Object[]\n"+
						"Path: test.object",
				),
			},
		},
		"set-zero-value": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"set": types.SetType{
						ElemType: types.StringType,
					},
				},
			},
			val: reflect.ValueOf(
				struct {
					Set types.Set `tfsdk:"set"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("set"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.SetType[basetypes.StringType] / underlying type: tftypes.Set[tftypes.String]\n"+
						"Received framework type from provider logic: types.SetType[!!! MISSING TYPE !!!] / underlying type: tftypes.Set[tftypes.DynamicPseudoType]\n"+
						"Path: test.set",
				),
			},
		},
		"embedded-exported-struct-with-pointer-not-supported": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					*EmbedSingleField
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("attr_1"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"struct { *reflect_test.EmbedSingleField } contains a struct embedded by a pointer which is not supported. Switch any embedded structs to be embedded by value.\n\n"+
						"Error: reflect: indirection through nil pointer to embedded struct field EmbedSingleField",
				),
			},
		},
		"embedded-unexported-struct-with-pointer-not-supported": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					*embedSingleField
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("attr_1"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"struct { *reflect_test.embedSingleField } contains a struct embedded by a pointer which is not supported. Switch any embedded structs to be embedded by value.\n\n"+
						"Error: reflect: indirection through nil pointer to embedded struct field embedSingleField",
				),
			},
		},
		"embedded-struct-has-empty-tfsdk-tag": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					embedSingleField `tfsdk:""`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: test.: embedded struct field embedSingleField cannot have tfsdk tag",
				),
			},
		},
		"embedded-struct-has-tfsdk-tag": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"attr_1": types.StringType,
				},
			},
			val: reflect.ValueOf(
				struct {
					embedSingleField `tfsdk:"attr_2"`
				}{},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from struct value. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error retrieving field names from struct tags: test.attr_2: embedded struct field embedSingleField cannot have tfsdk tag",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromStruct(
				context.Background(),
				testCase.typ,
				testCase.val,
				path.Root("test"),
			)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected result: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				for _, d := range diags {
					t.Logf("%s: %s\n%s\n", d.Severity(), d.Summary(), d.Detail())
				}
				t.Errorf("unexpected diagnostics: %s", diff)
			}
		})
	}
}

func TestFromStruct_structtags_ignores(t *testing.T) {
	t.Parallel()

	type s struct {
		ExportedAndTagged   string `tfsdk:"exported_and_tagged"`
		unexported          string //nolint:structcheck,unused
		unexportedAndTagged string `tfsdk:"unexported_and_tagged"`
		ExportedAndExcluded string `tfsdk:"-"`
	}
	testStruct := s{
		ExportedAndTagged:   "thisshouldstay",
		unexported:          "shouldntcopy",
		unexportedAndTagged: "shouldntcopy",
		ExportedAndExcluded: "shouldntcopy",
	}

	actualVal, diags := refl.FromStruct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"exported_and_tagged": types.StringType,
		},
	}, reflect.ValueOf(testStruct), path.Empty())
	if diags.HasError() {
		t.Fatalf("Unexpected error: %v", diags)
	}

	expectedVal := types.ObjectValueMust(
		map[string]attr.Type{
			"exported_and_tagged": types.StringType,
		},
		map[string]attr.Value{
			"exported_and_tagged": types.StringValue("thisshouldstay"),
		},
	)

	if diff := cmp.Diff(expectedVal, actualVal); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromStruct_embedded_structtags_ignores(t *testing.T) {
	t.Parallel()

	type s1 struct {
		ExportedAndTagged   string `tfsdk:"exported_and_tagged"`
		unexported          string //nolint:structcheck,unused
		unexportedAndTagged string `tfsdk:"unexported_and_tagged"`
		ExportedAndExcluded string `tfsdk:"-"`
	}

	type s2 struct {
		unexportedField string //nolint:structcheck,unused
	}

	type s struct {
		s1
		*s2 `tfsdk:"-"`
	}
	testStruct := s{
		s1{
			ExportedAndTagged:   "thisshouldstay",
			unexported:          "shouldntcopy",
			unexportedAndTagged: "shouldntcopy",
			ExportedAndExcluded: "shouldntcopy",
		},
		&s2{},
	}

	actualVal, diags := refl.FromStruct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"exported_and_tagged": types.StringType,
		},
	}, reflect.ValueOf(testStruct), path.Empty())
	if diags.HasError() {
		t.Fatalf("Unexpected error: %v", diags)
	}

	expectedVal := types.ObjectValueMust(
		map[string]attr.Type{
			"exported_and_tagged": types.StringType,
		},
		map[string]attr.Value{
			"exported_and_tagged": types.StringValue("thisshouldstay"),
		},
	)

	if diff := cmp.Diff(expectedVal, actualVal); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}
