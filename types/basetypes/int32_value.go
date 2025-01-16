// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/refinement"
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
)

var (
	_ Int32Valuable                   = Int32Value{}
	_ attr.ValueWithNotNullRefinement = Int32Value{}
)

// Int32Valuable extends attr.Value for int32 value types.
// Implement this interface to create a custom Int32 value type.
type Int32Valuable interface {
	attr.Value

	// ToInt32Value should convert the value type to an Int32.
	ToInt32Value(ctx context.Context) (Int32Value, diag.Diagnostics)
}

// Int32ValuableWithSemanticEquals extends Int32Valuable with semantic
// equality logic.
type Int32ValuableWithSemanticEquals interface {
	Int32Valuable

	// Int32SemanticEquals should return true if the given value is
	// semantically equal to the current value. This logic is used to prevent
	// Terraform data consistency errors and resource drift where a value change
	// may have inconsequential differences, such as rounding.
	//
	// Only known values are compared with this method as changing a value's
	// state implicitly represents a different value.
	Int32SemanticEquals(context.Context, Int32Valuable) (bool, diag.Diagnostics)
}

// NewInt32Null creates an Int32 with a null value. Determine whether the value is
// null via the Int32 type IsNull method.
func NewInt32Null() Int32Value {
	return Int32Value{
		state: attr.ValueStateNull,
	}
}

// NewInt32Unknown creates an Int32 with an unknown value. Determine whether the
// value is unknown via the Int32 type IsUnknown method.
func NewInt32Unknown() Int32Value {
	return Int32Value{
		state: attr.ValueStateUnknown,
	}
}

// NewInt32Value creates an Int32 with a known value. Access the value via the Int32
// type ValueInt32 method.
func NewInt32Value(value int32) Int32Value {
	return Int32Value{
		state: attr.ValueStateKnown,
		value: value,
	}
}

// NewInt32PointerValue creates an Int32 with a null value if nil or a known
// value. Access the value via the Int32 type ValueInt32Pointer method.
func NewInt32PointerValue(value *int32) Int32Value {
	if value == nil {
		return NewInt32Null()
	}

	return NewInt32Value(*value)
}

// Int32Value represents a 32-bit integer value, exposed as an int32.
type Int32Value struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value int32

	// refinements represents the unknown value refinement data associated with this Value.
	// This field is only populated for unknown values.
	refinements refinement.Refinements
}

// Equal returns true if `other` is an Int32 and has the same value as `i`.
func (i Int32Value) Equal(other attr.Value) bool {
	o, ok := other.(Int32Value)

	if !ok {
		return false
	}

	if i.state != o.state {
		return false
	}

	if len(i.refinements) != len(o.refinements) {
		return false
	}

	if len(i.refinements) > 0 && !i.refinements.Equal(o.refinements) {
		return false
	}

	if i.state != attr.ValueStateKnown {
		return true
	}

	return i.value == o.value
}

// ToTerraformValue returns the data contained in the Int32 as a tftypes.Value.
func (i Int32Value) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	switch i.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(tftypes.Number, i.value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Number, i.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.Number, nil), nil
	case attr.ValueStateUnknown:
		if len(i.refinements) == 0 {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil
		}

		unknownValRefinements := make(tfrefinement.Refinements, 0)
		for _, refn := range i.refinements {
			switch refnVal := refn.(type) {
			case refinement.NotNull:
				unknownValRefinements[tfrefinement.KeyNullness] = tfrefinement.NewNullness(false)
			case refinement.Int32LowerBound:
				lowerBound := new(big.Float).SetInt64(int64(refnVal.LowerBound()))
				unknownValRefinements[tfrefinement.KeyNumberLowerBound] = tfrefinement.NewNumberLowerBound(lowerBound, refnVal.IsInclusive())
			case refinement.Int32UpperBound:
				upperBound := new(big.Float).SetInt64(int64(refnVal.UpperBound()))
				unknownValRefinements[tfrefinement.KeyNumberUpperBound] = tfrefinement.NewNumberUpperBound(upperBound, refnVal.IsInclusive())
			}
		}
		unknownVal := tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)

		return unknownVal.Refine(unknownValRefinements), nil
	default:
		panic(fmt.Sprintf("unhandled Int32 state in ToTerraformValue: %s", i.state))
	}
}

// Type returns a Int32Type.
func (i Int32Value) Type(ctx context.Context) attr.Type {
	return Int32Type{}
}

// IsNull returns true if the Int32 represents a null value.
func (i Int32Value) IsNull() bool {
	return i.state == attr.ValueStateNull
}

