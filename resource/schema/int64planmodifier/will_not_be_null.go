// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int64planmodifier

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
//		// Will successfully evalutate during plan with a "not null" refinement on "int64_attribute"
//		count = examplecloud_thing.a.int64_attribute != null ? 1 : 0
//
//		// .. resource config
//	}
func WillNotBeNull() planmodifier.Int64 {
	return willNotBeNullModifier{}
}

type willNotBeNullModifier struct{}

func (m willNotBeNullModifier) Description(_ context.Context) string {
	return "Promises the value of this attribute will not be null once it becomes known"
}

func (m willNotBeNullModifier) MarkdownDescription(_ context.Context) string {
	return "Promises the value of this attribute will not be null once it becomes known"
}

func (m willNotBeNullModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
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
