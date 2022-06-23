package path

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// Expression represents an attribute path with expression steps, which can
// represent zero, one, or more actual Paths.
type Expression struct {
	// steps is the transversals included with the expression. In general,
	// operations against the path should protect against modification of the
	// original.
	steps ExpressionSteps
}

// AtAnyListIndex returns a copied expression with a new list index step at the
// end. The returned path is safe to modify without affecting the original.
func (e Expression) AtAnyListIndex() Expression {
	copiedPath := e.Copy()

	copiedPath.steps.Append(ExpressionStepElementKeyIntAny{})

	return copiedPath
}

// AtAnyMapKey returns a copied expression with a new map key step at the end.
// The returned path is safe to modify without affecting the original.
func (e Expression) AtAnyMapKey() Expression {
	copiedPath := e.Copy()

	copiedPath.steps.Append(ExpressionStepElementKeyStringAny{})

	return copiedPath
}

// AtAnySetValue returns a copied expression with a new set value step at the
// end. The returned path is safe to modify without affecting the original.
func (e Expression) AtAnySetValue() Expression {
	copiedPath := e.Copy()

	copiedPath.steps.Append(ExpressionStepElementKeyValueAny{})

	return copiedPath
}

// AtListIndex returns a copied expression with a new list index step at the
// end. The returned path is safe to modify without affecting the original.
func (e Expression) AtListIndex(index int) Expression {
	copiedPath := e.Copy()

	copiedPath.steps.Append(ExpressionStepElementKeyIntExact(index))

	return copiedPath
}

// AtMapKey returns a copied expression with a new map key step at the end.
// The returned path is safe to modify without affecting the original.
func (e Expression) AtMapKey(key string) Expression {
	copiedPath := e.Copy()

	copiedPath.steps.Append(ExpressionStepElementKeyStringExact(key))

	return copiedPath
}

// AtName returns a copied expression with a new attribute or block name step
// at the end. The returned path is safe to modify without affecting the
// original.
func (e Expression) AtName(name string) Expression {
	copiedPath := e.Copy()

	copiedPath.steps.Append(ExpressionStepAttributeNameExact(name))

	return copiedPath
}

// AtParent returns a copied expression with a new parent step at the end.
// The returned path is safe to modify without affecting the original.
func (e Expression) AtParent() Expression {
	copiedPath := e.Copy()

	copiedPath.steps.Append(ExpressionStepParent{})

	return copiedPath
}

// AtSetValue returns a copied expression with a new set value step at the end.
// The returned path is safe to modify without affecting the original.
func (e Expression) AtSetValue(value attr.Value) Expression {
	copiedPath := e.Copy()

	copiedPath.steps.Append(ExpressionStepElementKeyValueExact{Value: value})

	return copiedPath
}

// Copy returns a duplicate of the expression that is safe to modify without
// affecting the original.
func (e Expression) Copy() Expression {
	return Expression{
		steps: e.Steps(),
	}
}

// Equal returns true if the given expression is exactly equivalent.
func (e Expression) Equal(o Expression) bool {
	if e.steps == nil && o.steps == nil {
		return true
	}

	if e.steps == nil {
		return false
	}

	if !e.steps.Equal(o.steps) {
		return false
	}

	return true
}

// Matches returns true if the given Path is valid for the Expression.
func (e Expression) Matches(path Path) bool {
	return e.steps.Matches(path.Steps())
}

// Steps returns a copy of the underlying expression steps. Returns an empty
// collection of steps if expression is nil.
func (e Expression) Steps() ExpressionSteps {
	if len(e.steps) == 0 {
		return ExpressionSteps{}
	}

	return e.steps.Copy()
}

// String returns the human-readable representation of the path.
// It is intended for logging and error messages and is not protected by
// compatibility guarantees.
func (e Expression) String() string {
	return e.steps.String()
}

// MatchParent creates an attribute path expression starting with
// ExpressionStepParent. This allows creating a relative expression in
// nested schemas.
func MatchParent() Expression {
	return Expression{
		steps: ExpressionSteps{
			ExpressionStepParent{},
		},
	}
}

// MatchRoot creates an attribute path expression starting with
// ExpressionStepAttributeNameExact.
func MatchRoot(rootAttributeName string) Expression {
	return Expression{
		steps: ExpressionSteps{
			ExpressionStepAttributeNameExact(rootAttributeName),
		},
	}
}
