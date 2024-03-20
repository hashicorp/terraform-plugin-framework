// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamicplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// RequiresReplaceIfConfigured returns a plan modifier that conditionally requires
// resource replacement if:
//
//   - The resource is planned for update.
//   - The plan and state values are not equal.
//   - The configuration value is not null.
//
// Use RequiresReplace if the resource replacement should occur regardless of
// the presence of a configuration value. Use RequiresReplaceIf if the resource
// replacement should check provider-defined conditional logic.
func RequiresReplaceIfConfigured() planmodifier.Dynamic {
	return RequiresReplaceIf(
		func(_ context.Context, req planmodifier.DynamicRequest, resp *RequiresReplaceIfFuncResponse) {
			// This requires checking if the underlying value is also null.
			if req.ConfigValue.IsNull() || req.ConfigValue.IsUnderlyingValueNull() {
				return
			}

			resp.RequiresReplace = true
		},
		"If the value of this attribute is configured and changes, Terraform will destroy and recreate the resource.",
		"If the value of this attribute is configured and changes, Terraform will destroy and recreate the resource.",
	)
}
