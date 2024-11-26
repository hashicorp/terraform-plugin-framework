// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/refinement"
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
)

var (
	_ StringValuable                  = StringValue{}
	_ attr.ValueWithNotNullRefinement = StringValue{}
)

// StringValuable extends attr.Value for string value types.
// Implement this interface to create a custom String value type.
type StringValuable interface {
	attr.Value

	// ToStringValue should convert the value type to a String.
	ToStringValue(ctx context.Context) (StringValue, diag.Diagnostics)
}

// StringValuableWithSemanticEquals extends StringValuable with semantic
// equality logic.
type StringValuableWithSemanticEquals interface {
	StringValuable

	// StringSemanticEquals should return true if the given value is
	// semantically equal to the current value. This logic is used to prevent
	// Terraform data consistency errors and resource drift where a value change
	// may have inconsequential differences, such as spacing character removal
	// in JSON formatted strings.
	//
	// Only known values are compared with this method as changing a value's
	// state implicitly represents a different value.
	StringSemanticEquals(context.Context, StringValuable) (bool, diag.Diagnostics)
}

// NewStringNull creates a String with a null value. Determine whether the value is
// null via the String type IsNull method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func NewStringNull() StringValue {
	return StringValue{
		state: attr.ValueStateNull,
	}
}

// NewStringUnknown creates a String with an unknown value. Determine whether the
// value is unknown via the String type IsUnknown method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func NewStringUnknown() StringValue {
	return StringValue{
		state: attr.ValueStateUnknown,
	}
}

// NewStringValue creates a String with a known value. Access the value via the String
// type ValueString method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func NewStringValue(value string) StringValue {
	return StringValue{
		state: attr.ValueStateKnown,
		value: value,
	}
}

// NewStringPointerValue creates a String with a null value if nil or a known
// value. Access the value via the String type ValueStringPointer method.
func NewStringPointerValue(value *string) StringValue {
	if value == nil {
		return NewStringNull()
	}

	return NewStringValue(*value)
}

// StringValue represents a UTF-8 string value.
type StringValue struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value string

	// refinements represents the unknown value refinement data associated with this Value.
	// This field is only populated for unknown values.
	refinements refinement.Refinements
}

// Type returns a StringType.
func (s StringValue) Type(_ context.Context) attr.Type {
	return StringType{}
}

// ToTerraformValue returns the data contained in the *String as a tftypes.Value.
func (s StringValue) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	switch s.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(tftypes.String, s.value); err != nil {
			return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.String, s.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.String, nil), nil
	case attr.ValueStateUnknown:
		if len(s.refinements) == 0 {
			return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), nil
		}

		unknownValRefinements := make(tfrefinement.Refinements, 0)
		for _, refn := range s.refinements {
			switch refnVal := refn.(type) {
			case refinement.NotNull:
				unknownValRefinements[tfrefinement.KeyNullness] = tfrefinement.NewNullness(false)
			case refinement.StringPrefix:
				unknownValRefinements[tfrefinement.KeyStringPrefix] = tfrefinement.NewStringPrefix(refnVal.PrefixValue())
			}
		}
		unknownVal := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)

		return unknownVal.Refine(unknownValRefinements), nil
	default:
		panic(fmt.Sprintf("unhandled String state in ToTerraformValue: %s", s.state))
	}
}

// Equal returns true if `other` is a String and has the same value as `s`.
func (s StringValue) Equal(other attr.Value) bool {
	o, ok := other.(StringValue)

	if !ok {
		return false
	}

	if s.state != o.state {
		return false
	}

	if len(s.refinements) != len(o.refinements) {
		return false
	}

	if len(s.refinements) > 0 && !s.refinements.Equal(o.refinements) {
		return false
	}

	if s.state != attr.ValueStateKnown {
		return true
	}

	return s.value == o.value
}

// IsNull returns true if the String represents a null value.
func (s StringValue) IsNull() bool {
	return s.state == attr.ValueStateNull
}

