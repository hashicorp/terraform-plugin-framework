// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func NewFunctionError(severity diag.Severity, summary, detail string) error {
	return &functionError{
		detail:   detail,
		severity: severity,
		summary:  summary,
	}
}

type FunctionError interface {
	Detail() string
	Severity() diag.Severity
	Summary() string

	error
}

type functionError struct {
	detail   string
	severity diag.Severity
	summary  string
}

func (e *functionError) Detail() string {
	return e.detail
}

func (e *functionError) Error() string {
	return fmt.Sprintf("%s: %s\n\n%s", e.severity, e.summary, e.detail)
}

func (e *functionError) Severity() diag.Severity {
	return e.severity
}

func (e *functionError) Summary() string {
	return e.summary
}
