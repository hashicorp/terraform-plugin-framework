package types

import (
	"context"

	tfsdk "github.com/hashicorp/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func stringValueFromTerraform(ctx context.Context, in tftypes.Value) (tfsdk.AttributeValue, error) {
	s := new(String)
	err := s.SetTerraformValue(ctx, in)
	return s, err
}

// String represents a UTF-8 string value.
type String struct {
	// Unknown will be true if the value is not yet known.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	Value string
}

// ToTerraformValue returns the data contained in the AttributeValue as
// a Go type that tftypes.NewValue will accept.
func (s *String) ToTerraformValue(_ context.Context) (interface{}, error) {
	if s.Null {
		return nil, nil
	}
	if s.Unknown {
		return tftypes.UnknownValue, nil
	}
	return s.Value, nil
}

// Equal must return true if the AttributeValue is considered
// semantically equal to the AttributeValue passed as an argument.
func (s *String) Equal(other tfsdk.AttributeValue) bool {
	o, ok := other.(*String)
	if !ok {
		return false
	}
	return s.Value == o.Value
}

// SetTerraformValue updates `s` to reflect the data stored in `in`.
func (s *String) SetTerraformValue(ctx context.Context, in tftypes.Value) error {
	s.Unknown = false
	s.Null = false
	s.Value = ""
	if !in.IsKnown() {
		s.Unknown = true
		return nil
	}
	if in.IsNull() {
		s.Null = true
		return nil
	}
	err := in.As(&s.Value)
	return err
}
