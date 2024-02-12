// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type FunctionError interface {
	Detail() string
	Equal(FunctionError) bool
	Severity() diag.Severity
	Summary() string

	error
}

type FunctionErrorWithFunctionArgument interface {
	FunctionError

	// FunctionArgument points to a specific function argument position.
	//
	// If present, this enables the display of source configuration context for
	// supporting implementations such as Terraform CLI commands.
	FunctionArgument() int
}
