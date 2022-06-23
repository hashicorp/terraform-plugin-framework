package path

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// Path represents an attribute path with exact steps. Only exact path
// transversals are supported with this implementation as it must remain
// compatible with all protocol implementations.
type Path struct {
	// steps is the transversals included with the path. In general, operations
	// against the path should protect against modification of the original.
	steps PathSteps
}

// AtListIndex returns a copied path with a new list index step at the end.
// The returned path is safe to modify without affecting the original.
func (p Path) AtListIndex(index int) Path {
	copiedPath := p.Copy()

	copiedPath.steps.Append(PathStepElementKeyInt(index))

	return copiedPath
}

// AtMapKey returns a copied path with a new map key step at the end.
// The returned path is safe to modify without affecting the original.
func (p Path) AtMapKey(key string) Path {
	copiedPath := p.Copy()

	copiedPath.steps.Append(PathStepElementKeyString(key))

	return copiedPath
}

// AtName returns a copied path with a new attribute or block name step at the
// end. The returned path is safe to modify without affecting the original.
func (p Path) AtName(name string) Path {
	copiedPath := p.Copy()

	copiedPath.steps.Append(PathStepAttributeName(name))

	return copiedPath
}

// AtSetValue returns a copied path with a new set value step at the end.
// The returned path is safe to modify without affecting the original.
func (p Path) AtSetValue(value attr.Value) Path {
	copiedPath := p.Copy()

	copiedPath.steps.Append(PathStepElementKeyValue{Value: value})

	return copiedPath
}

// Copy returns a duplicate of the path that is safe to modify without
// affecting the original.
func (p Path) Copy() Path {
	return Path{
		steps: p.Steps(),
	}
}

// Equal returns true if the given path is exactly equivalent.
func (p Path) Equal(o Path) bool {
	if p.steps == nil && o.steps == nil {
		return true
	}

	if p.steps == nil {
		return false
	}

	if !p.steps.Equal(o.steps) {
		return false
	}

	return true
}

// Expression returns an Expression which exactly matches the Path.
func (p Path) Expression() Expression {
	return Expression{
		steps: p.steps.ExpressionSteps(),
	}
}

// ParentPath returns a copy of the path with the last step removed.
//
// If the current path is empty, an empty path is returned.
func (p Path) ParentPath() Path {
	if len(p.steps) == 0 {
		return Empty()
	}

	_, remainingSteps := p.steps.Copy().LastStep()

	return Path{
		steps: remainingSteps,
	}
}

// Steps returns a copy of the underlying path steps. Returns an empty
// collection of steps if path is nil.
func (p Path) Steps() PathSteps {
	if len(p.steps) == 0 {
		return PathSteps{}
	}

	return p.steps.Copy()
}

// String returns the human-readable representation of the path.
// It is intended for logging and error messages and is not protected by
// compatibility guarantees.
func (p Path) String() string {
	return p.steps.String()
}

// Empty creates an empty attribute path. Provider code should use Root.
func Empty() Path {
	return Path{
		steps: PathSteps{},
	}
}

// Root creates an attribute path starting with a PathStepAttributeName.
func Root(rootAttributeName string) Path {
	return Path{
		steps: PathSteps{
			PathStepAttributeName(rootAttributeName),
		},
	}
}
