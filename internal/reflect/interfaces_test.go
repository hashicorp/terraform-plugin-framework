package reflect_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type unknownableString struct {
	String  string
	Unknown bool
}

func (u *unknownableString) SetUnknown(_ context.Context, unknown bool) error {
	u.Unknown = unknown
	return nil
}

var _ refl.SetUnknownable = &unknownableString{}

type unknownableStringError struct {
	String  string
	Unknown bool
}

func (u *unknownableStringError) SetUnknown(_ context.Context, unknown bool) error {
	return errors.New("this is an error")
}

var _ refl.SetUnknownable = &unknownableStringError{}

type nullableString struct {
	String string
	Null   bool
}

var _ refl.SetNullable = &nullableString{}

func (n *nullableString) SetNull(_ context.Context, null bool) error {
	n.Null = null
	return nil
}

type nullableStringError struct {
	String string
	Null   bool
}

func (n *nullableStringError) SetNull(_ context.Context, null bool) error {
	return errors.New("this is an error")
}

var _ refl.SetNullable = &nullableStringError{}

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
		return nil
	}
	if in.IsNull() {
		v.null = true
		return nil
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

var _ tftypes.ValueConverter = &valueConverter{}

type valueConverterError struct {
	*valueConverter
}

func (v *valueConverterError) FromTerraform5Value(_ tftypes.Value) error {
	return errors.New("this is an error")
}

var _ tftypes.ValueConverter = &valueConverterError{}

func TestUnknownable_known(t *testing.T) {
	t.Parallel()

	unknownable := &unknownableString{
		Unknown: true,
	}
	res, err := refl.Unknownable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(unknownable), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*unknownableString)
	if got.Unknown != false {
		t.Errorf("Expected %v, got %v", false, got.Unknown)
	}
}

func TestUnknownable_unknown(t *testing.T) {
	t.Parallel()

	var unknownable *unknownableString
	res, err := refl.Unknownable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(unknownable), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*unknownableString)
	if got.Unknown != true {
		t.Errorf("Expected %v, got %v", true, got.Unknown)
	}
}

func TestUnknownable_error(t *testing.T) {
	t.Parallel()

	var unknownable *unknownableStringError
	_, err := refl.Unknownable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(unknownable), refl.Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}

func TestNullable_notNull(t *testing.T) {
	t.Parallel()

	nullable := &nullableString{
		Null: true,
	}
	res, err := refl.Nullable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(nullable), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*nullableString)
	if got.Null != false {
		t.Errorf("Expected %v, got %v", false, got.Null)
	}
}

func TestNullable_null(t *testing.T) {
	t.Parallel()

	var nullable *nullableString
	res, err := refl.Nullable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(nullable), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*nullableString)
	if got.Null != true {
		t.Errorf("Expected %v, got %v", true, got.Null)
	}
}

func TestNullable_error(t *testing.T) {
	t.Parallel()

	var nullable *nullableStringError
	_, err := refl.Nullable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(nullable), refl.Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}

func TestAttributeValue_unknown(t *testing.T) {
	t.Parallel()

	var av types.String
	res, err := refl.AttributeValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(av), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(types.String)
	expected := types.String{Unknown: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestAttributeValue_null(t *testing.T) {
	t.Parallel()

	var av types.String
	res, err := refl.AttributeValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(av), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(types.String)
	expected := types.String{Null: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestAttributeValue_value(t *testing.T) {
	t.Parallel()

	var av types.String
	res, err := refl.AttributeValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(av), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(types.String)
	expected := types.String{Value: "hello"}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestValueConverter_unknown(t *testing.T) {
	t.Parallel()

	var vc *valueConverter
	res, err := refl.ValueConverter(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(vc), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*valueConverter)
	expected := &valueConverter{unknown: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestValueConverter_null(t *testing.T) {
	t.Parallel()

	var vc *valueConverter
	res, err := refl.ValueConverter(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(vc), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*valueConverter)
	expected := &valueConverter{null: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestValueConverter_value(t *testing.T) {
	t.Parallel()

	var vc *valueConverter
	res, err := refl.ValueConverter(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(vc), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*valueConverter)
	expected := &valueConverter{value: "hello"}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestValueConverter_error(t *testing.T) {
	t.Parallel()

	var vc *valueConverterError
	_, err := refl.ValueConverter(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(vc), refl.Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}
