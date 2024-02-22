// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// NewFuncError returns a new function error with the
// given message.
func NewFuncError(text string) *FuncError {
	return &FuncError{
		Text: text,
	}
}

// NewArgumentFuncError returns a new function error with the
// given message and function argument.
func NewArgumentFuncError(functionArgument int64, text string) *FuncError {
	return &FuncError{
		Text:             text,
		FunctionArgument: &functionArgument,
	}
}

// FuncError is an error type specifically for function errors.
type FuncError struct {
	Text             string
	FunctionArgument *int64
}

// Equal returns true if the other function error is wholly equivalent.
func (fe *FuncError) Equal(other *FuncError) bool {
	var funcErr *FuncError

	ok := errors.As(other, &funcErr)

	if !ok {
		return false
	}

	if fe == nil && funcErr == nil {
		return true
	}

	if fe == nil || funcErr == nil {
		return false
	}

	if fe.Text != funcErr.Text {
		return false
	}

	if fe.FunctionArgument == nil && funcErr.FunctionArgument == nil {
		return true
	}

	if fe.FunctionArgument == nil || funcErr.FunctionArgument == nil {
		return false
	}

	return *fe.FunctionArgument == *funcErr.FunctionArgument
}

// Error returns the error text.
func (fe *FuncError) Error() string {
	return fe.Text
}

// HasError returns true if the function error is not nil, and
// has either text, or function argument fields populated .
func (fe *FuncError) HasError() bool {
	if fe == nil {
		return false
	}

	return fe.Text != "" || fe.FunctionArgument != nil
}

// ConcatFuncErrors returns a new function error with the text from all supplied
// function errors concatenated together. If any of the function errors have a
// function argument, the first one encountered will be used.
func ConcatFuncErrors(funcErr *FuncError, funcErrs ...*FuncError) *FuncError {
	var text string
	var functionArgument *int64

	if funcErr != nil {
		text = funcErr.Text
		functionArgument = funcErr.FunctionArgument
	}

	for _, f := range funcErrs {
		if f == nil {
			continue
		}

		if text != "" && f.Text != "" {
			text += "\n"
		}

		text += f.Text

		if functionArgument == nil {
			functionArgument = f.FunctionArgument
		}
	}

	if text != "" || functionArgument != nil {
		return &FuncError{
			Text:             text,
			FunctionArgument: functionArgument,
		}
	}

	return nil
}

// FuncErrorFromDiags iterates over the given diagnostics and returns a new function error
// with the text from all error diagnostics concatenated together. If any of the error diagnostics
// have a function argument, the first one encountered will be used.
func FuncErrorFromDiags(ctx context.Context, diags diag.Diagnostics) *FuncError {
	var funcErr *FuncError

	for _, d := range diags {
		switch d.Severity() {
		case diag.SeverityError:
			funcErr = ConcatFuncErrors(funcErr, NewFuncError(fmt.Sprintf("%s: %s", d.Summary(), d.Detail())))
		case diag.SeverityWarning:
			tflog.Warn(ctx, "warning: call function", map[string]interface{}{"summary": d.Summary(), "detail": d.Detail()})
		}
	}

	if funcErr.HasError() {
		return funcErr
	}

	return nil
}
