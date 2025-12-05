// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

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
