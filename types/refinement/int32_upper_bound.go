// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// Int32UpperBound represents an unknown value refinement that indicates the final value will not be greater than the specified
// int32 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Int32.
type Int32UpperBound struct {
	inclusive bool
	value     int32
}

func (i Int32UpperBound) Equal(other Refinement) bool {
	otherVal, ok := other.(Int32UpperBound)
	if !ok {
		return false
	}

	return i.IsInclusive() == otherVal.IsInclusive() && i.UpperBound() == otherVal.UpperBound()
}

func (i Int32UpperBound) String() string {
	rangeDescription := "inclusive"
	if !i.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("upper bound = %d (%s)", i.UpperBound(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `UpperBound` method is inclusive or exclusive.
func (i Int32UpperBound) IsInclusive() bool {
	return i.inclusive
}

// UpperBound returns the int32 value that the final value will not be greater than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (i Int32UpperBound) UpperBound() int32 {
	return i.value
}

func (i Int32UpperBound) unimplementable() {}

// NewInt32UpperBound returns the Int32UpperBound unknown value refinement that indicates the final value will not be greater than the specified
// int32 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Int32.
func NewInt32UpperBound(value int32, inclusive bool) Refinement {
	return Int32UpperBound{
		value:     value,
		inclusive: inclusive,
	}
}
