package refinement

type Int64LowerBound struct {
	inclusive bool
	value     int64
}

func (s Int64LowerBound) Equal(Refinement) bool {
	return false
}

func (s Int64LowerBound) String() string {
	return "todo - Int64LowerBound"
}

func (s Int64LowerBound) IsInclusive() bool {
	return s.inclusive
}

func (s Int64LowerBound) LowerBound() int64 {
	return s.value
}

func (s Int64LowerBound) unimplementable() {}

func NewInt64LowerBound(value int64, inclusive bool) Refinement {
	return Int64LowerBound{
		value:     value,
		inclusive: inclusive,
	}
}
