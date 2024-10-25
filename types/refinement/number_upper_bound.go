package refinement

import "math/big"

type NumberUpperBound struct {
	inclusive bool
	value     *big.Float
}

func (s NumberUpperBound) Equal(Refinement) bool {
	return false
}

func (s NumberUpperBound) String() string {
	return "todo - NumberUpperBound"
}

func (s NumberUpperBound) IsInclusive() bool {
	return s.inclusive
}

func (s NumberUpperBound) UpperBound() *big.Float {
	return s.value
}

func (s NumberUpperBound) unimplementable() {}

func NewNumberUpperBound(value *big.Float, inclusive bool) Refinement {
	return NumberUpperBound{
		value:     value,
		inclusive: inclusive,
	}
}
