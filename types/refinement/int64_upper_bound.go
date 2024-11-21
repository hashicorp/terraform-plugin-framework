package refinement

type Int64UpperBound struct {
	inclusive bool
	value     int64
}

func (s Int64UpperBound) Equal(Refinement) bool {
	return false
}

func (s Int64UpperBound) String() string {
	return "todo - Int64UpperBound"
}

func (s Int64UpperBound) IsInclusive() bool {
	return s.inclusive
}

func (s Int64UpperBound) UpperBound() int64 {
	return s.value
}

func (s Int64UpperBound) unimplementable() {}

func NewInt64UpperBound(value int64, inclusive bool) Refinement {
	return Int64UpperBound{
		value:     value,
		inclusive: inclusive,
	}
}
