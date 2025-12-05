// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestCallFunctionRequest(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input              *tfprotov5.CallFunctionRequest
		function           function.Function
		functionDefinition function.Definition
		expected           *fwserver.CallFunctionRequest
		expectedFuncError  *function.FuncError
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"arguments": {
			input: &tfprotov5.CallFunctionRequest{
				Arguments: []*tfprotov5.DynamicValue{
					DynamicValueMust(tftypes.NewValue(tftypes.Bool, nil)),
					DynamicValueMust(tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)),
					DynamicValueMust(tftypes.NewValue(tftypes.String, "arg2")),
				},
				Name: "testfunction",
			},
			functionDefinition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.Int64Parameter{},
					function.StringParameter{},
				},
				Return: function.StringReturn{},
			},
			expected: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData([]attr.Value{
					basetypes.NewBoolNull(),
					basetypes.NewInt64Unknown(),
					basetypes.NewStringValue("arg2"),
				}),
				FunctionDefinition: function.Definition{
					Parameters: []function.Parameter{
						function.BoolParameter{},
						function.Int64Parameter{},
						function.StringParameter{},
					},
					Return: function.StringReturn{},
				},
			},
		},
		"name": {
			input: &tfprotov5.CallFunctionRequest{
				Name: "testfunction",
			},
			functionDefinition: function.Definition{
				Return: function.StringReturn{},
			},
			expected: &fwserver.CallFunctionRequest{
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.CallFunctionRequest(context.Background(), testCase.input, testCase.function, testCase.functionDefinition)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedFuncError); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
