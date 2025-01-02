// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// Float32LowerBound represents an unknown value refinement that indicates the final value will not be less than the specified
// float32 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Float32.
type Float32LowerBound struct {
	inclusive bool
	value     float32
}

func (i Float32LowerBound) Equal(other Refinement) bool {
	otherVal, ok := other.(Float32LowerBound)
	if !ok {
		return false
	}

	return i.IsInclusive() == otherVal.IsInclusive() && i.LowerBound() == otherVal.LowerBound()
}

func (i Float32LowerBound) String() string {
	rangeDescription := "inclusive"
	if !i.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("lower bound = %f (%s)", i.LowerBound(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `LowerBound` method is inclusive or exclusive.
func (i Float32LowerBound) IsInclusive() bool {
	return i.inclusive
}

// LowerBound returns the float32 value that the final value will not be less than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (i Float32LowerBound) LowerBound() float32 {
	return i.value
}

func (i Float32LowerBound) unimplementable() {}

// NewFloat32LowerBound returns the Float32LowerBound unknown value refinement that indicates the final value will not be less than the specified
// float32 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Float32.
func NewFloat32LowerBound(value float32, inclusive bool) Refinement {
	return Float32LowerBound{
		value:     value,
		inclusive: inclusive,
	}
}
