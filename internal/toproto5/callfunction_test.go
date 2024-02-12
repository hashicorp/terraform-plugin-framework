// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/fwerror"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestCallFunctionResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.CallFunctionResponse
		expected *tfprotov5.CallFunctionResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"error": {
			input: &fwserver.CallFunctionResponse{
				Error: fwerror.FunctionErrors{
					fwerror.NewWarningFunctionError("warning summary", "warning detail"),
					fwerror.NewErrorFunctionError("error summary", "error detail"),
				},
			},
			expected: &tfprotov5.CallFunctionResponse{
				Error: &tfprotov5.FunctionError{
					Text: "Warning: warning summary\n\nwarning detail\n\nError: error summary\n\nerror detail\n\n",
				},
			},
		},
		"result": {
			input: &fwserver.CallFunctionResponse{
				Result: function.NewResultData(basetypes.NewBoolValue(true)),
			},
			expected: &tfprotov5.CallFunctionResponse{
				Result: DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.CallFunctionResponse(context.Background(), testCase.input)

			// Handling error comparison
			equateErrors := cmp.Comparer(func(x, y error) bool {
				if x == nil || y == nil {
					return x == nil && y == nil
				}

				return x.Error() == y.Error()
			})

			if diff := cmp.Diff(got, testCase.expected, equateErrors); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
