// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func TestValidateActionConfigResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.ValidateActionConfigResponse
		expected *tfprotov5.ValidateActionConfigResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.ValidateActionConfigResponse{},
			expected: &tfprotov5.ValidateActionConfigResponse{},
		},
		"diagnostics": {
			input: &fwserver.ValidateActionConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov5.ValidateActionConfigResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "test warning summary",
						Detail:   "test warning details",
					},
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "test error summary",
						Detail:   "test error details",
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.ValidateActionConfigResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
