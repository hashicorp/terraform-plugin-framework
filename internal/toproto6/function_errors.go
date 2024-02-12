// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/fwerror"
)

// FunctionError converts the function errors into the tfprotov6 function error.
func FunctionError(ctx context.Context, funcErrs fwerror.FunctionErrors) *tfprotov6.FunctionError {
	var text string
	var funcArg *int64

	for _, funcErr := range funcErrs {
		text += fmt.Sprintf("%s: %s\n\n%s\n\n", funcErr.Severity(), funcErr.Summary(), funcErr.Detail())

		var funcErrWithFunctionArgument fwerror.FunctionErrorWithFunctionArgument

		if errors.As(funcErr, &funcErrWithFunctionArgument) && funcArg == nil {
			fArg := int64(funcErrWithFunctionArgument.FunctionArgument())

			funcArg = &fArg
		}
	}

	if text == "" {
		return nil
	}

	return &tfprotov6.FunctionError{
		Text:             text,
		FunctionArgument: funcArg,
	}
}
