package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

var (
	_ attr.Value = String{}
)

// StringNull creates a String with a null value. Determine whether the value is
// null via the String type IsNull method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func StringNull() String {
	return String{
		state: valueStateNull,
	}
}

// StringUnknown creates a String with an unknown value. Determine whether the
// value is unknown via the String type IsUnknown method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func StringUnknown() String {
	return String{
		state: valueStateUnknown,
	}
}

// StringValue creates a String with a known value. Access the value via the String
// type ValueString method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func StringValue(value string) String {
	return String{
		state: valueStateKnown,
		value: value,
	}
}

func stringValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return String{
			Unknown: true,
			state:   valueStateDeprecated,
		}, nil
	}
	if in.IsNull() {
		return String{
			Null:  true,
			state: valueStateDeprecated,
		}, nil
	}
	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}
	return String{
		Value: s,
		state: valueStateDeprecated,
	}, nil
}

// String represents a UTF-8 string value.
type String struct {
	// Unknown will be true if the value is not yet known.
	//
	// If the String was created with the StringValue, StringNull, or StringUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the StringUnknown function to create an unknown String
	// value or use the IsUnknown method to determine whether the String value
	// is unknown instead.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	//
	// If the String was created with the StringValue, StringNull, or StringUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the StringNull function to create a null String value or
	// use the IsNull method to determine whether the String value is null
	// instead.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	//
	// If the String was created with the StringValue, StringNull, or StringUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the StringValue function to create a known String value or
	// use the ValueString method to retrieve the String value instead.
	Value string

	// state represents whether the String is null, unknown, or known. During the
	// exported field deprecation period, this state can also be "deprecated",
	// which remains the zero-value for compatibility to ensure exported field
	// updates take effect. The zero-value will be changed to null in a future
	// version.
	state valueState

	// value contains the known value, if not null or unknown.
	value string
}

func (s String) ToFrameworkValue() attr.Value {
	return s
}

// Type returns a StringType.
func (s String) Type(_ context.Context) attr.Type {
	return StringType
}

// ToTerraformValue returns the data contained in the *String as a tftypes.Value.
func (s String) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	switch s.state {
	case valueStateDeprecated:
		if s.Null {
			return tftypes.NewValue(tftypes.String, nil), nil
		}
		if s.Unknown {
			return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), nil
		}
		if err := tftypes.ValidateValue(tftypes.String, s.Value); err != nil {
			return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(tftypes.String, s.Value), nil
	case valueStateKnown:
		if err := tftypes.ValidateValue(tftypes.String, s.value); err != nil {
			return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.String, s.value), nil
	case valueStateNull:
		return tftypes.NewValue(tftypes.String, nil), nil
	case valueStateUnknown:
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
	if s.state == valueStateKnown {
		return s.value == o.value
	}
	if s.Unknown != o.Unknown {
		return false
	}
	if s.Null != o.Null {
		return false
	}
	return s.Value == o.Value
}

// IsNull returns true if the String represents a null value.
func (s String) IsNull() bool {
	if s.state == valueStateNull {
		return true
	}

	return s.state == valueStateDeprecated && s.Null
}

// IsUnknown returns true if the String represents a currently unknown value.
func (s String) IsUnknown() bool {
	if s.state == valueStateUnknown {
		return true
	}

	return s.state == valueStateDeprecated && s.Unknown
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

	if s.state == valueStateKnown {
		return fmt.Sprintf("%q", s.value)
	}

	return fmt.Sprintf("%q", s.Value)
}

// ValueString returns the known string value. If String is null or unknown, returns
// "".
func (s String) ValueString() string {
	if s.state == valueStateDeprecated {
		return s.Value
	}

	return s.value
}
