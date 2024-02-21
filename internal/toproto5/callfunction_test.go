// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
		"diagnostics": {
			input: &fwserver.CallFunctionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("warning summary", "warning detail"),
					diag.NewErrorDiagnostic("error summary", "error detail"),
				},
			},
			expected: &tfprotov5.CallFunctionResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
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

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
