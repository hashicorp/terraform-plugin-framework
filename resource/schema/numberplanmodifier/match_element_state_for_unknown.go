package numberplanmodifier

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// MatchElementStateForUnknown returns a plan modifier that copies a known prior
// state value into the planned value based on list or set element identifying
// paths. Use this when it is known that an unconfigured value under a list or
// set nested attribute or block will remain the same after a resource update.
//
// Identifying path expression(s) should be in the form:
//
//	path.MatchRelative().AtParent().AtName("another_configurable_attribute")
//
// To prevent Terraform errors, the framework automatically sets unconfigured
// and Computed attributes to an unknown value "(known after apply)" on update.
// Using this plan modifier will instead display the prior state value from the
// matching element in the plan, unless a prior plan modifier set the value to
// null or a known value.
//
// To prevent errant implementations, this plan modifier will raise an error
// diagnostic if:
//
//   - Implemented on an attribute which is not beneath a list or set. Use the
//     UseStateForUnknown plan modifier instead.
//   - Zero path expressions are given.
//   - A given path expression is not relative.
//   - A given path expression is not a parent step, then a name step.
//   - A given path expression self-references the attribute where the plan
//     modifier is implemented.
//   - A given path expression references a path to a Computed-only attribute.
//     Only configurable attributes are valid for this plan modifier, otherwise
//     the value at the path will always be unknown and plan modifier will never
//     return an expected value.
func MatchElementStateForUnknown(expressions ...path.Expression) planmodifier.Number {
	return fwplanmodifier.MatchElementStateForUnknownModifier{
		Expressions: expressions,
	}
}
