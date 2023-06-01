// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifierdiag

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// UseStateForUnknownUnderListOrSet returns an error diagnostic intended for
// when the UseStateForUnknown schema plan modifier is under a list or set.
func UseStateForUnknownUnderListOrSet(p path.Path) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		p,
		"Invalid Attribute Schema",
		"Attributes under a list or set cannot use the UseStateForUnknown() plan modifier. "+
			// TODO: Implement MatchElementStateForUnknown plan modifiers.
			// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/717
			// "Use the MatchElementStateForUnknown() plan modifier instead. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("Path: %s\n", p),
	)
}
