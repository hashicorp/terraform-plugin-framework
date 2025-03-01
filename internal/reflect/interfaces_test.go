// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

func TestNewUnknownable(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		val           tftypes.Value
		target        reflect.Value
		expected      bool
		expectedDiags diag.Diagnostics
	}{
		"known": {
			val: tftypes.NewValue(tftypes.String, "hello"),
			target: reflect.ValueOf(&unknownableString{
				Unknown: true,
			}),
			expected: false,
		},
		"unknown": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			target:   reflect.ValueOf(new(unknownableString)),
			expected: true,
		},
		"error": {
			val:    tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			target: reflect.ValueOf(new(unknownableStringError)),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert into a value. This is always an error in the provider. Please report the following to the provider developer:\n\nreflection error: this is an error",
				),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res, diags := refl.NewUnknownable(context.Background(), types.StringType, tc.val, tc.target, refl.Options{}, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Fatalf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diags.HasError() {
				return
			}

			got, ok := res.Interface().(*unknownableString)
			if !ok {
				t.Fatalf("Expected type of *unknownableString, got %T", res.Interface())
			}

			if got.Unknown != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, got.Unknown)
			}
		})
	}
}

func TestFromUnknownable(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.Type
		val           refl.Unknownable
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"unknown": {
			typ: types.StringType,
			val: &unknownableString{
				Unknown: true,
			},
			expected: types.StringUnknown(),
		},
		"unknown-validate-error": {
			typ: testtypes.StringTypeWithValidateError{},
			val: &unknownableString{
				Unknown: true,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
		"unknown-validate-attribute-error": {
			typ: testtypes.StringTypeWithValidateAttributeError{},
			val: &unknownableString{
				Unknown: true,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
		"value": {
			typ: types.StringType,
			val: &unknownableString{
				String: "hello, world",
			},
			expected: types.StringValue("hello, world"),
		},
		"value-validate-error": {
			typ: testtypes.StringTypeWithValidateError{},
			val: &unknownableString{
				String: "hello, world",
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
		"value-validate-attribute-error": {
			typ: testtypes.StringTypeWithValidateAttributeError{},
			val: &unknownableString{
				String: "hello, world",
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromUnknownable(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestNewNullable(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		val           tftypes.Value
		target        reflect.Value
		expected      bool
		expectedDiags diag.Diagnostics
	}{
		"not-null": {
			val: tftypes.NewValue(tftypes.String, "hello"),
			target: reflect.ValueOf(&nullableString{
				Null: true,
			}),
			expected: false,
		},
		"null": {
			val:      tftypes.NewValue(tftypes.String, nil),
			target:   reflect.ValueOf(new(nullableString)),
			expected: true,
		},
		"error": {
			val:    tftypes.NewValue(tftypes.String, "hello"),
			target: reflect.ValueOf(new(nullableStringError)),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert into a value. This is always an error in the provider. Please report the following to the provider developer:\n\nreflection error: this is an error",
				),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res, diags := refl.NewNullable(context.Background(), types.StringType, tc.val, tc.target, refl.Options{}, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diags.HasError() {
				return
			}

			got, ok := res.Interface().(*nullableString)
			if !ok {
				t.Fatalf("Expected type of *nullableString, got %T", res.Interface())
			}

			if got.Null != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, got.Null)
			}
		})
	}
}

func TestFromNullable(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.Type
		val           refl.Nullable
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"null": {
			typ: types.StringType,
			val: &nullableString{
				Null: true,
			},
			expected: types.StringNull(),
		},
		"null-validate-error": {
			typ: testtypes.StringTypeWithValidateError{},
			val: &nullableString{
				Null: true,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
		"null-validate-attribute-error": {
			typ: testtypes.StringTypeWithValidateAttributeError{},
			val: &nullableString{
				Null: true,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
		"value": {
			typ: types.StringType,
			val: &nullableString{
				String: "hello, world",
			},
			expected: types.StringValue("hello, world"),
		},
		"value-validate-error": {
			typ: testtypes.StringTypeWithValidateError{},
			val: &nullableString{
				String: "hello, world",
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
		"value-validate-attribute-error": {
			typ: testtypes.StringTypeWithValidateAttributeError{},
			val: &nullableString{
				String: "hello, world",
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromNullable(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestNewAttributeValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.Type
		val           tftypes.Value
		target        reflect.Value
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"unknown": {
			typ:      types.StringType,
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			target:   reflect.ValueOf(types.String{}),
			expected: types.StringUnknown(),
		},
		"null": {
			typ:      types.StringType,
			val:      tftypes.NewValue(tftypes.String, nil),
			target:   reflect.ValueOf(types.String{}),
			expected: types.StringNull(),
		},
		"value": {
			typ:      types.StringType,
			val:      tftypes.NewValue(tftypes.String, "hello"),
			target:   reflect.ValueOf(types.String{}),
			expected: types.StringValue("hello"),
		},
		"validate-error": {
			typ: testtypes.StringTypeWithValidateError{},
			val: tftypes.NewValue(tftypes.String, "hello"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
		"validate-attribute-error": {
			typ: testtypes.StringTypeWithValidateAttributeError{},
			val: tftypes.NewValue(tftypes.String, "hello"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res, diags := refl.NewAttributeValue(context.Background(), tc.typ, tc.val, tc.target, refl.Options{}, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diags.HasError() {
				return
			}

			got, ok := res.Interface().(types.String)
			if !ok {
				t.Fatalf("Expected type of types.String, got %T", res.Interface())
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFromAttributeValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.Type
		val           attr.Value
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"BoolType-BoolValue": {
			typ:      types.BoolType,
			val:      types.BoolNull(),
			expected: types.BoolNull(),
		},
		"BoolTypable-BoolValuable": {
			typ: testtypes.BoolType{},
			val: testtypes.Bool{
				CreatedBy: testtypes.BoolType{},
			},
			expected: testtypes.Bool{
				CreatedBy: testtypes.BoolType{},
			},
		},
		"BoolTypable-BoolValue": {
			typ:      testtypes.BoolTypeWithSemanticEquals{},
			val:      types.BoolNull(),
			expected: types.BoolNull(),
		},
		"BoolTypable-BoolValue-ValidateError": {
			typ:      testtypes.BoolTypeWithValidateError{},
			val:      types.BoolNull(),
			expected: types.BoolNull(),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "Error Diagnostic", "This is an error."),
			},
		},
		"BoolTypable-BoolValue-ValidateAttributeError": {
			typ: testtypes.BoolType{},
			val: testtypes.BoolValueWithValidateAttributeError{
				Bool: testtypes.Bool{
					CreatedBy: testtypes.BoolTypeWithValidateAttributeError{},
				},
			},
			expected: testtypes.BoolValueWithValidateAttributeError{Bool: testtypes.Bool{CreatedBy: testtypes.BoolTypeWithValidateAttributeError{}}},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "Error Diagnostic", "This is an error."),
			},
		},
		"Float32Type-Float32Value": {
			typ:      types.Float32Type,
			val:      types.Float32Null(),
			expected: types.Float32Null(),
		},
		"Float32Typable-Float32Value": {
			typ:      testtypes.Float32TypeWithSemanticEquals{},
			val:      types.Float32Null(),
			expected: types.Float32Null(),
		},
		"Float64Type-Float64Value": {
			typ:      types.Float64Type,
			val:      types.Float64Null(),
			expected: types.Float64Null(),
		},
		"Float64Typable-Float64Value": {
			typ:      testtypes.Float64TypeWithSemanticEquals{},
			val:      types.Float64Null(),
			expected: types.Float64Null(),
		},
		"Int32Type-Int32Value": {
			typ:      types.Int32Type,
			val:      types.Int32Null(),
			expected: types.Int32Null(),
		},
		"Int32Typable-Int32Value": {
			typ:      testtypes.Int32TypeWithSemanticEquals{},
			val:      types.Int32Null(),
			expected: types.Int32Null(),
		},
		"Int64Type-Int64Value": {
			typ:      types.Int64Type,
			val:      types.Int64Null(),
			expected: types.Int64Null(),
		},
		"Int64Typable-Int64Value": {
			typ:      testtypes.Int64TypeWithSemanticEquals{},
			val:      types.Int64Null(),
			expected: types.Int64Null(),
		},
		"ListType-ListValue-matching-elements": {
			typ:      types.ListType{ElemType: types.StringType},
			val:      types.ListNull(types.StringType),
			expected: types.ListNull(types.StringType),
		},
		"ListType-ListValue-mismatching-elements": {
			typ:      types.ListType{ElemType: types.StringType},
			val:      types.ListNull(types.BoolType),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.ListType[basetypes.StringType] / underlying type: tftypes.List[tftypes.String]\n"+
						"Received framework type from provider logic: types.ListType[basetypes.BoolType] / underlying type: tftypes.List[tftypes.Bool]\n"+
						"Path: test",
				),
			},
		},
		"ListTypable-ListValue-matching-elements": {
			typ: testtypes.ListType{
				ListType: types.ListType{ElemType: types.StringType},
			},
			val:      types.ListNull(types.StringType),
			expected: types.ListNull(types.StringType),
		},
		"MapType-MapValue-matching-elements": {
			typ:      types.MapType{ElemType: types.StringType},
			val:      types.MapNull(types.StringType),
			expected: types.MapNull(types.StringType),
		},
		"MapType-MapValue-mismatching-elements": {
			typ:      types.MapType{ElemType: types.StringType},
			val:      types.MapNull(types.BoolType),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.MapType[basetypes.StringType] / underlying type: tftypes.Map[tftypes.String]\n"+
						"Received framework type from provider logic: types.MapType[basetypes.BoolType] / underlying type: tftypes.Map[tftypes.Bool]\n"+
						"Path: test",
				),
			},
		},
		"MapTypable-MapValue-matching-elements": {
			typ: testtypes.MapType{
				MapType: types.MapType{ElemType: types.StringType},
			},
			val:      types.MapNull(types.StringType),
			expected: types.MapNull(types.StringType),
		},
		"NumberType-NumberValue": {
			typ:      types.NumberType,
			val:      types.NumberNull(),
			expected: types.NumberNull(),
		},
		"NumberTypable-NumberValuable": {
			typ: testtypes.NumberType{},
			val: testtypes.Number{
				CreatedBy: testtypes.NumberType{},
			},
			expected: testtypes.Number{
				CreatedBy: testtypes.NumberType{},
			},
		},
		"NumberTypable-NumberValue": {
			typ:      testtypes.NumberTypeWithSemanticEquals{},
			val:      types.NumberNull(),
			expected: types.NumberNull(),
		},
		"ObjectType-ObjectValue-matching-attributes": {
			typ:      types.ObjectType{AttrTypes: map[string]attr.Type{"test_attr": types.StringType}},
			val:      types.ObjectNull(map[string]attr.Type{"test_attr": types.StringType}),
			expected: types.ObjectNull(map[string]attr.Type{"test_attr": types.StringType}),
		},
		"ObjectType-ObjectValue-mismatching-attributes": {
			typ:      types.ObjectType{AttrTypes: map[string]attr.Type{"test_attr": types.StringType}},
			val:      types.ObjectNull(map[string]attr.Type{"not_test_attr": types.StringType}),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.ObjectType[\"test_attr\":basetypes.StringType] / underlying type: tftypes.Object[\"test_attr\":tftypes.String]\n"+
						"Received framework type from provider logic: types.ObjectType[\"not_test_attr\":basetypes.StringType] / underlying type: tftypes.Object[\"not_test_attr\":tftypes.String]\n"+
						"Path: test",
				),
			},
		},
		"ObjectType-ObjectValue-mismatching-attribute-types": {
			typ:      types.ObjectType{AttrTypes: map[string]attr.Type{"test_attr": types.StringType}},
			val:      types.ObjectNull(map[string]attr.Type{"test_attr": types.BoolType}),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.ObjectType[\"test_attr\":basetypes.StringType] / underlying type: tftypes.Object[\"test_attr\":tftypes.String]\n"+
						"Received framework type from provider logic: types.ObjectType[\"test_attr\":basetypes.BoolType] / underlying type: tftypes.Object[\"test_attr\":tftypes.Bool]\n"+
						"Path: test",
				),
			},
		},
		"ObjectTypable-ObjectValue-matching-attributes": {
			typ: testtypes.ObjectType{
				ObjectType: types.ObjectType{AttrTypes: map[string]attr.Type{"test_attr": types.StringType}},
			},
			val:      types.ObjectNull(map[string]attr.Type{"test_attr": types.StringType}),
			expected: types.ObjectNull(map[string]attr.Type{"test_attr": types.StringType}),
		},
		"SetType-SetValue-matching-elements": {
			typ:      types.SetType{ElemType: types.StringType},
			val:      types.SetNull(types.StringType),
			expected: types.SetNull(types.StringType),
		},
		"SetType-SetValue-mismatching-elements": {
			typ:      types.SetType{ElemType: types.StringType},
			val:      types.SetNull(types.BoolType),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.SetType[basetypes.StringType] / underlying type: tftypes.Set[tftypes.String]\n"+
						"Received framework type from provider logic: types.SetType[basetypes.BoolType] / underlying type: tftypes.Set[tftypes.Bool]\n"+
						"Path: test",
				),
			},
		},
		"SetTypable-SetValue-matching-elements": {
			typ: testtypes.SetType{
				SetType: types.SetType{ElemType: types.StringType},
			},
			val:      types.SetNull(types.StringType),
			expected: types.SetNull(types.StringType),
		},
		"StringType-StringValue-null": {
			typ:      types.StringType,
			val:      types.StringNull(),
			expected: types.StringNull(),
		},
		"StringType-StringValue-unknown": {
			typ:      types.StringType,
			val:      types.StringUnknown(),
			expected: types.StringUnknown(),
		},
		"StringType-StringValue-value": {
			typ:      types.StringType,
			val:      types.StringValue("hello, world"),
			expected: types.StringValue("hello, world"),
		},
		"StringTypable-StringValuable": {
			typ: testtypes.StringType{},
			val: testtypes.String{
				CreatedBy: testtypes.StringType{},
			},
			expected: testtypes.String{
				CreatedBy: testtypes.StringType{},
			},
		},
		"StringTypable-StringValue": {
			typ:      testtypes.StringTypeWithSemanticEquals{},
			val:      types.StringNull(),
			expected: types.StringNull(),
		},
		"TupleType-TupleValue-matching-elements": {
			typ:      types.TupleType{ElemTypes: []attr.Type{types.StringType, types.BoolType}},
			val:      types.TupleNull([]attr.Type{types.StringType, types.BoolType}),
			expected: types.TupleNull([]attr.Type{types.StringType, types.BoolType}),
		},
		"TupleType-TupleValue-mismatching-elements": {
			typ:      types.TupleType{ElemTypes: []attr.Type{types.StringType, types.BoolType}},
			val:      types.TupleNull([]attr.Type{types.BoolType, types.StringType}),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Expected framework type from provider logic: types.TupleType[basetypes.StringType, basetypes.BoolType] / underlying type: tftypes.Tuple[tftypes.String, tftypes.Bool]\n"+
						"Received framework type from provider logic: types.TupleType[basetypes.BoolType, basetypes.StringType] / underlying type: tftypes.Tuple[tftypes.Bool, tftypes.String]\n"+
						"Path: test",
				),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromAttributeValue(context.Background(), tc.typ, tc.val, path.Root("test"))

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestNewValueConverter(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		val           tftypes.Value
		target        reflect.Value
		expected      *valueConverter
		expectedDiags diag.Diagnostics
	}{
		"unknown": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			target:   reflect.ValueOf(new(valueConverter)),
			expected: &valueConverter{unknown: true},
		},
		"null": {
			val:      tftypes.NewValue(tftypes.String, nil),
			target:   reflect.ValueOf(new(valueConverter)),
			expected: &valueConverter{null: true},
		},
		"value": {
			val:      tftypes.NewValue(tftypes.String, "hello"),
			target:   reflect.ValueOf(new(valueConverter)),
			expected: &valueConverter{value: "hello"},
		},
		"error": {
			val:    tftypes.NewValue(tftypes.String, "hello"),
			target: reflect.ValueOf(new(valueConverterError)),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert into a value. This is always an error in the provider. Please report the following to the provider developer:\n\nreflection error: this is an error",
				),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res, diags := refl.NewValueConverter(context.Background(), types.StringType, tc.val, tc.target, refl.Options{}, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diags.HasError() {
				return
			}

			got, ok := res.Interface().(*valueConverter)
			if !ok {
				t.Fatalf("Expected type of *valueConverter, got %T", res.Interface())
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFromValueCreator(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.Type
		vc            *valueCreator
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"null": {
			typ: types.StringType,
			vc: &valueCreator{
				null: true,
			},
			expected: types.StringNull(),
		},
		"unknown": {
			typ: types.StringType,
			vc: &valueCreator{
				unknown: true,
			},
			expected: types.StringUnknown(),
		},
		"value": {
			typ: types.StringType,
			vc: &valueCreator{
				value: "hello, world",
			},
			expected: types.StringValue("hello, world"),
		},
		"validate-error": {
			typ: testtypes.StringTypeWithValidateError{},
			vc: &valueCreator{
				value: "hello, world",
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
		"validate-attribute-error": {
			typ: testtypes.StringTypeWithValidateAttributeError{},
			vc: &valueCreator{
				value: "hello, world",
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "Error Diagnostic", "This is an error."),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromValueCreator(context.Background(), tc.typ, tc.vc, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}
