package refinement

import "math/big"

type NumberLowerBound struct {
	inclusive bool
	value     *big.Float
}

func (s NumberLowerBound) Equal(Refinement) bool {
	return false
}

func (s NumberLowerBound) String() string {
	return "todo - NumberLowerBound"
}

func (s NumberLowerBound) IsInclusive() bool {
	return s.inclusive
}

func (s NumberLowerBound) LowerBound() *big.Float {
	return s.value
}

func (s NumberLowerBound) unimplementable() {}

func NewNumberLowerBound(value *big.Float, inclusive bool) Refinement {
	return NumberLowerBound{
		value:     value,
		inclusive: inclusive,
	}
}
