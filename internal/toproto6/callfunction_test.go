// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestCallFunctionResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.CallFunctionResponse
		expected *tfprotov6.CallFunctionResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"error": {
			input: &fwserver.CallFunctionResponse{
				Errors: function.FunctionErrors{
					function.NewFunctionError("error summary one: error detail one"),
					function.NewArgumentFunctionError(0, "error summary two: error detail two"),
				},
			},
			expected: &tfprotov6.CallFunctionResponse{
				Error: &tfprotov6.FunctionError{
					Text:             "error summary one: error detail one\nerror summary two: error detail two\n",
					FunctionArgument: pointer(int64(0)),
				},
			},
		},
		"result": {
			input: &fwserver.CallFunctionResponse{
				Result: function.NewResultData(basetypes.NewBoolValue(true)),
			},
			expected: &tfprotov6.CallFunctionResponse{
				Result: DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.CallFunctionResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
