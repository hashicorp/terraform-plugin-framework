// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror

type FunctionErrors []FunctionError

// AddArgumentError adds an argument error to the collection.
func (f *FunctionErrors) AddArgumentError(functionArgument int, summary string, detail string) {
	f.Append(NewArgumentErrorFunctionError(functionArgument, summary, detail))
}

// AddError adds a generic error to the collection.
func (f *FunctionErrors) AddError(summary string, detail string) {
	f.Append(NewErrorFunctionError(summary, detail))
}

// AddWarning adds a generic warning to the collection.
func (f *FunctionErrors) AddWarning(summary string, detail string) {
	f.Append(NewWarningFunctionError(summary, detail))
}

func (f *FunctionErrors) Append(in ...FunctionError) {
	for _, fe := range in {
		if fe == nil {
			continue
		}

		if f.Contains(fe) {
			continue
		}

		if f == nil {
			*f = FunctionErrors{fe}
		} else {
			*f = append(*f, fe)
		}
	}
}

func (f *FunctionErrors) Contains(in FunctionError) bool {
	if f == nil {
		return false
	}

	for _, fe := range *f {
		if fe.Equal(in) {
			return true
		}
	}

	return false
}

func (f *FunctionErrors) Error() string {
	var errStr string

	if f == nil {
		return ""
	}

	for _, err := range *f {
		errStr += err.Error() + "\n"
	}

	return errStr
}

func (f *FunctionErrors) HasError() bool {
	if f == nil {
		return false
	}

	return len(*f) > 0
}
