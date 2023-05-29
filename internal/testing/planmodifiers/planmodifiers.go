// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type TestAttrPlanPrivateModifierGet struct{}

func (t TestAttrPlanPrivateModifierGet) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

	key := "providerKeyOne"
	got, diags := req.Private.GetKey(ctx, key)

	resp.Diagnostics.Append(diags...)

	if string(got) != expected {
		resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
	}
}

func (t TestAttrPlanPrivateModifierGet) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
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

func (t TestAttrPlanPrivateModifierSet) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

	resp.Diagnostics.Append(diags...)
}

func (t TestAttrPlanPrivateModifierSet) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

	resp.Diagnostics.Append(diags...)
}

func (t TestAttrPlanPrivateModifierSet) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

	resp.Diagnostics.Append(diags...)
}

func (t TestAttrPlanPrivateModifierSet) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

	resp.Diagnostics.Append(diags...)
}

func (t TestAttrPlanPrivateModifierSet) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

	resp.Diagnostics.Append(diags...)
}

func (t TestAttrPlanPrivateModifierSet) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t TestAttrPlanPrivateModifierSet) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}
