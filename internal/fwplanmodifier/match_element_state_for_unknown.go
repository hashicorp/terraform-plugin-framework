package fwplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/parentpath"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ planmodifier.Bool    = MatchElementStateForUnknownModifier{}
	_ planmodifier.Float64 = MatchElementStateForUnknownModifier{}
	_ planmodifier.Int64   = MatchElementStateForUnknownModifier{}
	_ planmodifier.List    = MatchElementStateForUnknownModifier{}
	_ planmodifier.Map     = MatchElementStateForUnknownModifier{}
	_ planmodifier.Number  = MatchElementStateForUnknownModifier{}
	_ planmodifier.Object  = MatchElementStateForUnknownModifier{}
	_ planmodifier.Set     = MatchElementStateForUnknownModifier{}
	_ planmodifier.String  = MatchElementStateForUnknownModifier{}
)

// MatchElementStateForUnknownModifier implements the plan modifier.
type MatchElementStateForUnknownModifier struct {
	Expressions path.Expressions
}

// Description returns a human-readable description of the plan modifier.
func (m MatchElementStateForUnknownModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m MatchElementStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// PlanModify implements the shared plan modification logic. This logic is
// currently written with the expectation that a valid setup of this plan
// modifier is when the request path is:
//
//   - Zero or more preceding path steps
//   - AttributeName path step representing a list/set nested attribute/block
//   - ElementKeyInt/ElementKeyValue path step of nested object elements
//   - AttributeName path step representing a computed attribute on which the
//     plan modifier is implemented.
//
// The given expressions are similarly expected to represent another attribute
// within the same nested object as where the plan modifier is implemented. This
// introduces the following validation rules for expressions:
//
//   - They must resolve to the same number of path steps as the request path.
//     For relative expressions, this generally means
//     path.MatchRelative().AtParent().AtName() so it is an another attribute
//     within the same nested object as where the plan modifier is implemented.
//   - They must be relative to properly match the current element's identifying
//     values. A root based path would never be possible beyond a single element
//     list/set because it would either need to be hardcoded to the first
//     element or introduce an "any" element path step (such as AtAnyListIndex()
//     or AtAnySetValue()), which could return multiple values.
//   - They must not resolve above or outside the list/set nested
//     attribute/block as this introduces another order of complexity.
//
// The logic is written with validation based on these expectations. The inputs
// are generic path expressions, so if the logic can safely handle other
// assumptions, theoretically developer-facing changes are not required for that
// type of enhancement. Otherwise, if breaking changes would be necessary,
// another plan modifier could be introduced to handle more complex, but
// common and generically verifiable use cases.
//
// Beyond the validation logic, the general algorithm of this is:
//
//   - For each expression, fetch the plan data using the expression. This
//     validates that the expression is actually valid for the schema beyond the
//     initial validation and will yield the identifying values that later will
//     be checked against all elements in the prior state for element alignment.
//   - If any fetched plan data is unknown, return early.
//   - If the request state value is null, return early.
//   - If the request plan value is known already, return early.
//   - If the request configuration value is unknown, return early to prevent
//     unexpected behavior.
//   - For all list/set elements under the parent of the request path, check all
//     identifying values. If there is an element where they all match, set the
//     response plan value to the prior state value from the matching element.
func (m MatchElementStateForUnknownModifier) PlanModify(ctx context.Context, req MatchElementStateForUnknownRequest, resp *MatchElementStateForUnknownResponse) {
	// Verify this plan modifier is only being used beneath a list or set.
	if !parentpath.HasListOrSet(req.Path) {
		resp.Diagnostics.Append(MatchElementStateForUnknownOutsideListOrSetDiag(req.Path))

		return
	}

	// Verify at least one expression was given.
	if len(m.Expressions) == 0 {
		resp.Diagnostics.Append(MatchElementStateForUnknownMissingExpressionsDiag(req.Path))

		return
	}

	// Collect any initial expression validation issues.
	for _, expression := range m.Expressions {
		if expression.IsRoot() {
			resp.Diagnostics.Append(MatchElementStateForUnknownRootExpressionDiag(req.Path, expression))

			continue
		}

		expressionSteps := expression.Steps()

		if len(expressionSteps) != 2 {
			resp.Diagnostics.Append(MatchElementStateForUnknownInvalidExpressionDiag(req.Path, expression))

			continue
		}

		if _, ok := expressionSteps[0].(path.ExpressionStepParent); !ok {
			resp.Diagnostics.Append(MatchElementStateForUnknownInvalidExpressionDiag(req.Path, expression))

			continue
		}

		if _, ok := expressionSteps[1].(path.ExpressionStepAttributeNameExact); !ok {
			resp.Diagnostics.Append(MatchElementStateForUnknownInvalidExpressionDiag(req.Path, expression))

			continue
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Merge the request path expression with each given expression and begin
	// the process of saving any identifying values into a mapping of attribute
	// name to value.
	identifyingValues := make(map[string]attr.Value, len(m.Expressions))

	for _, expression := range req.PathExpression.MergeExpressions(m.Expressions...) {
		// Verify expression is not a self-reference.
		if expression.Matches(req.Path) {
			resp.Diagnostics.Append(MatchElementStateForUnknownInvalidExpressionDiag(req.Path, expression))

			continue
		}

		// Verify expression points to an actual path and is not a Computed-only
		// attribute. Only configurable attributes are valid for this plan
		// modifier, otherwise the value at the path will always be unknown.
		matchedPaths, diags := req.Plan.PathMatches(ctx, expression)

		resp.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		// Verify at least one matched path was returned. Each expression should
		// always have plan data, which is necessary for identifying the element
		// alignment.
		if len(matchedPaths) == 0 {
			resp.Diagnostics.Append(MatchElementStateForUnknownInvalidExpressionDiag(req.Path, expression))

			continue
		}

		// Verify there are not multiple matching paths. If so, the logic would
		// not have a single set of identifying values. This is defensive logic
		// in case prior validation did not already catch the case of an
		// expression step which might cause multiple matches. This cannot be
		// unit tested while proper validation is in place.
		if len(matchedPaths) > 1 {
			resp.Diagnostics.Append(
				FrameworkImplementationErrorDiag(
					req.Path,
					"MatchElementStateForUnknown received multiple matching paths for an expression.\n"+
						fmt.Sprintf("Expression: %s", expression),
				),
			)

			return
		}

		// There should only be one matched path per expression at this point.
		matchedPath := matchedPaths[0]

		var matchedPathValue attr.Value

		diags = req.Plan.GetAttribute(ctx, matchedPath, &matchedPathValue)

		resp.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		// Last step represents the attribute name of the identifying value.
		lastStep, _ := matchedPath.Steps().LastStep()
		attributeNameStep, ok := lastStep.(path.PathStepAttributeName)

		if !ok {
			resp.Diagnostics.Append(
				FrameworkImplementationErrorDiag(
					req.Path,
					"MatchElementStateForUnknown matched path last step was not AttributeName.\n"+
						fmt.Sprintf("Expression: %s", expression),
				),
			)

			return
		}

		identifyingValues[string(attributeNameStep)] = matchedPathValue
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Delay plan modification until all identifying attributes have a
	// known value.
	for _, identifyingValue := range identifyingValues {
		if identifyingValue.IsUnknown() {
			return
		}
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise
	// interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Get the list/set attribute path to fetch the prior state and all its
	// elements.
	collectionPath := req.Path.ParentPath().ParentPath()

	// Prior state should only ever contain null or known values. The elements
	// will always be objects per the prior validation.
	var priorStateObjects []basetypes.ObjectValue

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, collectionPath, &priorStateObjects)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var matchedPriorStateValue attr.Value

	// Loop through all prior state elements to find one which matches all
	// identifying values.
	for _, priorStateObject := range priorStateObjects {
		objectAttributes := priorStateObject.Attributes()

		// Loop through all identifying values to check each against this
		// element. All values must exist and be equal.
		var matching bool

		for attributeName, identifyingValue := range identifyingValues {
			objectAttributeValue, ok := objectAttributes[attributeName]

			// Allow non-existent prior state attributes, rather than raise an
			// error, because the schema may have had new nested object
			// attributes added that are considered identifying, but not yet in
			// the prior state.
			if !ok {
				matching = false

				break
			}

			if !objectAttributeValue.Equal(identifyingValue) {
				matching = false

				break
			}

			matching = true
		}

		if matching {
			// Last step represents the attribute name of the attribute with the
			// plan modifier.
			lastStep, _ := req.Path.Steps().LastStep()
			attributeNameStep, ok := lastStep.(path.PathStepAttributeName)

			if !ok {
				resp.Diagnostics.Append(
					FrameworkImplementationErrorDiag(
						req.Path,
						"MatchElementStateForUnknown request path last step was not AttributeName.",
					),
				)

				return
			}

			objectAttributeValue, ok := objectAttributes[string(attributeNameStep)]

			if ok {
				matchedPriorStateValue = objectAttributeValue
			}

			break
		}
	}

	// Do nothing if there is no prior state value or it is null.
	if matchedPriorStateValue == nil || matchedPriorStateValue.IsNull() {
		return
	}

	resp.PlanValue = matchedPriorStateValue
}

// PlanModifyBool implements the Bool plan modification logic.
func (m MatchElementStateForUnknownModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	genericReq := MatchElementStateForUnknownRequest{
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Plan:           req.Plan,
		PlanValue:      req.PlanValue,
		State:          req.State,
	}
	genericResp := &MatchElementStateForUnknownResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModify(ctx, genericReq, genericResp)

	resp.Diagnostics = genericResp.Diagnostics

	planValue, ok := genericResp.PlanValue.(basetypes.BoolValue)

	if !ok {
		resp.Diagnostics.Append(PlanValueTypeAssertionDiag(req.Path, req.PlanValue, genericResp.PlanValue))
	}

	resp.PlanValue = planValue
}

// PlanModifyFloat64 implements the Float64 plan modification logic.
func (m MatchElementStateForUnknownModifier) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	genericReq := MatchElementStateForUnknownRequest{
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Plan:           req.Plan,
		PlanValue:      req.PlanValue,
		State:          req.State,
	}
	genericResp := &MatchElementStateForUnknownResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModify(ctx, genericReq, genericResp)

	resp.Diagnostics = genericResp.Diagnostics

	planValue, ok := genericResp.PlanValue.(basetypes.Float64Value)

	if !ok {
		resp.Diagnostics.Append(PlanValueTypeAssertionDiag(req.Path, req.PlanValue, genericResp.PlanValue))
	}

	resp.PlanValue = planValue
}

// PlanModifyInt64 implements the Int64 plan modification logic.
func (m MatchElementStateForUnknownModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	genericReq := MatchElementStateForUnknownRequest{
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Plan:           req.Plan,
		PlanValue:      req.PlanValue,
		State:          req.State,
	}
	genericResp := &MatchElementStateForUnknownResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModify(ctx, genericReq, genericResp)

	resp.Diagnostics = genericResp.Diagnostics

	planValue, ok := genericResp.PlanValue.(basetypes.Int64Value)

	if !ok {
		resp.Diagnostics.Append(PlanValueTypeAssertionDiag(req.Path, req.PlanValue, genericResp.PlanValue))
	}

	resp.PlanValue = planValue
}

// PlanModifyList implements the List plan modification logic.
func (m MatchElementStateForUnknownModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	genericReq := MatchElementStateForUnknownRequest{
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Plan:           req.Plan,
		PlanValue:      req.PlanValue,
		State:          req.State,
	}
	genericResp := &MatchElementStateForUnknownResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModify(ctx, genericReq, genericResp)

	resp.Diagnostics = genericResp.Diagnostics

	planValue, ok := genericResp.PlanValue.(basetypes.ListValue)

	if !ok {
		resp.Diagnostics.Append(PlanValueTypeAssertionDiag(req.Path, req.PlanValue, genericResp.PlanValue))
	}

	resp.PlanValue = planValue
}

// PlanModifyMap implements the Map plan modification logic.
func (m MatchElementStateForUnknownModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	genericReq := MatchElementStateForUnknownRequest{
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Plan:           req.Plan,
		PlanValue:      req.PlanValue,
		State:          req.State,
	}
	genericResp := &MatchElementStateForUnknownResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModify(ctx, genericReq, genericResp)

	resp.Diagnostics = genericResp.Diagnostics

	planValue, ok := genericResp.PlanValue.(basetypes.MapValue)

	if !ok {
		resp.Diagnostics.Append(PlanValueTypeAssertionDiag(req.Path, req.PlanValue, genericResp.PlanValue))
	}

	resp.PlanValue = planValue
}

// PlanModifyNumber implements the Number plan modification logic.
func (m MatchElementStateForUnknownModifier) PlanModifyNumber(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
	genericReq := MatchElementStateForUnknownRequest{
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Plan:           req.Plan,
		PlanValue:      req.PlanValue,
		State:          req.State,
	}
	genericResp := &MatchElementStateForUnknownResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModify(ctx, genericReq, genericResp)

	resp.Diagnostics = genericResp.Diagnostics

	planValue, ok := genericResp.PlanValue.(basetypes.NumberValue)

	if !ok {
		resp.Diagnostics.Append(PlanValueTypeAssertionDiag(req.Path, req.PlanValue, genericResp.PlanValue))
	}

	resp.PlanValue = planValue
}

// PlanModifyObject implements the Object plan modification logic.
func (m MatchElementStateForUnknownModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	genericReq := MatchElementStateForUnknownRequest{
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Plan:           req.Plan,
		PlanValue:      req.PlanValue,
		State:          req.State,
	}
	genericResp := &MatchElementStateForUnknownResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModify(ctx, genericReq, genericResp)

	resp.Diagnostics = genericResp.Diagnostics

	planValue, ok := genericResp.PlanValue.(basetypes.ObjectValue)

	if !ok {
		resp.Diagnostics.Append(PlanValueTypeAssertionDiag(req.Path, req.PlanValue, genericResp.PlanValue))
	}

	resp.PlanValue = planValue
}

// PlanModifySet implements the Set plan modification logic.
func (m MatchElementStateForUnknownModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	genericReq := MatchElementStateForUnknownRequest{
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Plan:           req.Plan,
		PlanValue:      req.PlanValue,
		State:          req.State,
	}
	genericResp := &MatchElementStateForUnknownResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModify(ctx, genericReq, genericResp)

	resp.Diagnostics = genericResp.Diagnostics

	planValue, ok := genericResp.PlanValue.(basetypes.SetValue)

	if !ok {
		resp.Diagnostics.Append(PlanValueTypeAssertionDiag(req.Path, req.PlanValue, genericResp.PlanValue))
	}

	resp.PlanValue = planValue
}

// PlanModifyString implements the String plan modification logic.
func (m MatchElementStateForUnknownModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	genericReq := MatchElementStateForUnknownRequest{
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Plan:           req.Plan,
		PlanValue:      req.PlanValue,
		State:          req.State,
	}
	genericResp := &MatchElementStateForUnknownResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModify(ctx, genericReq, genericResp)

	resp.Diagnostics = genericResp.Diagnostics

	planValue, ok := genericResp.PlanValue.(basetypes.StringValue)

	if !ok {
		resp.Diagnostics.Append(PlanValueTypeAssertionDiag(req.Path, req.PlanValue, genericResp.PlanValue))
	}

	resp.PlanValue = planValue
}

// MatchElementStateForUnknownRequest is the shared request type for
// the plan modification logic.
type MatchElementStateForUnknownRequest struct {
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
	Plan           tfsdk.Plan
	PlanValue      attr.Value
	State          tfsdk.State
}

// MatchElementStateForUnknownResponse is the shared response type for
// the plan modification logic.
type MatchElementStateForUnknownResponse struct {
	Diagnostics diag.Diagnostics
	PlanValue   attr.Value
}

// MatchElementStateForUnknownMissingExpressionsDiag returns an error diagnostic
// when the MatchElementStateForUnknown schema plan modifier was passed zero
// expressions.
func MatchElementStateForUnknownMissingExpressionsDiag(p path.Path) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		p,
		"Invalid Attribute Schema",
		"The MatchElementStateForUnknown() plan modifier has no path expressions. "+
			"At least one path expression must be given for matching the prior state. "+
			"For example:\n\n"+
			"MatchElementStateForUnknown(\n"+
			"  path.MatchRelative().AtParent().AtName(\"another_element_attribute\"),\n"+
			"),\n\n"+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("Path: %s", p),
	)
}

// MatchElementStateForUnknownOutsideListOrSetDiag returns an error diagnostic
// intended for when the MatchElementStateForUnknown schema plan modifier is not
// under a list or set.
func MatchElementStateForUnknownOutsideListOrSetDiag(p path.Path) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		p,
		"Invalid Attribute Schema",
		"The MatchElementStateForUnknown() plan modifier is only intended for nested object attributes under a list or set. "+
			"Use the UseStateForUnknown() plan modifier instead. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("Path: %s", p),
	)
}

