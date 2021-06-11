package reflect_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func (u *unknownableString) GetUnknown(_ context.Context) bool {
	return u.Unknown
}

func (u *unknownableString) SetValue(_ context.Context, value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("can't set type %T", value)
	}
	u.String = v
	return nil
}

func (u *unknownableString) GetValue(_ context.Context) interface{} {
	return u.String
}

var _ refl.Unknownable = &unknownableString{}

type unknownableStringError struct {
	String  string
	Unknown bool
}

func (u *unknownableStringError) SetUnknown(_ context.Context, unknown bool) error {
	return errors.New("this is an error")
}

func (u *unknownableStringError) SetValue(_ context.Context, val interface{}) error {
	v, ok := val.(string)
	if !ok {
		return fmt.Errorf("can't set type %T", val)
	}
	u.String = v
	return nil
}

func (u *unknownableStringError) GetUnknown(_ context.Context) bool {
	return u.Unknown
}

func (u *unknownableStringError) GetValue(_ context.Context) interface{} {
	return u.String
}

var _ refl.Unknownable = &unknownableStringError{}

type nullableString struct {
	String string
	Null   bool
}

var _ refl.Nullable = &nullableString{}

func (n *nullableString) SetNull(_ context.Context, null bool) error {
	n.Null = null
	return nil
}

func (n *nullableString) SetValue(_ context.Context, value interface{}) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("can't set type %T", value)
	}
	n.String = val
	return nil
}

func (n *nullableString) GetNull(_ context.Context) bool {
	return n.Null
}

func (n *nullableString) GetValue(_ context.Context) interface{} {
	return n.String
}

type nullableStringError struct {
	String string
	Null   bool
}

func (n *nullableStringError) SetNull(_ context.Context, null bool) error {
	return errors.New("this is an error")
}

func (n *nullableStringError) SetValue(_ context.Context, value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("can't set type %T", value)
	}
	n.String = v
	return nil
}

func (n *nullableStringError) GetNull(_ context.Context) bool {
	return n.Null
}

func (n *nullableStringError) GetValue(_ context.Context) interface{} {
	return n.String
}

var _ refl.Nullable = &nullableStringError{}

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

type valueCreator struct {
	value   string
	unknown bool
	null    bool
}

func (v *valueCreator) ToTerraform5Value() (interface{}, error) {
	if v.unknown {
		return tftypes.UnknownValue, nil
	}
	if v.null {
		return nil, nil
	}
	return v.value, nil
}

func (v *valueCreator) Equal(o *valueCreator) bool {
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

var _ tftypes.ValueCreator = &valueCreator{}

func TestNewUnknownable_known(t *testing.T) {
	t.Parallel()

	unknownable := &unknownableString{
		Unknown: true,
	}
	res, err := refl.NewUnknownable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(unknownable), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*unknownableString)
	if got.Unknown != false {
		t.Errorf("Expected %v, got %v", false, got.Unknown)
	}
}

func TestNewUnknownable_unknown(t *testing.T) {
	t.Parallel()

	var unknownable *unknownableString
	res, err := refl.NewUnknownable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(unknownable), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*unknownableString)
	if got.Unknown != true {
		t.Errorf("Expected %v, got %v", true, got.Unknown)
	}
}

func TestNewUnknownable_error(t *testing.T) {
	t.Parallel()

	var unknownable *unknownableStringError
	_, err := refl.NewUnknownable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(unknownable), refl.Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}

