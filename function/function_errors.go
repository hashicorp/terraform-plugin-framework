// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// FunctionErrors represents a collection of function errors.
//
// While this collection is ordered, the order is not guaranteed as reliable
// or consistent.
type FunctionErrors []FunctionError

// AddArgumentError adds an argument error to the collection.
func (fe *FunctionErrors) AddArgumentError(functionArgument int, msg string) {
	fe.Append(NewArgumentFunctionError(functionArgument, msg))
}

// AddError adds a generic error to the collection.
func (fe *FunctionErrors) AddError(msg string) {
	fe.Append(NewFunctionError(msg))
}

// Append adds non-empty and non-duplicate function errors to the collection.
func (fe *FunctionErrors) Append(in ...FunctionError) {
	for _, funcErr := range in {
		if funcErr == nil {
			continue
		}

		if fe.Contains(funcErr) {
			continue
		}

		if funcErr == nil {
			*fe = FunctionErrors{funcErr}
		} else {
			*fe = append(*fe, funcErr)
		}
	}
}

// Contains returns true if the collection contains an equal FunctionError.
func (fe *FunctionErrors) Contains(in FunctionError) bool {
	if fe == nil {
		return false
	}

	for _, funcErr := range *fe {
		if funcErr.Equal(in) {
			return true
		}
	}

	return false
}

// Equal returns true if all given function errors are equivalent in order and
// content, based on the underlying (FunctionError).Equal() method of each.
func (fe *FunctionErrors) Equal(other *FunctionErrors) bool {
	if fe == nil && other == nil {
		return true
	}

	if fe == nil || other == nil {
		return false
	}

	if len(*fe) != len(*other) {
		return false
	}

	o := *other

	for funcErrIndex, funcErr := range *fe {
		if !funcErr.Equal(o[funcErrIndex]) {
			return false
		}
	}

	return true
}

// Error returns a string representation of the collection.
func (fe *FunctionErrors) Error() string {
	var errStr string

	if fe == nil {
		return ""
	}

	for _, err := range *fe {
		errStr += err.Error() + "\n"
	}

	return errStr
}

// HasError returns true if the collection has a FunctionError.
func (fe *FunctionErrors) HasError() bool {
	if fe == nil {
		return false
	}

	return len(*fe) > 0
}

// FunctionErrorsFromDiags iterates over the given diagnostics and returns FunctionErrors populated
// with a FunctionError for each [diag.Diagnostic] with an Error severity. Each warning severity
// [diag.Diagnostic] is logged, but not converted into a FunctionError.
func FunctionErrorsFromDiags(ctx context.Context, diags diag.Diagnostics) FunctionErrors {
	var funcErrs FunctionErrors

	for _, d := range diags {
		switch d.Severity() {
		case diag.SeverityError:
			funcErrs.AddError(fmt.Sprintf("%s: %s", d.Summary(), d.Detail()))
		case diag.SeverityWarning:
			tflog.Warn(ctx, "warning: call function", map[string]interface{}{"summary": d.Summary(), "detail": d.Detail()})
		}
	}

	return funcErrs
}
