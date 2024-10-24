package refinement

type NotNull struct{}

func (n NotNull) Equal(Refinement) bool {
	return false
}

func (n NotNull) String() string {
	return "todo - NotNull"
}

func (n NotNull) unimplementable() {}

// TODO: Should this accept a value? If a value is unknown and the it's refined to be null
// then the value should be a known value of null instead.
func NewNotNull() Refinement {
	return NotNull{}
}
