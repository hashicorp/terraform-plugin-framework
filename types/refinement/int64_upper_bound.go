// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// Int64UpperBound represents an unknown value refinement that indicates the final value will not be greater than the specified
// int64 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Int64.
type Int64UpperBound struct {
	inclusive bool
	value     int64
}

func (i Int64UpperBound) Equal(other Refinement) bool {
	otherVal, ok := other.(Int64UpperBound)
	if !ok {
		return false
	}

	return i.IsInclusive() == otherVal.IsInclusive() && i.UpperBound() == otherVal.UpperBound()
}

func (i Int64UpperBound) String() string {
	rangeDescription := "inclusive"
	if !i.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("upper bound = %d (%s)", i.UpperBound(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `UpperBound` method is inclusive or exclusive.
func (i Int64UpperBound) IsInclusive() bool {
	return i.inclusive
}

// UpperBound returns the int64 value that the final value will not be greater than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (i Int64UpperBound) UpperBound() int64 {
	return i.value
}

func (i Int64UpperBound) unimplementable() {}

// NewInt64UpperBound returns the Int64UpperBound unknown value refinement that indicates the final value will not be greater than the specified
// int64 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Int64.
func NewInt64UpperBound(value int64, inclusive bool) Refinement {
	return Int64UpperBound{
		value:     value,
		inclusive: inclusive,
	}
}
