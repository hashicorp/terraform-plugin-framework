// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xfwfunction_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwfunction/xfwfunction"
)

func TestParameter(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		definition        function.Definition
		position          int
		expected          function.Parameter
		expectedFuncError *function.FuncError
	}{
		"none": {
			definition: function.Definition{
				// no Parameters or VariadicParameter
			},
			position: 0,
			expected: nil,
			expectedFuncError: function.NewArgumentFuncError(
				int64(0),
				"Invalid Parameter Position for Definition: "+
					"When determining the parameter for the given argument position, an invalid value was given. "+
					"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
					"Function does not implement parameters.\n"+
					"Given position: 0",
			),
		},
		"parameters-first": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.Int64Parameter{},
					function.StringParameter{},
				},
			},
			position: 0,
			expected: function.BoolParameter{},
		},
		"parameters-last": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.Int64Parameter{},
					function.StringParameter{},
				},
			},
			position: 2,
			expected: function.StringParameter{},
		},
		"parameters-middle": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.Int64Parameter{},
					function.StringParameter{},
				},
			},
			position: 1,
			expected: function.Int64Parameter{},
		},
		"parameters-only": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
			},
			position: 0,
			expected: function.BoolParameter{},
		},
		"parameters-over": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
			},
			position: 1,
			expected: nil,
			expectedFuncError: function.NewArgumentFuncError(
				int64(1),
				"Invalid Parameter Position for Definition: "+
					"When determining the parameter for the given argument position, an invalid value was given. "+
					"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
					"Max argument position: 0\n"+
					"Given position: 1",
			),
		},
		"variadicparameter-and-parameters-select-parameter": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			position: 0,
			expected: function.BoolParameter{},
		},
		"variadicparameter-and-parameters-select-variadicparameter": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			position: 1,
			expected: function.StringParameter{},
		},
		"variadicparameter-only": {
			definition: function.Definition{
				VariadicParameter: function.StringParameter{},
			},
			position: 0,
			expected: function.StringParameter{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, funcError := xfwfunction.Parameter(context.Background(), testCase.definition, testCase.position)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(funcError, testCase.expectedFuncError); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
