// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// Int64LowerBound represents an unknown value refinement that indicates the final value will not be less than the specified
// int64 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Int64.
type Int64LowerBound struct {
	inclusive bool
	value     int64
}

func (i Int64LowerBound) Equal(other Refinement) bool {
	otherVal, ok := other.(Int64LowerBound)
	if !ok {
		return false
	}

	return i.IsInclusive() == otherVal.IsInclusive() && i.LowerBound() == otherVal.LowerBound()
}

func (i Int64LowerBound) String() string {
	rangeDescription := "inclusive"
	if !i.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("lower bound = %d (%s)", i.LowerBound(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `LowerBound` method is inclusive or exclusive.
func (i Int64LowerBound) IsInclusive() bool {
	return i.inclusive
}

// LowerBound returns the int64 value that the final value will not be less than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (i Int64LowerBound) LowerBound() int64 {
	return i.value
}

func (i Int64LowerBound) unimplementable() {}

// NewInt64LowerBound returns the Int64LowerBound unknown value refinement that indicates the final value will not be less than the specified
// int64 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Int64.
func NewInt64LowerBound(value int64, inclusive bool) Refinement {
	return Int64LowerBound{
		value:     value,
		inclusive: inclusive,
	}
}
