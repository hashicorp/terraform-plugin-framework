// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package boolplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// WillNotBeNull returns a plan modifier that will add a refinement to an unknown planned value
// which promises that the final value will not be null.
//
// This unknown value refinement allows Terraform to validate more of the configuration during plan
// and evaluate conditional logic in meta-arguments such as "count":
//
//	resource "examplecloud_thing" "b" {
//		// Will successfully evaluate during plan with a "not null" refinement on "bool_attribute"
//		count = examplecloud_thing.a.bool_attribute != null ? 1 : 0
//
//		// .. resource config
//	}
func WillNotBeNull() planmodifier.Bool {
	return willNotBeNullModifier{}
}

type willNotBeNullModifier struct{}

func (m willNotBeNullModifier) Description(_ context.Context) string {
	return "Promises the value of this attribute will not be null once it becomes known"
}

func (m willNotBeNullModifier) MarkdownDescription(_ context.Context) string {
	return "Promises the value of this attribute will not be null once it becomes known"
}

func (m willNotBeNullModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.RefineAsNotNull()
}
