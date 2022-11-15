package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	_ StringValuable = String{}
)

// StringValuable extends attr.Value for string value types.
// Implement this interface to create a custom String value type.
type StringValuable interface {
	attr.Value

	// ToStringValue should convert the value type to a String.
	ToStringValue(ctx context.Context) (String, diag.Diagnostics)
}

// StringNull creates a String with a null value. Determine whether the value is
// null via the String type IsNull method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func StringNull() String {
	return String{
		state: attr.ValueStateNull,
	}
}

// StringUnknown creates a String with an unknown value. Determine whether the
// value is unknown via the String type IsUnknown method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func StringUnknown() String {
	return String{
		state: attr.ValueStateUnknown,
	}
}

// StringValue creates a String with a known value. Access the value via the String
// type ValueString method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func StringValue(value string) String {
	return String{
		state: attr.ValueStateKnown,
		value: value,
	}
}

func stringValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return StringUnknown(), nil
	}
	if in.IsNull() {
		return StringNull(), nil
	}
	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}
	return StringValue(s), nil
}

// String represents a UTF-8 string value.
type String struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value string
}

// Type returns a StringType.
func (s String) Type(_ context.Context) attr.Type {
	return StringType
}

// ToTerraformValue returns the data contained in the *String as a tftypes.Value.
func (s String) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	switch s.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(tftypes.String, s.value); err != nil {
			return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.String, s.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.String, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled String state in ToTerraformValue: %s", s.state))
	}
}

// Equal returns true if `other` is a String and has the same value as `s`.
func (s String) Equal(other attr.Value) bool {
	o, ok := other.(String)

	if !ok {
		return false
	}

	if s.state != o.state {
		return false
	}

	if s.state != attr.ValueStateKnown {
		return true
	}

	return s.value == o.value
}

// IsNull returns true if the String represents a null value.
func (s String) IsNull() bool {
	return s.state == attr.ValueStateNull
}

// IsUnknown returns true if the String represents a currently unknown value.
func (s String) IsUnknown() bool {
	return s.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the String value. Use
// the ValueString method for Terraform data handling instead.
//
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (s String) String() string {
	if s.IsUnknown() {
		return attr.UnknownValueString
	}

	if s.IsNull() {
		return attr.NullValueString
	}

	return fmt.Sprintf("%q", s.value)
}

// ValueString returns the known string value. If String is null or unknown, returns
// "".
func (s String) ValueString() string {
	return s.value
}

// ToStringValue returns String.
func (s String) ToStringValue(context.Context) (String, diag.Diagnostics) {
	return s, nil
}
