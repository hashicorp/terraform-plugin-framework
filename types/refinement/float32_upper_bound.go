// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// Float32UpperBound represents an unknown value refinement that indicates the final value will not be greater than the specified
// float32 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Float32.
type Float32UpperBound struct {
	inclusive bool
	value     float32
}

func (i Float32UpperBound) Equal(other Refinement) bool {
	otherVal, ok := other.(Float32UpperBound)
	if !ok {
		return false
	}

	return i.IsInclusive() == otherVal.IsInclusive() && i.UpperBound() == otherVal.UpperBound()
}

func (i Float32UpperBound) String() string {
	rangeDescription := "inclusive"
	if !i.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("upper bound = %f (%s)", i.UpperBound(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `UpperBound` method is inclusive or exclusive.
func (i Float32UpperBound) IsInclusive() bool {
	return i.inclusive
}

// UpperBound returns the float32 value that the final value will not be greater than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (i Float32UpperBound) UpperBound() float32 {
	return i.value
}

func (i Float32UpperBound) unimplementable() {}

// NewFloat32UpperBound returns the Float32UpperBound unknown value refinement that indicates the final value will not be greater than the specified
// float32 value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Float32.
func NewFloat32UpperBound(value float32, inclusive bool) Refinement {
	return Float32UpperBound{
		value:     value,
		inclusive: inclusive,
	}
}
