// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import (
	"fmt"
	"math/big"
)

// NumberUpperBound represents an unknown value refinement that indicates the final value will not be greater than the specified
// *big.Float value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Number.
type NumberUpperBound struct {
	inclusive bool
	value     *big.Float
}

func (n NumberUpperBound) Equal(other Refinement) bool {
	otherVal, ok := other.(NumberUpperBound)
	if !ok {
		return false
	}

	return n.IsInclusive() == otherVal.IsInclusive() && n.UpperBound().Cmp(otherVal.UpperBound()) == 0
}

func (n NumberUpperBound) String() string {
	rangeDescription := "inclusive"
	if !n.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("upper bound = %s (%s)", n.UpperBound().String(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `UpperBound` method is inclusive or exclusive.
func (n NumberUpperBound) IsInclusive() bool {
	return n.inclusive
}

// UpperBound returns the *big.Float value that the final value will not be greater than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (n NumberUpperBound) UpperBound() *big.Float {
	return n.value
}

func (n NumberUpperBound) unimplementable() {}

// NewNumberUpperBound returns the NumberUpperBound unknown value refinement that indicates the final value will not be greater than the specified
// *big.Float value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Number.
func NewNumberUpperBound(value *big.Float, inclusive bool) Refinement {
	return NumberUpperBound{
		value:     value,
		inclusive: inclusive,
	}
}
