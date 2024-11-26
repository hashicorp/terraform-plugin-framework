// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// Float64LowerBound represents an unknown value refinement that indicates the final value will not be less than the specified
// float64 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Float64.
type Float64LowerBound struct {
	inclusive bool
	value     float64
}

func (i Float64LowerBound) Equal(other Refinement) bool {
	otherVal, ok := other.(Float64LowerBound)
	if !ok {
		return false
	}

	return i.IsInclusive() == otherVal.IsInclusive() && i.LowerBound() == otherVal.LowerBound()
}

func (i Float64LowerBound) String() string {
	rangeDescription := "inclusive"
	if !i.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("lower bound = %f (%s)", i.LowerBound(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `LowerBound` method is inclusive or exclusive.
func (i Float64LowerBound) IsInclusive() bool {
	return i.inclusive
}

// LowerBound returns the float64 value that the final value will not be less than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (i Float64LowerBound) LowerBound() float64 {
	return i.value
}

func (i Float64LowerBound) unimplementable() {}

// NewFloat64LowerBound returns the Float64LowerBound unknown value refinement that indicates the final value will not be less than the specified
// float64 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Float64.
func NewFloat64LowerBound(value float64, inclusive bool) Refinement {
	return Float64LowerBound{
		value:     value,
		inclusive: inclusive,
	}
}
