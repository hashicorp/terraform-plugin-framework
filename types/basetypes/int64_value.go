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
	tfrefinements "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
)

var (
	_ Int64Valuable                   = Int64Value{}
	_ attr.ValueWithNotNullRefinement = Int64Value{}
)

// Int64Valuable extends attr.Value for int64 value types.
// Implement this interface to create a custom Int64 value type.
type Int64Valuable interface {
	attr.Value

	// ToInt64Value should convert the value type to an Int64.
	ToInt64Value(ctx context.Context) (Int64Value, diag.Diagnostics)
}

// Int64ValuableWithSemanticEquals extends Int64Valuable with semantic
// equality logic.
type Int64ValuableWithSemanticEquals interface {
	Int64Valuable

	// Int64SemanticEquals should return true if the given value is
	// semantically equal to the current value. This logic is used to prevent
	// Terraform data consistency errors and resource drift where a value change
	// may have inconsequential differences, such as rounding.
	//
	// Only known values are compared with this method as changing a value's
	// state implicitly represents a different value.
	Int64SemanticEquals(context.Context, Int64Valuable) (bool, diag.Diagnostics)
}

// NewInt64Null creates a Int64 with a null value. Determine whether the value is
// null via the Int64 type IsNull method.
func NewInt64Null() Int64Value {
	return Int64Value{
		state: attr.ValueStateNull,
	}
}

// NewInt64Unknown creates a Int64 with an unknown value. Determine whether the
// value is unknown via the Int64 type IsUnknown method.
func NewInt64Unknown() Int64Value {
	return Int64Value{
		state: attr.ValueStateUnknown,
	}
}

// NewInt64Value creates a Int64 with a known value. Access the value via the Int64
// type ValueInt64 method.
func NewInt64Value(value int64) Int64Value {
	return Int64Value{
		state: attr.ValueStateKnown,
		value: value,
	}
}

// NewInt64PointerValue creates a Int64 with a null value if nil or a known
// value. Access the value via the Int64 type ValueInt64Pointer method.
func NewInt64PointerValue(value *int64) Int64Value {
	if value == nil {
		return NewInt64Null()
	}

	return NewInt64Value(*value)
}

// Int64Value represents a 64-bit integer value, exposed as an int64.
type Int64Value struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value int64

	// TODO: doc
	refinements refinement.Refinements
}

// Equal returns true if `other` is an Int64 and has the same value as `i`.
func (i Int64Value) Equal(other attr.Value) bool {
	o, ok := other.(Int64Value)

	if !ok {
		return false
	}

	if i.state != o.state {
		return false
	}

	if i.state != attr.ValueStateKnown {
		return true
	}

	// TODO: compare refinements? I might not be able to... to allow future refinements?

	return i.value == o.value
}

// ToTerraformValue returns the data contained in the Int64 as a tftypes.Value.
func (i Int64Value) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
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

		unknownValRefinements := make(tfrefinements.Refinements, 0)
		for _, refn := range i.refinements {
			switch refnVal := refn.(type) {
			case refinement.NotNull:
				unknownValRefinements[tfrefinements.KeyNullness] = tfrefinements.NewNullness(false)
			case refinement.NumberLowerBound:
				unknownValRefinements[tfrefinements.KeyNumberLowerBound] = tfrefinements.NewNumberLowerBound(refnVal.LowerBound(), refnVal.IsInclusive())
			case refinement.NumberUpperBound:
				unknownValRefinements[tfrefinements.KeyNumberUpperBound] = tfrefinements.NewNumberUpperBound(refnVal.UpperBound(), refnVal.IsInclusive())
			}
		}
		unknownVal := tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)

		return unknownVal.Refine(unknownValRefinements), nil
	default:
		panic(fmt.Sprintf("unhandled Int64 state in ToTerraformValue: %s", i.state))
	}
}

// Type returns a Int64Type.
func (i Int64Value) Type(ctx context.Context) attr.Type {
	return Int64Type{}
}

// IsNull returns true if the Int64 represents a null value.
func (i Int64Value) IsNull() bool {
	return i.state == attr.ValueStateNull
}

