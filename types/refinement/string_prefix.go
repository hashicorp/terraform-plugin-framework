package refinement

type StringPrefix struct {
	value string
}

func (s StringPrefix) Equal(Refinement) bool {
	return false
}

func (s StringPrefix) String() string {
	return "todo - stringPrefix"
}

func (s StringPrefix) PrefixValue() string {
	return s.value
}

func (s StringPrefix) unimplementable() {}

func NewStringPrefix(value string) Refinement {
	return StringPrefix{
		value: value,
	}
}
