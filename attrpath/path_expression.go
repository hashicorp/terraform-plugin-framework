package attrpath

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

// Expression is like a Path, but allows for using wildcards to apply to
// any attribute or element. It's useful for specifying validation rules and
// other scenarios where Path is too limited.
//
// In theory, a Expression is a superset of a Path, in that a
// Expression doesn't necessarily have to use wildcards, and thus could
// identify a specific attribute or element, just like a Path. But
// Expressions can also specify a class of attributes or elements, i.e.
// "this attribute for every object in this list".
type Expression struct {
	steps []expressionStep
}

// NewExpression returns an Expression that is ready to be used.
func NewExpression() Expression {
	return Expression{}
}

// IsEmpty returns true if a Expression has no steps and effectively points
// to the root of the value.
func (a Expression) IsEmpty() bool {
	return len(a.steps) < 1
}

// Matches returns true if the passed Path matches the Expression.
func (a Expression) Matches(ctx context.Context, p Path) bool {
	ctx = tfsdklog.With(ctx, "path", p)
	ctx = tfsdklog.With(ctx, "path_expression", a)
	if len(a.steps) != len(p.steps) {
		tfsdklog.Trace(ctx, "path not considered a match, different number of steps")
		return false
	}
	for pos, exp := range a.steps {
		step := p.steps[pos]
		ctx := tfsdklog.With(ctx, "path_step", step)
		ctx = tfsdklog.With(ctx, "path_expression_step", exp)
		switch expStep := exp.(type) {
		case attributeNameExpression:
			attrStep, ok := step.(attributeName)
			if !ok {
				tfsdklog.Trace(ctx, "path not considered a match, expression is for an attribute but path is not")
				return false
			}
			if expStep.any {
				tfsdklog.Trace(ctx, "path matches so far, expression allows for any attribute name")
				continue
			}
			if expStep.exact != nil && *expStep.exact != string(attrStep) {
				tfsdklog.Trace(ctx, "path not considered a match, expression is for another attribute")
				return false
			}
			for _, exclude := range expStep.except {
				if exclude == string(attrStep) {
					tfsdklog.Trace(ctx, "path not considered a match, uses excluded attribute")
					return false
				}
			}
			tfsdklog.Trace(ctx, "path matches so far, expression doesn't exclude attribute")
			continue
		case elementKeyStringExpression:
			eksStep, ok := step.(elementKeyString)
			if !ok {
				tfsdklog.Trace(ctx, "path not considered a match, expression is for an element key but path is not")
				return false
			}
			if expStep.any {
				tfsdklog.Trace(ctx, "path matches so far, expression allows for any element key")
				continue
			}
			if expStep.exact != nil && *expStep.exact != string(eksStep) {
				tfsdklog.Trace(ctx, "path not considered a match, expression is for another element key")
				return false
			}
			for _, exclude := range expStep.except {
				if exclude == string(eksStep) {
					tfsdklog.Trace(ctx, "path not considered a match, uses excluded element key")
					return false
				}
			}
			tfsdklog.Trace(ctx, "path matches so far, expression doesn't exclude element key")
			continue
		case elementKeyIntExpression:
			ekiStep, ok := step.(elementKeyInt)
			if !ok {
				tfsdklog.Trace(ctx, "path not considered a match, expression is for an element position but path is not")
				return false
			}
			if expStep.any {
				tfsdklog.Trace(ctx, "path matches so far, expression allows for any element position")
				continue
			}
			if expStep.exact != nil && *expStep.exact != int(ekiStep) {
				tfsdklog.Trace(ctx, "path not considered a match, expression is for another element position")
				return false
			}
			for _, exclude := range expStep.except {
				if exclude == int(ekiStep) {
					tfsdklog.Trace(ctx, "path not considered a match, uses excluded element position")
					return false
				}
			}
			tfsdklog.Trace(ctx, "path matches so far, expression doesn't exclude element position")
			continue
			/*
				case elementKeyValueExpression:
					ekvStep, ok := step.(elementKeyValue)
					if !ok {
						tfsdklog.Trace(ctx, "path not considered a match, expression is for an element value but path is not")
						return false
					}
					if expStep.any {
						tfsdklog.Trace(ctx, "path matches so far, expression allows for any element value")
						continue
					}
					if expStep.exact != nil && !expStep.exact.Equal(ekvStep.Value) {
						tfsdklog.Trace(ctx, "path not considered a match, expression is for another element")
						return false
					}
					for _, exclude := range expStep.except {
						if exclude.Equal(ekvStep.Value) {
							tfsdklog.Trace(ctx, "path not considered a match, uses excluded element value")
							return false
						}
					}
					continue
			*/
		default:
			tfsdklog.Error(ctx, "unknown path expression step type")
		}
	}
	return true
}

