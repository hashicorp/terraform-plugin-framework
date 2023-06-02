// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func TestErrorDiagnostic(path path.Path) diag.DiagnosticWithPath {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Error Diagnostic",
		"This is an error.",
	)
}

func TestWarningDiagnostic(path path.Path) diag.DiagnosticWithPath {
	return diag.NewAttributeWarningDiagnostic(
		path,
		"Warning Diagnostic",
		"This is a warning.",
	)
}
