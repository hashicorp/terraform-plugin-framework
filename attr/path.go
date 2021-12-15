package attr

import (
	"strconv"
	"strings"
)

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

// ChildOf returns true if `a` can be considered a child of `p`. `a` can be
// considered a child of `p` if `a` contains all the same steps `p` has, in the
// same order, though `a` can contain additional steps after all of the steps
// it shares with `p`.
func (a Path) ChildOf(p Path) bool {
	if len(a.steps) <= len(p.steps) {
		return false
	}
	for pos, pStep := range p.steps {
		aStep := a.steps[pos]
		if !pStep.Equal(aStep) {
			return false
		}
	}
	return true
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

// String returns a human-friendly string representation of the Path. There are
// no compatibility guarantees about its formatting, it is not considered part
// of the API contract and may change without notice. It is meant to be used in
// logging and other debugging situations, and is not meant to be parsed.
func (a Path) String() string {
	var res strings.Builder
	for pos, step := range a.steps {
		if pos != 0 {
			res.WriteString(".")
		}
		res.WriteString(step.String())
	}
	return res.String()
}

type step interface {
	unimplementablePathStep()
	Equal(step) bool
	String() string
}

// attributeName is a step that selects a single attribute, by its name. It is
// used on attr.TypeWithAttributeTypes, tfsdk.NestedAttributes, and
// tfsdk.Schema types.
type attributeName string

func (a attributeName) unimplementablePathStep() {}

func (a attributeName) Equal(o step) bool {
	other, ok := o.(attributeName)
	if !ok {
		return false
	}
	return a == other
}

func (a attributeName) String() string {
	return "AttributeName(\"" + string(a) + "\")"
}

// elementKeyString is a step that selects a single element using a string
// index. It is used on attr.TypeWithElementType types that return a
// tftypes.Map from their TerraformType method. It is also used on
// tfsdk.MapNestedAttribute nested attributes.
type elementKeyString string

func (e elementKeyString) unimplementablePathStep() {}

func (e elementKeyString) Equal(o step) bool {
	other, ok := o.(elementKeyString)
	if !ok {
		return false
	}
	return e == other
}

func (e elementKeyString) String() string {
	return "ElementKeyString(\"" + string(e) + "\")"
}

// elementKeyInt is a step that selects a single element using an integer
// index. It is used on attr.TypeWithElementTypes types and
// attr.TypeWithElementType types that return a tftypes.List from their
// TerraformType method. It is also used on tfsdk.ListNestedAttribute nested
// attributes.
type elementKeyInt int

func (e elementKeyInt) unimplementablePathStep() {}

func (e elementKeyInt) Equal(o step) bool {
	other, ok := o.(elementKeyInt)
	if !ok {
		return false
	}
	return e == other
}

func (e elementKeyInt) String() string {
	return "ElementKeyInt(" + strconv.FormatInt(int64(e), 10) + ")"
}

// elementKeyValue is a step that selects a single element using an attr.Value
// index. It is used on attr.TypeWithElementType types that return a
// tftypes.Set from their TerraformType method. It is also used on
// tfsdk.SetNestedAttribute nested attributes.
type elementKeyValue struct {
	Value Value
}

func (e elementKeyValue) unimplementablePathStep() {}

func (e elementKeyValue) Equal(o step) bool {
	other, ok := o.(elementKeyValue)
	if !ok {
		return false
	}
	return e.Value.Equal(other.Value)
}

func (e elementKeyValue) String() string {
	return "ElementKeyValue(" + e.Value.String() + ")"
}
