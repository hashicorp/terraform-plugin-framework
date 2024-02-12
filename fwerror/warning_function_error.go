// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func NewWarningFunctionError(summary, detail string) FunctionError {
	return &warningFunctionError{
		detail:   detail,
		severity: diag.SeverityWarning,
		summary:  summary,
	}
}

type warningFunctionError struct {
	detail   string
	severity diag.Severity
	summary  string
}

func (f *warningFunctionError) Detail() string {
	return f.detail
}

func (f *warningFunctionError) Equal(other FunctionError) bool {
	var wfe *warningFunctionError

	ok := errors.As(other, &wfe)

	if !ok {
		return false
	}

	if f == nil && wfe == nil {
		return true
	}

	if f == nil || wfe == nil {
		return false
	}

	return f.Detail() == wfe.Detail() && f.Severity() == wfe.Severity() && f.Summary() == wfe.Summary()
}

func (f *warningFunctionError) Error() string {
	return fmt.Sprintf("%s: %s\n\n%s", f.severity, f.summary, f.detail)
}

func (f *warningFunctionError) Severity() diag.Severity {
	return f.severity
}

func (f *warningFunctionError) Summary() string {
	return f.summary
}
