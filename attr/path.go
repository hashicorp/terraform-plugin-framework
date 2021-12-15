package attr

// Path is used to identify part of a tfsdk.Schema, attr.Type, or attr.Value.
// It is a series of steps, describing a traversal of a Terraform type. Each
// Terraform type has its own limits on what steps can be used to traverse it.
type Path struct {
	steps []step
}

// New returns an Path that is ready to be used.
func NewPath() Path {
	return Path{}
}

func copySteps(in []step) []step {
	out := make([]step, len(in))
	copy(out, in)
	return out
}

// IsEmpty returns true if a Path has no steps and effectively points to the
// root of the value.
func (a Path) IsEmpty() bool {
	return len(a.steps) < 1
}

// Parent returns a Path pointing to the parent of the attribute or element
// that `a` points to. If `a` has no parent, an empty Path is returned.
func (a Path) Parent() Path {
	if len(a.steps) < 1 {
		return Path{steps: []step{}}
	}
	return Path{
		steps: copySteps(a.steps)[:len(a.steps)-1],
	}
}

// Attribute returns a copy of `a`, with another step added to select the named
// attribute of the object that `a` points to.
func (a Path) Attribute(name string) Path {
	return Path{
		steps: append(copySteps(a.steps), attributeName(name)),
	}
}

// ElementKey returns a copy of `a` with another step added to select the named
// key of the map that `a` points to.
func (a Path) ElementKey(name string) Path {
	return Path{
		steps: append(copySteps(a.steps), elementKeyString(name)),
	}
}

// ElementPos returns a copy of `a` with another step added to select the
// element that is in the specified position of the list or tuple that `a`
// points to.
func (a Path) ElementPos(pos int) Path {
	return Path{
		steps: append(copySteps(a.steps), elementKeyInt(pos)),
	}
}

// Element returns a copy of `a` with another step added to select the
// specified element of the set that `a` points to.
func (a Path) Element(val Value) Path {
	return Path{
		steps: append(copySteps(a.steps), elementKeyValue{Value: val}),
	}
}

type step interface {
	unimplementablePathStep()
}

// attributeName is a step that selects a single attribute, by its name. It is
// used on attr.TypeWithAttributeTypes, tfsdk.NestedAttributes, and
// tfsdk.Schema types.
type attributeName string

func (a attributeName) unimplementablePathStep() {}

// elementKeyString is a step that selects a single element using a string
// index. It is used on attr.TypeWithElementType types that return a
// tftypes.Map from their TerraformType method. It is also used on
// tfsdk.MapNestedAttribute nested attributes.
type elementKeyString string

func (e elementKeyString) unimplementablePathStep() {}

// elementKeyInt is a step that selects a single element using an integer
// index. It is used on attr.TypeWithElementTypes types and
// attr.TypeWithElementType types that return a tftypes.List from their
// TerraformType method. It is also used on tfsdk.ListNestedAttribute nested
// attributes.
type elementKeyInt int

func (e elementKeyInt) unimplementablePathStep() {}

// elementKeyValue is a step that selects a single element using an attr.Value
// index. It is used on attr.TypeWithElementType types that return a
// tftypes.Set from their TerraformType method. It is also used on
// tfsdk.SetNestedAttribute nested attributes.
type elementKeyValue struct {
	Value Value
}

func (e elementKeyValue) unimplementablePathStep() {}
