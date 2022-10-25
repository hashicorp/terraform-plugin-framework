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

	if attrVal.ValueString() == "TESTATTRONE" {
		resp.AttributePlan = types.StringValue("TESTATTRTWO")
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

	if attrVal.ValueString() == "TESTATTRTWO" {
		resp.AttributePlan = types.StringValue("MODIFIED_TWO")
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

	if configVal.IsNull() {
		resp.AttributePlan = types.StringValue("DEFAULTVALUE")
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

type TestAttrPlanPrivateModifierGet struct{}

func (t TestAttrPlanPrivateModifierGet) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

	key := "providerKeyOne"
	got, diags := req.Private.GetKey(ctx, key)

	resp.Diagnostics.Append(diags...)

	if string(got) != expected {
		resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
	}
}

func (t TestAttrPlanPrivateModifierGet) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t TestAttrPlanPrivateModifierGet) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type TestAttrPlanPrivateModifierSet struct{}

func (t TestAttrPlanPrivateModifierSet) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

	resp.Diagnostics.Append(diags...)
}

func (t TestAttrPlanPrivateModifierSet) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t TestAttrPlanPrivateModifierSet) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}
