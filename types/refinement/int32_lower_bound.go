// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// Int32LowerBound represents an unknown value refinement that indicates the final value will not be less than the specified
// int32 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Int32.
type Int32LowerBound struct {
	inclusive bool
	value     int32
}

func (i Int32LowerBound) Equal(other Refinement) bool {
	otherVal, ok := other.(Int32LowerBound)
	if !ok {
		return false
	}

	return i.IsInclusive() == otherVal.IsInclusive() && i.LowerBound() == otherVal.LowerBound()
}

func (i Int32LowerBound) String() string {
	rangeDescription := "inclusive"
	if !i.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("lower bound = %d (%s)", i.LowerBound(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `LowerBound` method is inclusive or exclusive.
func (i Int32LowerBound) IsInclusive() bool {
	return i.inclusive
}

// LowerBound returns the int32 value that the final value will not be less than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (i Int32LowerBound) LowerBound() int32 {
	return i.value
}

func (i Int32LowerBound) unimplementable() {}

// NewInt32LowerBound returns the Int32LowerBound unknown value refinement that indicates the final value will not be less than the specified
// int32 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Int32.
func NewInt32LowerBound(value int32, inclusive bool) Refinement {
	return Int32LowerBound{
		value:     value,
		inclusive: inclusive,
	}
}