// IsUnknown returns true if the Int32 represents a currently unknown value.
func (i Int32Value) IsUnknown() bool {
	return i.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the Int32 value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (i Int32Value) String() string {
	if i.IsUnknown() {
		if len(i.refinements) == 0 {
			return attr.UnknownValueString
		}

		return fmt.Sprintf("<unknown, %s>", i.refinements.String())
	}

	if i.IsNull() {
		return attr.NullValueString
	}

	return fmt.Sprintf("%d", i.value)
}

// ValueInt32 returns the known int32 value. If Int32 is null or unknown, returns
// 0.
func (i Int32Value) ValueInt32() int32 {
	return i.value
}

// ValueInt32Pointer returns a pointer to the known int32 value, nil for a
// null value, or a pointer to 0 for an unknown value.
func (i Int32Value) ValueInt32Pointer() *int32 {
	if i.IsNull() {
		return nil
	}

	return &i.value
}

// ToInt32Value returns Int32.
func (i Int32Value) ToInt32Value(context.Context) (Int32Value, diag.Diagnostics) {
	return i, nil
}

// RefineAsNotNull will return an unknown Int32Value that includes a value refinement that:
//   - Indicates the int32 value will not be null once it becomes known.
//
// If the provided Int32Value is null or known, then the Int32Value will be returned unchanged.
func (i Int32Value) RefineAsNotNull() Int32Value {
	if !i.IsUnknown() {
		return i
	}

	newRefinements := make(refinement.Refinements, len(i.refinements))
	for i, refn := range i.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()

	newUnknownVal := NewInt32Unknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// RefineWithLowerBound will return an unknown Int32Value that includes a value refinement that:
//   - Indicates the int32 value will not be null once it becomes known.
//   - Indicates the int32 value will not be less than the int32 provided (lowerBound) once it becomes known.
//
// If the provided Int32Value is null or known, then the Int32Value will be returned unchanged.
func (i Int32Value) RefineWithLowerBound(lowerBound int32, inclusive bool) Int32Value {
	if !i.IsUnknown() {
		return i
	}

	newRefinements := make(refinement.Refinements, len(i.refinements))
	for i, refn := range i.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()
	newRefinements[refinement.KeyNumberLowerBound] = refinement.NewInt32LowerBound(lowerBound, inclusive)

	newUnknownVal := NewInt32Unknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// RefineWithUpperBound will return an unknown Int32Value that includes a value refinement that:
//   - Indicates the int32 value will not be null once it becomes known.
//   - Indicates the int32 value will not be greater than the int32 provided (upperBound) once it becomes known.
//
// If the provided Int32Value is null or known, then the Int32Value will be returned unchanged.
func (i Int32Value) RefineWithUpperBound(upperBound int32, inclusive bool) Int32Value {
	if !i.IsUnknown() {
		return i
	}

	newRefinements := make(refinement.Refinements, len(i.refinements))
	for i, refn := range i.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()
	newRefinements[refinement.KeyNumberUpperBound] = refinement.NewInt32UpperBound(upperBound, inclusive)

	newUnknownVal := NewInt32Unknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// NotNullRefinement returns value refinement data and a boolean indicating if a NotNull refinement
// exists on the given Int32Value. If an Int32Value contains a NotNull refinement, this indicates that
// the int32 value is unknown, but the eventual known value will not be null.
//
// A NotNull value refinement can be added to an unknown value via the `RefineAsNotNull` method.
func (i Int32Value) NotNullRefinement() (*refinement.NotNull, bool) {
	if !i.IsUnknown() {
		return nil, false
	}

	refn, ok := i.refinements[refinement.KeyNotNull]
	if !ok {
		return nil, false
	}

	notNullRefn, ok := refn.(refinement.NotNull)
	if !ok {
		return nil, false
	}

	return &notNullRefn, true
}

// LowerBoundRefinement returns value refinement data and a boolean indicating if a Int32LowerBound refinement
// exists on the given Int32Value. If an Int32Value contains a Int32LowerBound refinement, this indicates that
// the int32 value is unknown, but the eventual known value will not be less than the specified int32 value
// (either inclusive or exclusive) once it becomes known. The returned boolean should be checked before accessing
// refinement data.
//
// An Int32LowerBound value refinement can be added to an unknown value via the `RefineWithLowerBound` method.
func (i Int32Value) LowerBoundRefinement() (*refinement.Int32LowerBound, bool) {
	if !i.IsUnknown() {
		return nil, false
	}

	refn, ok := i.refinements[refinement.KeyNumberLowerBound]
	if !ok {
		return nil, false
	}

	lowerBoundRefn, ok := refn.(refinement.Int32LowerBound)
	if !ok {
		return nil, false
	}

	return &lowerBoundRefn, true
}

// UpperBoundRefinement returns value refinement data and a boolean indicating if a Int32UpperBound refinement
// exists on the given Int32Value. If an Int32Value contains a Int32UpperBound refinement, this indicates that
// the int32 value is unknown, but the eventual known value will not be greater than the specified int32 value
// (either inclusive or exclusive) once it becomes known. The returned boolean should be checked before accessing
// refinement data.
//
// A Int32UpperBound value refinement can be added to an unknown value via the `RefineWithUpperBound` method.
func (i Int32Value) UpperBoundRefinement() (*refinement.Int32UpperBound, bool) {
	if !i.IsUnknown() {
		return nil, false
	}

	refn, ok := i.refinements[refinement.KeyNumberUpperBound]
	if !ok {
		return nil, false
	}

	upperBoundRefn, ok := refn.(refinement.Int32UpperBound)
	if !ok {
		return nil, false
	}

	return &upperBoundRefn, true
}
