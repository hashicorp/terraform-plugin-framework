// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// Float64UpperBound represents an unknown value refinement that indicates the final value will not be greater than the specified
// float64 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Float64.
type Float64UpperBound struct {
	inclusive bool
	value     float64
}

func (i Float64UpperBound) Equal(other Refinement) bool {
	otherVal, ok := other.(Float64UpperBound)
	if !ok {
		return false
	}

	return i.IsInclusive() == otherVal.IsInclusive() && i.UpperBound() == otherVal.UpperBound()
}

func (i Float64UpperBound) String() string {
	rangeDescription := "inclusive"
	if !i.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("upper bound = %f (%s)", i.UpperBound(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `UpperBound` method is inclusive or exclusive.
func (i Float64UpperBound) IsInclusive() bool {
	return i.inclusive
}

// UpperBound returns the float64 value that the final value will not be greater than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (i Float64UpperBound) UpperBound() float64 {
	return i.value
}

func (i Float64UpperBound) unimplementable() {}

// NewFloat64UpperBound returns the Float64UpperBound unknown value refinement that indicates the final value will not be greater than the specified
// float64 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Float64.
func NewFloat64UpperBound(value float64, inclusive bool) Refinement {
	return Float64UpperBound{
		value:     value,
		inclusive: inclusive,
	}
}
