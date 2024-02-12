// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func NewErrorFunctionError(summary, detail string) FunctionError {
	return &errorFunctionError{
		detail:   detail,
		severity: diag.SeverityError,
		summary:  summary,
	}
}

type errorFunctionError struct {
	detail   string
	severity diag.Severity
	summary  string
}

func (f *errorFunctionError) Detail() string {
	return f.detail
}

func (f *errorFunctionError) Equal(other FunctionError) bool {
	var efe *errorFunctionError

	ok := errors.As(other, &efe)

	if !ok {
		return false
	}

	if f == nil && efe == nil {
		return true
	}

	if f == nil || efe == nil {
		return false
	}

	return f.Detail() == efe.Detail() && f.Severity() == efe.Severity() && f.Summary() == efe.Summary()
}

func (f *errorFunctionError) Error() string {
	return fmt.Sprintf("%s: %s\n\n%s", f.severity, f.summary, f.detail)
}

func (f *errorFunctionError) Severity() diag.Severity {
	return f.severity
}

func (f *errorFunctionError) Summary() string {
	return f.summary
}

func FunctionErrorsFromDiags(diags diag.Diagnostics) FunctionErrors {
	var funcErrs FunctionErrors

	for _, d := range diags {
		switch d.Severity() {
		case diag.SeverityError:
			funcErrs.AddError(d.Summary(), d.Detail())
		case diag.SeverityWarning:
			funcErrs.AddWarning(d.Summary(), d.Detail())
		}
	}

	return funcErrs
}
