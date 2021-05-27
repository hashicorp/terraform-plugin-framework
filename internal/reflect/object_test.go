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
	err := reflectObjectIntoStruct(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), Options{}, tftypes.NewAttributePath())
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
	err := reflectObjectIntoStruct(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(&s), Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `error retrieving field names from struct tags: : can't get struct tags of *string, is not a struct`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestReflectObjectIntoStruct_objectMissingFields(t *testing.T) {
	t.Parallel()

	var s struct {
		A string `tfsdk:"a"`
	}
	err := reflectObjectIntoStruct(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}, map[string]tftypes.Value{}), reflect.ValueOf(&s), Options{}, tftypes.NewAttributePath())
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
	err := reflectObjectIntoStruct(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(&s), Options{}, tftypes.NewAttributePath())
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
	err := reflectObjectIntoStruct(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"b": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"b": tftypes.NewValue(tftypes.String, "hello"),
	}), reflect.ValueOf(&s), Options{}, tftypes.NewAttributePath())
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
	err := reflectObjectIntoStruct(context.Background(), tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
			"b": tftypes.Number,
			"c": tftypes.Bool,
		},
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello"),
		"b": tftypes.NewValue(tftypes.Number, 123),
		"c": tftypes.NewValue(tftypes.Bool, true),
	}), reflect.ValueOf(&s), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestReflectObjectIntoStruct_complex(t *testing.T) {
	t.Parallel()
}