// IsUnknown returns true if the Int64 represents a currently unknown value.
func (i Int64Value) IsUnknown() bool {
	return i.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the Int64 value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (i Int64Value) String() string {
	if i.IsUnknown() {
		// TODO: Also print out unknown value refinements?
		return attr.UnknownValueString
	}

	if i.IsNull() {
		return attr.NullValueString
	}

	return fmt.Sprintf("%d", i.value)
}

// ValueInt64 returns the known int64 value. If Int64 is null or unknown, returns
// 0.
func (i Int64Value) ValueInt64() int64 {
	return i.value
}

// ValueInt64Pointer returns a pointer to the known int64 value, nil for a
// null value, or a pointer to 0 for an unknown value.
func (i Int64Value) ValueInt64Pointer() *int64 {
	if i.IsNull() {
		return nil
	}

	return &i.value
}

// ToInt64Value returns Int64.
func (i Int64Value) ToInt64Value(context.Context) (Int64Value, diag.Diagnostics) {
	return i, nil
}

// RefineAsNotNull will return an unknown Int64Value that includes a value refinement that:
//   - Indicates the int64 value will not be null once it becomes known.
//
// If the Int64Value is not unknown, then no refinement will be added and the provided Int64Value will be returned.
func (i Int64Value) RefineAsNotNull() Int64Value {
	// TODO: Should we return an error?
	if !i.IsUnknown() {
		return i
	}

	// TODO: Do I need to do a full copy of this map? Do we need to copy any of this at all? Since it's operating on the value struct?
	refns := make(refinement.Refinements, len(i.refinements))
	for i, refn := range i.refinements {
		refns[i] = refn
	}
	refns[refinement.KeyNotNull] = refinement.NewNotNull()

	newUnknownVal := NewInt64Unknown()
	newUnknownVal.refinements = refns

	return newUnknownVal
}

// RefineWithLowerBound will return an unknown Int64Value that includes a value refinement that:
//   - Indicates the int64 value will not be null once it becomes known.
//   - Indicates the int64 value will not be less than the int64 provided (lowerBound) once it becomes known.
//
// If the Int64Value is not unknown, then no refinement will be added and the provided Int64Value will be returned.
func (i Int64Value) RefineWithLowerBound(lowerBound int64, inclusive bool) Int64Value {
	// TODO: Should we return an error?
	if !i.IsUnknown() {
		return i
	}

	// TODO: Do I need to do a full copy of this map? Do we need to copy any of this at all? Since it's operating on the value struct?
	refns := make(refinement.Refinements, len(i.refinements))
	for i, refn := range i.refinements {
		refns[i] = refn
	}
	refns[refinement.KeyNotNull] = refinement.NewNotNull()
	refns[refinement.KeyNumberLowerBound] = refinement.NewNumberLowerBound(new(big.Float).SetInt64(lowerBound), inclusive)

	newUnknownVal := NewInt64Unknown()
	newUnknownVal.refinements = refns

	return newUnknownVal
}

// RefineWithUpperBound will return an unknown Int64Value that includes a value refinement that:
//   - Indicates the int64 value will not be null once it becomes known.
//   - Indicates the int64 value will not be greater than the int64 provided (upperBound) once it becomes known.
//
// If the Int64Value is not unknown, then no refinement will be added and the provided Int64Value will be returned.
func (i Int64Value) RefineWithUpperBound(upperBound int64, inclusive bool) Int64Value {
	// TODO: Should we return an error?
	if !i.IsUnknown() {
		return i
	}

	// TODO: Do I need to do a full copy of this map? Do we need to copy any of this at all? Since it's operating on the value struct?
	refns := make(refinement.Refinements, len(i.refinements))
	for i, refn := range i.refinements {
		refns[i] = refn
	}
	refns[refinement.KeyNotNull] = refinement.NewNotNull()
	refns[refinement.KeyNumberUpperBound] = refinement.NewNumberUpperBound(new(big.Float).SetInt64(upperBound), inclusive)

	newUnknownVal := NewInt64Unknown()
	newUnknownVal.refinements = refns

	return newUnknownVal
}

// NotNullRefinement returns a value refinement, if one exists, that indicates an unknown int64 value
// will not be null once it becomes known.
//
// A NotNull value refinement can be added to an unknown value via the `RefineAsNotNull` method.
func (i Int64Value) NotNullRefinement() *refinement.NotNull {
	if !i.IsUnknown() {
		return nil
	}

	refn, ok := i.refinements[refinement.KeyNotNull]
	if !ok {
		return nil
	}

	notNullRefn, ok := refn.(refinement.NotNull)
	if !ok {
		return nil
	}

	return &notNullRefn
}

// LowerBoundRefinement returns a value refinement, if one exists, that indicates an unknown int64 value
// will not be less than the specified int64 value once it becomes known.
//
// A NumberLowerBound value refinement can be added to an unknown value via the `RefineWithLowerBound` method.
func (i Int64Value) LowerBoundRefinement() *refinement.NumberLowerBound {
	if !i.IsUnknown() {
		return nil
	}

	refn, ok := i.refinements[refinement.KeyNumberLowerBound]
	if !ok {
		return nil
	}

	lowerBoundRefn, ok := refn.(refinement.NumberLowerBound)
	if !ok {
		return nil
	}

	return &lowerBoundRefn
}

// UpperBoundRefinement returns a value refinement, if one exists, that indicates an unknown int64 value
// will not be greater than the specified int64 value once it becomes known.
//
// A NumberUpperBound value refinement can be added to an unknown value via the `RefineWithUpperBound` method.
func (i Int64Value) UpperBoundRefinement() *refinement.NumberUpperBound {
	if !i.IsUnknown() {
		return nil
	}

	refn, ok := i.refinements[refinement.KeyNumberUpperBound]
	if !ok {
		return nil
	}

	upperBoundRefn, ok := refn.(refinement.NumberUpperBound)
	if !ok {
		return nil
	}

	return &upperBoundRefn
}