// IsUnknown returns true if the String represents a currently unknown value.
func (s StringValue) IsUnknown() bool {
	return s.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the String value. Use
// the ValueString method for Terraform data handling instead.
//
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (s StringValue) String() string {
	if s.IsUnknown() {
		if len(s.refinements) == 0 {
			return attr.UnknownValueString
		}

		return fmt.Sprintf("<unknown, %s>", s.refinements.String())
	}

	if s.IsNull() {
		return attr.NullValueString
	}

	return fmt.Sprintf("%q", s.value)
}

// ValueString returns the known string value. If String is null or unknown, returns
// "".
func (s StringValue) ValueString() string {
	return s.value
}

// ValueStringPointer returns a pointer to the known string value, nil for a
// null value, or a pointer to "" for an unknown value.
func (s StringValue) ValueStringPointer() *string {
	if s.IsNull() {
		return nil
	}

	return &s.value
}

// ToStringValue returns String.
func (s StringValue) ToStringValue(context.Context) (StringValue, diag.Diagnostics) {
	return s, nil
}

// RefineAsNotNull will return a new unknown StringValue that includes a value refinement that:
//   - Indicates the string value will not be null once it becomes known.
//
// If the provided StringValue is null or known, then the StringValue will be returned unchanged.
func (s StringValue) RefineAsNotNull() StringValue {
	if !s.IsUnknown() {
		return s
	}

	newRefinements := make(refinement.Refinements, len(s.refinements))
	for i, refn := range s.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()

	newUnknownVal := NewStringUnknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// RefineWithPrefix will return an unknown StringValue that includes a value refinement that:
//   - Indicates the string value will not be null once it becomes known.
//   - Indicates the string value will have the specified prefix once it becomes known.
//
// Prefixes that exceed 256 characters in length will be truncated and empty string prefixes
// will be ignored. If the provided StringValue is null or known, then the StringValue will be
// returned unchanged.
func (s StringValue) RefineWithPrefix(prefix string) StringValue {
	if !s.IsUnknown() {
		return s
	}

	newRefinements := make(refinement.Refinements, len(s.refinements))
	for i, refn := range s.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()

	// No need to encode an empty prefix, since terraform-plugin-go will ignore it anyways.
	if prefix != "" {
		newRefinements[refinement.KeyStringPrefix] = refinement.NewStringPrefix(prefix)
	}

	newUnknownVal := NewStringUnknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// NotNullRefinement returns value refinement data and a boolean indicating if a NotNull refinement
// exists on the given StringValue. If a StringValue contains a NotNull refinement, this indicates
// that the string is unknown, but the eventual known value will not be null.
//
// A NotNull value refinement can be added to an unknown value via the `RefineAsNotNull` method.
func (s StringValue) NotNullRefinement() (*refinement.NotNull, bool) {
	if !s.IsUnknown() {
		return nil, false
	}

	refn, ok := s.refinements[refinement.KeyNotNull]
	if !ok {
		return nil, false
	}

	notNullRefn, ok := refn.(refinement.NotNull)
	if !ok {
		return nil, false
	}

	return &notNullRefn, true
}

// PrefixRefinement returns value refinement data and a boolean indicating if a StringPrefix refinement
// exists on the given StringValue. If a StringValue contains a StringPrefix refinement, this indicates
// that the string is unknown, but the eventual known value will have a specified string prefix.
// The returned boolean should be checked before accessing refinement data.
//
// A StringPrefix value refinement can be added to an unknown value via the `RefineWithPrefix` method.
func (s StringValue) PrefixRefinement() (*refinement.StringPrefix, bool) {
	if !s.IsUnknown() {
		return nil, false
	}

	refn, ok := s.refinements[refinement.KeyStringPrefix]
	if !ok {
		return nil, false
	}

	prefixRefn, ok := refn.(refinement.StringPrefix)
	if !ok {
		return nil, false
	}

	return &prefixRefn, true
}
