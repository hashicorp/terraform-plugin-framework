package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TestWarningDiagModifier struct{}

func (t TestWarningDiagModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	resp.Diagnostics.AddWarning(
		"Warning diag",
		"This is a warning",
	)
}

func (t TestWarningDiagModifier) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t TestWarningDiagModifier) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type TestErrorDiagModifier struct{}

func (t TestErrorDiagModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	resp.Diagnostics.AddError(
		"Error diag",
		"This is an error",
	)
}

func (t TestErrorDiagModifier) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t TestErrorDiagModifier) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type TestAttrPlanValueModifierOne struct{}

func (t TestAttrPlanValueModifierOne) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	attrVal, ok := req.AttributePlan.(types.String)
	if !ok {
		return
	}

	if attrVal.Value == "TESTATTRONE" {
		resp.AttributePlan = types.String{
			Value: "TESTATTRTWO",
		}
	}
}

func (t TestAttrPlanValueModifierOne) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t TestAttrPlanValueModifierOne) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type TestAttrPlanValueModifierTwo struct{}

func (t TestAttrPlanValueModifierTwo) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	attrVal, ok := req.AttributePlan.(types.String)
	if !ok {
		return
	}

	if attrVal.Value == "TESTATTRTWO" {
		resp.AttributePlan = types.String{
			Value: "MODIFIED_TWO",
		}
	}
}

func (t TestAttrPlanValueModifierTwo) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t TestAttrPlanValueModifierTwo) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type TestAttrDefaultValueModifier struct{}

func (t TestAttrDefaultValueModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil && req.AttributeConfig == nil {
		return
	}

	configVal := req.AttributeConfig.(types.String)

	if configVal.Null {
		resp.AttributePlan = types.String{Value: "DEFAULTVALUE"}
	}
}

func (t TestAttrDefaultValueModifier) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t TestAttrDefaultValueModifier) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

// testRequiresReplaceModifier is an AttributePlanModifier that sets RequiresReplace
// on the attribute.
type TestRequiresReplaceFalseModifier struct{}

// Modify sets RequiresReplace on the response to true.
func (m TestRequiresReplaceFalseModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	resp.RequiresReplace = false
}

// Description returns a human-readable description of the plan modifier.
func (m TestRequiresReplaceFalseModifier) Description(ctx context.Context) string {
	return "Always unsets requires replace."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m TestRequiresReplaceFalseModifier) MarkdownDescription(ctx context.Context) string {
	return "Always unsets requires replace."
}
