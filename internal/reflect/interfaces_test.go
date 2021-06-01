package reflect

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

var _ setUnknownable = &unknownableString{}

type unknownableStringError struct {
	String  string
	Unknown bool
}

func (u *unknownableStringError) SetUnknown(_ context.Context, unknown bool) error {
	return errors.New("this is an error")
}

var _ setUnknownable = &unknownableStringError{}

type nullableString struct {
	String string
	Null   bool
}

var _ setNullable = &nullableString{}

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

var _ setNullable = &nullableStringError{}

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

func (a *attributeValue) Equal(o attr.Value) bool {
	other, ok := o.(*attributeValue)
	if !ok {
		return false
	}
	return a.Value == other.Value && a.Null == other.Null && a.Unknown == other.Unknown
}

var _ attr.Value = &attributeValue{}

type attributeValueError struct {
	*attributeValue
}

func (a *attributeValueError) SetTerraformValue(_ context.Context, _ tftypes.Value) error {
	return errors.New("this is an error")
}

var _ attr.Value = &attributeValueError{}

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

func TestReflectUnknownable_known(t *testing.T) {
	t.Parallel()

	unknownable := &unknownableString{
		Unknown: true,
	}
	res, err := reflectUnknownable(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(unknownable), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*unknownableString)
	if got.Unknown != false {
		t.Errorf("Expected %v, got %v", false, got.Unknown)
	}
}

func TestReflectUnknownable_unknown(t *testing.T) {
	t.Parallel()

	var unknownable *unknownableString
	res, err := reflectUnknownable(context.Background(), tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(unknownable), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*unknownableString)
	if got.Unknown != true {
		t.Errorf("Expected %v, got %v", true, got.Unknown)
	}
}

func TestReflectUnknownable_error(t *testing.T) {
	t.Parallel()

	var unknownable *unknownableStringError
	_, err := reflectUnknownable(context.Background(), tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(unknownable), Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}

func TestReflectNullable_notNull(t *testing.T) {
	t.Parallel()

	nullable := &nullableString{
		Null: true,
	}
	res, err := reflectNullable(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(nullable), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*nullableString)
	if got.Null != false {
		t.Errorf("Expected %v, got %v", false, got.Null)
	}
}

func TestReflectNullable_null(t *testing.T) {
	t.Parallel()

	var nullable *nullableString
	res, err := reflectNullable(context.Background(), tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(nullable), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*nullableString)
	if got.Null != true {
		t.Errorf("Expected %v, got %v", true, got.Null)
	}
}

func TestReflectNullable_error(t *testing.T) {
	t.Parallel()

	var nullable *nullableStringError
	_, err := reflectNullable(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(nullable), Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}

func TestReflectAttributeValue_unknown(t *testing.T) {
	t.Parallel()

	var av *attributeValue
	res, err := reflectAttributeValue(context.Background(), tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(av), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*attributeValue)
	expected := &attributeValue{Unknown: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestReflectAttributeValue_null(t *testing.T) {
	t.Parallel()

	var av *attributeValue
	res, err := reflectAttributeValue(context.Background(), tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(av), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*attributeValue)
	expected := &attributeValue{Null: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestReflectAttributeValue_value(t *testing.T) {
	t.Parallel()

	var av *attributeValue
	res, err := reflectAttributeValue(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(av), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*attributeValue)
	expected := &attributeValue{Value: "hello"}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestReflectAttributeValue_error(t *testing.T) {
	t.Parallel()

	var av *attributeValueError
	_, err := reflectAttributeValue(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(av), Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}

func TestReflectValueConverter_unknown(t *testing.T) {
	t.Parallel()

	var vc *valueConverter
	res, err := reflectValueConverter(context.Background(), tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(vc), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*valueConverter)
	expected := &valueConverter{unknown: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestReflectValueConverter_null(t *testing.T) {
	t.Parallel()

	var vc *valueConverter
	res, err := reflectValueConverter(context.Background(), tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(vc), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*valueConverter)
	expected := &valueConverter{null: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestReflectValueConverter_value(t *testing.T) {
	t.Parallel()

	var vc *valueConverter
	res, err := reflectValueConverter(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(vc), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*valueConverter)
	expected := &valueConverter{value: "hello"}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestReflectValueConverter_error(t *testing.T) {
	t.Parallel()

	var vc *valueConverterError
	_, err := reflectValueConverter(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(vc), Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}
