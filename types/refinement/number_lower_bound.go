// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import (
	"fmt"
	"math/big"
)

// NumberLowerBound represents an unknown value refinement that indicates the final value will not be less than the specified
// *big.Float value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Number.
type NumberLowerBound struct {
	inclusive bool
	value     *big.Float
}

func (n NumberLowerBound) Equal(other Refinement) bool {
	otherVal, ok := other.(NumberLowerBound)
	if !ok {
		return false
	}

	return n.IsInclusive() == otherVal.IsInclusive() && n.LowerBound().Cmp(otherVal.LowerBound()) == 0
}

func (n NumberLowerBound) String() string {
	rangeDescription := "inclusive"
	if !n.IsInclusive() {
		rangeDescription = "exclusive"
	}

	return fmt.Sprintf("lower bound = %s (%s)", n.LowerBound().String(), rangeDescription)
}

// IsInclusive returns whether the bound returned by the `LowerBound` method is inclusive or exclusive.
func (n NumberLowerBound) IsInclusive() bool {
	return n.inclusive
}

// LowerBound returns the *big.Float value that the final value will not be less than. The `IsInclusive` method must also be used during
// comparison to determine whether the bound is inclusive or exclusive.
func (n NumberLowerBound) LowerBound() *big.Float {
	return n.value
}

func (n NumberLowerBound) unimplementable() {}

// NewNumberLowerBound returns the NumberLowerBound unknown value refinement that indicates the final value will not be less than the specified
// *big.Float value, as well as whether that bound is inclusive or exclusive. This refinement can only be applied to types.Number.
func NewNumberLowerBound(value *big.Float, inclusive bool) Refinement {
	return NumberLowerBound{
		value:     value,
		inclusive: inclusive,
	}
}
