// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestResultDataSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		resultData  function.ResultData
		value       any
		expected    attr.Value
		expectedErr function.FunctionErrors
	}{
		"nil": {
			resultData: function.NewResultData(basetypes.NewBoolUnknown()),
			value:      nil,
			expected:   basetypes.NewBoolUnknown(),
			expectedErr: function.FunctionErrors{
				function.NewFunctionError("Value Conversion Error: An unexpected error was encountered trying to convert from value. " +
					"This is always an error in the provider. Please report the following to the provider developer:\n\n" +
					"cannot construct attr.Type from <nil> (invalid)"),
			},
		},
		"invalid-type": {
			resultData: function.NewResultData(basetypes.NewBoolUnknown()),
			value:      basetypes.NewStringValue("test"),
			expected:   basetypes.NewBoolUnknown(),
			expectedErr: function.FunctionErrors{
				function.NewFunctionError("Value Conversion Error: An unexpected error was encountered while verifying an attribute value matched its expected type to prevent unexpected behavior or panics. " +
					"This is always an error in the provider. Please report the following to the provider developer:\n\n" +
					"Expected framework type from provider logic: basetypes.BoolType / underlying type: tftypes.Bool\n" +
					"Received framework type from provider logic: basetypes.StringType / underlying type: tftypes.String\n" +
					"Path: "),
			},
		},
		"framework-type": {
			resultData: function.NewResultData(basetypes.NewBoolUnknown()),
			value:      basetypes.NewBoolValue(true),
			expected:   basetypes.NewBoolValue(true),
		},
		"reflection": {
			resultData: function.NewResultData(basetypes.NewBoolUnknown()),
			value:      true,
			expected:   basetypes.NewBoolValue(true),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.resultData.Set(context.Background(), testCase.value)

			if diff := cmp.Diff(testCase.resultData.Value(), testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(err, testCase.expectedErr); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