func TestFromUnknownable_unknown(t *testing.T) {
	t.Parallel()

	foo := &unknownableString{
		Unknown: true,
	}
	expected := types.String{Unknown: true}
	got, err := refl.FromUnknownable(context.Background(), types.StringType, foo, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromUnknownable_value(t *testing.T) {
	t.Parallel()

	foo := &unknownableString{
		String: "hello, world",
	}
	expected := types.String{Value: "hello, world"}
	got, err := refl.FromUnknownable(context.Background(), types.StringType, foo, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestNewNullable_notNull(t *testing.T) {
	t.Parallel()

	nullable := &nullableString{
		Null: true,
	}
	res, err := refl.NewNullable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(nullable), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*nullableString)
	if got.Null != false {
		t.Errorf("Expected %v, got %v", false, got.Null)
	}
}

func TestNewNullable_null(t *testing.T) {
	t.Parallel()

	var nullable *nullableString
	res, err := refl.NewNullable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(nullable), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*nullableString)
	if got.Null != true {
		t.Errorf("Expected %v, got %v", true, got.Null)
	}
}

func TestNewNullable_error(t *testing.T) {
	t.Parallel()

	var nullable *nullableStringError
	_, err := refl.NewNullable(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(nullable), refl.Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}

func TestFromNullable_null(t *testing.T) {
	t.Parallel()

	foo := &nullableString{
		Null: true,
	}
	expected := types.String{Null: true}
	got, err := refl.FromNullable(context.Background(), types.StringType, foo, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromNullable_value(t *testing.T) {
	t.Parallel()

	foo := &nullableString{
		String: "hello, world",
	}
	expected := types.String{Value: "hello, world"}
	got, err := refl.FromNullable(context.Background(), types.StringType, foo, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestNewAttributeValue_unknown(t *testing.T) {
	t.Parallel()

	var av types.String
	res, err := refl.NewAttributeValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(av), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(types.String)
	expected := types.String{Unknown: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestNewAttributeValue_null(t *testing.T) {
	t.Parallel()

	var av types.String
	res, err := refl.NewAttributeValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(av), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(types.String)
	expected := types.String{Null: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestFromAttributeValue_null(t *testing.T) {
	t.Parallel()

	expected := types.String{Null: true}
	got, err := refl.FromAttributeValue(context.Background(), types.StringType, expected, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromAttributeValue_unknown(t *testing.T) {
	t.Parallel()

	expected := types.String{Unknown: true}
	got, err := refl.FromAttributeValue(context.Background(), types.StringType, expected, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromAttributeValue_value(t *testing.T) {
	t.Parallel()

	expected := types.String{Value: "hello, world"}
	got, err := refl.FromAttributeValue(context.Background(), types.StringType, expected, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestNewAttributeValue_value(t *testing.T) {
	t.Parallel()

	var av types.String
	res, err := refl.NewAttributeValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(av), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(types.String)
	expected := types.String{Value: "hello"}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestNewValueConverter_unknown(t *testing.T) {
	t.Parallel()

	var vc *valueConverter
	res, err := refl.NewValueConverter(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(vc), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*valueConverter)
	expected := &valueConverter{unknown: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestNewValueConverter_null(t *testing.T) {
	t.Parallel()

	var vc *valueConverter
	res, err := refl.NewValueConverter(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(vc), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*valueConverter)
	expected := &valueConverter{null: true}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestNewValueConverter_value(t *testing.T) {
	t.Parallel()

	var vc *valueConverter
	res, err := refl.NewValueConverter(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(vc), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	got := res.Interface().(*valueConverter)
	expected := &valueConverter{value: "hello"}
	if !got.Equal(expected) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestNewValueConverter_error(t *testing.T) {
	t.Parallel()

	var vc *valueConverterError
	_, err := refl.NewValueConverter(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(vc), refl.Options{}, tftypes.NewAttributePath())
	if expected := ": this is an error"; err == nil || err.Error() != expected {
		t.Errorf("Expected error to be %q, got %v", expected, err)
	}
}

func TestFromValueCreator_null(t *testing.T) {
	t.Parallel()

	vc := &valueCreator{
		null: true,
	}
	expected := types.String{Null: true}
	got, err := refl.FromValueCreator(context.Background(), types.StringType, vc, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromValueCreator_unknown(t *testing.T) {
	t.Parallel()

	vc := &valueCreator{
		unknown: true,
	}
	expected := types.String{Unknown: true}
	got, err := refl.FromValueCreator(context.Background(), types.StringType, vc, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromValueCreator_value(t *testing.T) {
	t.Parallel()

	vc := &valueCreator{
		value: "hello, world",
	}
	expected := types.String{Value: "hello, world"}
	got, err := refl.FromValueCreator(context.Background(), types.StringType, vc, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}
