package reflect_test

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStruct_notAnObject(t *testing.T) {
	t.Parallel()

	var s struct{}
	_, err := refl.Struct(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: can't reflect tftypes.String into a struct, must be an object`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestStruct_notAStruct(t *testing.T) {
	t.Parallel()

	var s string
	_, err := refl.Struct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"a": types.StringType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: expected a struct type, got string`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestStruct_objectMissingFields(t *testing.T) {
	t.Parallel()

	var s struct {
		A string `tfsdk:"a"`
	}
	_, err := refl.Struct(context.Background(), types.ObjectType{}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}, map[string]tftypes.Value{}), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: mismatch between struct and object: Struct defines fields not found in object: a.`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestStruct_structMissingProperties(t *testing.T) {
	t.Parallel()

	var s struct{}
	_, err := refl.Struct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"a": types.StringType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: mismatch between struct and object: Object defines fields not found in struct: a.`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestStruct_objectMissingFieldsAndStructMissingProperties(t *testing.T) {
	t.Parallel()

	var s struct {
		A string `tfsdk:"a"`
	}
	_, err := refl.Struct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"a": types.StringType,
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"b": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"b": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: mismatch between struct and object: Struct defines fields not found in object: a. Object defines fields not found in struct: b.`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestStruct_primitives(t *testing.T) {
	t.Parallel()

	var s struct {
		A string     `tfsdk:"a"`
		B *big.Float `tfsdk:"b"`
		C bool       `tfsdk:"c"`
	}
	result, err := refl.Struct(context.Background(), types.ObjectType{
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
	}), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
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

func TestStruct_complex(t *testing.T) {
	t.Parallel()

	type myStruct struct {
		Slice          []string `tfsdk:"slice"`
		SliceOfStructs []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		} `tfsdk:"slice_of_structs"`
		Struct struct {
			A     bool      `tfsdk:"a"`
			Slice []float64 `tfsdk:"slice"`
		} `tfsdk:"struct"`
		// TODO: uncomment when maps are supported
		// Map              map[string][]string `tfsdk:"map"`
		Pointer          *string            `tfsdk:"pointer"`
		Unknownable      *unknownableString `tfsdk:"unknownable"`
		Nullable         *nullableString    `tfsdk:"nullable"`
		AttributeValue   types.String       `tfsdk:"attribute_value"`
		ValueConverter   *valueConverter    `tfsdk:"value_converter"`
		UnhandledNull    string             `tfsdk:"unhandled_null"`
		UnhandledUnknown string             `tfsdk:"unhandled_unknown"`
	}
	var s myStruct
	result, err := refl.Struct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"slice": types.ListType{
				ElemType: types.StringType,
			},
			"slice_of_structs": types.ListType{
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
			// TODO: uncomment when maps are supported
			/*
				"map": testMapType{
					ElemType: types.ListType{
						ElemType: types.StringType,
					},
				},
			*/
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
			"slice": tftypes.List{
				ElementType: tftypes.String,
			},
			"slice_of_structs": tftypes.List{
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
			// TODO: uncomment when maps are supported
			/*
				"map": tftypes.Map{
					AttributeType: tftypes.List{
						ElementType: tftypes.String,
					},
				},
			*/
			"pointer":           tftypes.String,
			"unknownable":       tftypes.String,
			"nullable":          tftypes.String,
			"attribute_value":   tftypes.String,
			"value_converter":   tftypes.String,
			"unhandled_null":    tftypes.String,
			"unhandled_unknown": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"slice": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.String,
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "red"),
			tftypes.NewValue(tftypes.String, "blue"),
			tftypes.NewValue(tftypes.String, "green"),
		}),
		"slice_of_structs": tftypes.NewValue(tftypes.List{
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
		// TODO: uncomment when maps are supported
		/*
			"map": tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.List{
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
		*/
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
	}, tftypes.NewAttributePath())
	reflect.ValueOf(&s).Elem().Set(result)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	str := "pointed"
	expected := myStruct{
		Slice: []string{"red", "blue", "green"},
		SliceOfStructs: []struct {
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
		// TODO: uncomment when maps are supported
		/*
			Map: map[string][]string{
				"colors": {"red", "orange", "yellow"},
				"fruits": {"apple", "banana"},
			},
		*/
		Pointer: &str,
		Unknownable: &unknownableString{
			Unknown: true,
		},
		Nullable: &nullableString{
			Null: true,
		},
		AttributeValue: types.String{
			Unknown: true,
		},
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

func TestFromStruct(t *testing.T) {
	type disk struct {
		Name string `tfsdk:"name"`
		// bool
	}
	disk1 := disk{
		Name: "myfirstdisk",
	}

	actualVal, err := refl.FromStruct(context.Background(), types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
		},
	}, reflect.ValueOf(disk1), refl.OutOfOptions{}, tftypes.NewAttributePath())
	if err != nil {
		t.Fatal(err)
	}

	expectedVal := types.Object{
		Attrs: map[string]attr.Value{
			"name": types.String{Value: "myfirstdisk"},
		},
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
		},
	}

	if diff := cmp.Diff(expectedVal, actualVal); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}
