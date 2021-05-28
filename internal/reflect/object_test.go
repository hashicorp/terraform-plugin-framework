package reflect

import (
	"context"
	"math/big"
	"reflect"
	"testing"

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

func TestReflectObjectIntoStruct_complex(t *testing.T) {
	t.Parallel()

	var s struct {
		Slice          []string `tfsdk:"slice"`
		SliceOfStructs []struct {
			A string `tfsdk:"a"`
			B int    `tfsdk:"b"`
		} `tfsdk:"slice_of_structs"`
		Struct struct {
			A     bool      `tfsdk:"a"`
			Slice []float64 `tfsdk:"slice"`
		} `tfsdk:"struct"`
		// TODO: add map
		// TODO: add tfsdk.AttributeValue
		// TODO: add setUnknownAble
		// TODO: add setNullable
		// TODO: add tftypes.ValueConverter
	}
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
	}), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
	reflect.ValueOf(&s).Elem().Set(result)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}
