package reflect

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	tfsdk "github.com/hashicorp/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestReflectObjectIntoStruct_notAnObject(t *testing.T) {
	t.Parallel()

	var s struct{}
	_, err := reflectStructFromObject(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: can't reflect tftypes.String into a struct, must be an object`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestReflectObjectIntoStruct_notAStruct(t *testing.T) {
	t.Parallel()

	var s string
	_, err := reflectStructFromObject(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: expected a struct type, got string`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestReflectObjectIntoStruct_objectMissingFields(t *testing.T) {
	t.Parallel()

	var s struct {
		A string `tfsdk:"a"`
	}
	_, err := reflectStructFromObject(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}, map[string]tftypes.Value{}), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: mismatch between struct and object: Struct defines fields not found in object: a.`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestReflectObjectIntoStruct_structMissingProperties(t *testing.T) {
	t.Parallel()

	var s struct{}
	_, err := reflectStructFromObject(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: mismatch between struct and object: Object defines fields not found in struct: a.`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestReflectObjectIntoStruct_objectMissingFieldsAndStructMissingProperties(t *testing.T) {
	t.Parallel()

	var s struct {
		A string `tfsdk:"a"`
	}
	_, err := reflectStructFromObject(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"b": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"b": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: mismatch between struct and object: Struct defines fields not found in object: a. Object defines fields not found in struct: b.`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestReflectObjectIntoStruct_primitives(t *testing.T) {
	t.Parallel()

	var s struct {
		A string     `tfsdk:"a"`
		B *big.Float `tfsdk:"b"`
		C bool       `tfsdk:"c"`
	}
	result, err := reflectStructFromObject(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
			"b": tftypes.Number,
			"c": tftypes.Bool,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
		"b": tftypes.NewValue(tftypes.Number, 123),
		"c": tftypes.NewValue(tftypes.Bool, true),
	}), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
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

type unknownableString struct {
	String  string
	Unknown bool
}

func (u *unknownableString) SetUnknown(_ context.Context, unknown bool) error {
	u.Unknown = unknown
	return nil
}

type nullableString struct {
	String string
	Null   bool
}

func (n *nullableString) SetNull(_ context.Context, null bool) error {
	n.Null = null
	return nil
}

type attributeValue struct {
	Value   string
	Null    bool
	Unknown bool
}

func (a *attributeValue) ToTerraformValue(_ context.Context) (interface{}, error) {
	var val interface{}
	if a.Null {
		val = nil
	}
	if a.Value != "" {
		val = a.Value
	}
	if a.Unknown {
		val = tftypes.UnknownValue
	}
	return val, nil
}

func (a *attributeValue) SetTerraformValue(_ context.Context, val tftypes.Value) error {
	a.Value = ""
	a.Null = false
	a.Unknown = false
	if val.IsNull() {
		a.Null = true
		return nil
	}
	if !val.IsKnown() {
		a.Unknown = true
		return nil
	}
	err := val.As(&a.Value)
	return err
}

func (a *attributeValue) Equal(o tfsdk.AttributeValue) bool {
	other, ok := o.(*attributeValue)
	if !ok {
		return false
	}
	return a.Value == other.Value && a.Null == other.Null && a.Unknown == other.Unknown
}

type valueConverter struct {
	value   string
	unknown bool
	null    bool
}

func (v *valueConverter) FromTerraform5Value(in tftypes.Value) error {
	v.value = ""
	v.unknown = false
	v.null = false
	if !in.IsKnown() {
		v.unknown = true
	}
	if in.IsNull() {
		v.null = true
	}
	return in.As(&v.value)
}

func (v *valueConverter) Equal(o *valueConverter) bool {
	if v == nil && o == nil {
		return true
	}
	if v == nil {
		return false
	}
	if o == nil {
		return false
	}
	if v.unknown != o.unknown {
		return false
	}
	if v.null != o.null {
		return false
	}
	return v.value == o.value
}

func TestReflectObjectIntoStruct_complex(t *testing.T) {
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
		Map            map[string][]string `tfsdk:"map"`
		Pointer        *string             `tfsdk:"pointer"`
		Unknownable    *unknownableString  `tfsdk:"unknownable"`
		Nullable       *nullableString     `tfsdk:"nullable"`
		AttributeValue *attributeValue     `tfsdk:"attribute_value"`
		ValueConverter *valueConverter     `tfsdk:"value_converter"`
		// TODO: add unhandled null
		// TODO: add unhandled unknown
	}
	var s myStruct
	result, err := reflectStructFromObject(context.Background(), tftypes.NewValue(tftypes.Object{
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
			"map": tftypes.Map{
				AttributeType: tftypes.List{
					ElementType: tftypes.String,
				},
			},
			"pointer":         tftypes.String,
			"unknownable":     tftypes.String,
			"nullable":        tftypes.String,
			"attribute_value": tftypes.String,
			"value_converter": tftypes.String,
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
		"pointer":         tftypes.NewValue(tftypes.String, "pointed"),
		"unknownable":     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"nullable":        tftypes.NewValue(tftypes.String, nil),
		"attribute_value": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_converter": tftypes.NewValue(tftypes.String, nil),
	}), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
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
		AttributeValue: &attributeValue{
			Unknown: true,
		},
		ValueConverter: &valueConverter{
			null: true,
		},
	}
	if diff := cmp.Diff(s, expected); diff != "" {
		t.Errorf("Didn't get expected value. Diff (+ is expected, - is result): %s", diff)
	}
}
