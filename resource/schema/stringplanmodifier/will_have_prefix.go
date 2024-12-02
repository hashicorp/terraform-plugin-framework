// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stringplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// WillHavePrefix returns a plan modifier that will add a refinement to an unknown planned value
// which promises that:
//   - The final value will not be null.
//   - The final value will have a specified string prefix.
//
// This unknown value refinement allows Terraform to validate more of the configuration during plan
// and evaluate conditional logic in meta-arguments such as "count". String prefixes that exceed 256
// characters in length will be truncated and empty string prefixes will be ignored.
func WillHavePrefix(prefix string) planmodifier.String {
	return willHavePrefixModifier{
		prefix: prefix,
	}
}

type willHavePrefixModifier struct {
	prefix string
}

func (m willHavePrefixModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will have the prefix %q once it becomes known", m.prefix)
}

func (m willHavePrefixModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will have the prefix %q once it becomes known", m.prefix)
}

func (m willHavePrefixModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.RefineWithPrefix(m.prefix)
}