// Parent returns a Expression pointing to the parent of the attribute or
// element that `a` points to. If `a` has no parent, an empty Expression is
// returned.
func (a Expression) Parent() Expression {
	if len(a.steps) < 1 {
		return Expression{steps: []expressionStep{}}
	}
	return Expression{
		steps: copyExpressionSteps(a.steps)[:len(a.steps)-1],
	}
}

// Attribute returns a copy of `a`, with another step added to select the named
// attribute of the object that `a` points to.
func (a Expression) Attribute(name string) Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), attributeNameExpression{exact: &name}),
	}
}

// AnyAttribute returns a copy of `a`, with another step added that matches any
// attribute of the object that `a` points to.
func (a Expression) AnyAttribute() Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), attributeNameExpression{any: true}),
	}
}

// AnyAttributeExcept returns a copy of `a`, with another step added that
// matches any attribute of the object that `a` points to except the named
// attributes.
func (a Expression) AnyAttributeExcept(name ...string) Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), attributeNameExpression{except: name}),
	}
}

// ElementKey returns a copy of `a` with another step added to select the named
// key of the map that `a` points to.
func (a Expression) ElementKey(name string) Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), elementKeyStringExpression{exact: &name}),
	}
}

// Any ElementKey returns a copy of `a` with another step added that matches
// any key of the map that `a` points to.
func (a Expression) AnyElementKey() Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), elementKeyStringExpression{any: true}),
	}
}

// Any ElementKey returns a copy of `a` with another step added that matches
// any key of the map that `a` points to except for the named keys.
func (a Expression) AnyElementKeyExcept(key ...string) Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), elementKeyStringExpression{except: key}),
	}
}

// ElementPos returns a copy of `a` with another step added to select the
// element that is in the specified position of the list or tuple that `a`
// points to.
func (a Expression) ElementPos(pos int) Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), elementKeyIntExpression{exact: &pos}),
	}
}

// AnyElementPos returns a copy of `a` with another step added that matches any
// element of the list or tuple that `a` points to.
func (a Expression) AnyElementPos() Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), elementKeyIntExpression{any: true}),
	}
}

// AnyElementPosExcept returns a copy of `a` with another step added that
// matches any element of the list or tuple that `a` points to except those in
// the specified positions.
func (a Expression) AnyElementPosExcept(pos ...int) Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), elementKeyIntExpression{except: pos}),
	}
}

// Element returns a copy of `a` with another step added to select the
// specified element of the set that `a` points to.
/*
func (a Expression) Element(val Value) Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), elementKeyValueExpression{exact: val}),
	}
}

// AnyElement returns a copy of `a` with another step added that matches any
// element of the set that `a` points to.
func (a Expression) AnyElement() Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), elementKeyValueExpression{any: true}),
	}
}

// AnyElementExcept returns a copy of `a` with another step added that matches
// any element of the set that `a` points to except the passed values.
func (a Expression) AnyElementExcept(val ...Value) Expression {
	return Expression{
		steps: append(copyExpressionSteps(a.steps), elementKeyValueExpression{except: val}),
	}
}
*/

type expressionStep interface {
	unimplementableExpressionStep()
}

// attributeNameExpression is a step that selects one or more attributes by
// name. It is used on attr.TypeWithAttributeTypes, tfsdk.NestedAttributes, and
// tfsdk.Schema types.
type attributeNameExpression struct {
	any    bool
	exact  *string
	except []string
}

func (a attributeNameExpression) unimplementableExpressionStep() {}

// elementKeyStringExpression is a step that selects one or more elements using
// a string index. It is used on attr.TypeWithElementType types that return a
// tftypes.Map from their TerraformType method. It is also used on
// tfsdk.MapNestedAttribute nested attributes.
type elementKeyStringExpression struct {
	any    bool
	exact  *string
	except []string
}

func (e elementKeyStringExpression) unimplementableExpressionStep() {}

// elementKeyIntExpression is a step that selects one or more elements using an
// integer index. It is used on attr.TypeWithElementTypes types and
// attr.TypeWithElementType types that return a tftypes.List from their
// TerraformType method. It is also used on tfsdk.ListNestedAttribute nested
// attributes.
type elementKeyIntExpression struct {
	any    bool
	exact  *int
	except []int
}

func (e elementKeyIntExpression) unimplementableExpressionStep() {}

// elementKeyValueExpression is a step that selects one ore more elements using
// an attr.Value index. It is used on attr.TypeWithElementType types that
// return a tftypes.Set from their TerraformType method. It is also used on
// tfsdk.SetNestedAttribute nested attributes.
/*
type elementKeyValueExpression struct {
	any    bool
	exact  Value
	except []Value
}

func (e elementKeyValueExpression) unimplementableExpressionStep() {}
*/

func copyExpressionSteps(in []expressionStep) []expressionStep {
	out := make([]expressionStep, len(in))
	copy(out, in)
	return out
}
