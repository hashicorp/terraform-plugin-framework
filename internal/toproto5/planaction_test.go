// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
)

func TestPlanActionResponse(t *testing.T) {
	t.Parallel()

	testDeferral := &action.Deferred{
		Reason: action.DeferredReasonAbsentPrereq,
	}

	testProto5Deferred := &tfprotov5.Deferred{
		Reason: tfprotov5.DeferredReasonAbsentPrereq,
	}

	testCases := map[string]struct {
		input    *fwserver.PlanActionResponse
		expected *tfprotov5.PlanActionResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.PlanActionResponse{},
			expected: &tfprotov5.PlanActionResponse{},
		},
		"diagnostics": {
			input: &fwserver.PlanActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov5.PlanActionResponse{
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
		"deferral": {
			input: &fwserver.PlanActionResponse{
				Deferred: testDeferral,
			},
			expected: &tfprotov5.PlanActionResponse{
				Deferred: testProto5Deferred,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.PlanActionResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