// MatchElementStateForUnknownInvalidExpressionDiag returns an error diagnostic when
// the MatchElementStateForUnknown schema plan modifier was passed an expression
// that would match data outside the list or set of the current path.
func MatchElementStateForUnknownInvalidExpressionDiag(p path.Path, e path.Expression) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		p,
		"Invalid Attribute Schema",
		"The MatchElementStateForUnknown() plan modifier was given an invalid path expression. "+
			"Expressions should be relative and match a different, identifying, and configurable attribute within the same nested object. "+
			"For example:\n\n"+
			"MatchElementStateForUnknown(\n"+
			"  path.MatchRelative().AtParent().AtName(\"another_element_attribute\"),\n"+
			"),\n\n"+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("Path: %s\n", p)+
			fmt.Sprintf("Given Expression: %s", e),
	)
}

// MatchElementStateForUnknownRootExpressionDiag returns an error diagnostic
// when the MatchElementStateForUnknown schema plan modifier was passed a root
// expression.
func MatchElementStateForUnknownRootExpressionDiag(p path.Path, e path.Expression) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		p,
		"Invalid Attribute Schema",
		"The MatchElementStateForUnknown() plan modifier was given a root path expression. "+
			"Expressions should only be relative and reference attributes at the same level. "+
			"For example:\n\n"+
			"MatchElementStateForUnknown(\n"+
			"  path.MatchRelative().AtParent().AtName(\"another_element_attribute\"),\n"+
			"),\n\n"+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("Path: %s\n", p)+
			fmt.Sprintf("Given Expression: %s", e),
	)
}
