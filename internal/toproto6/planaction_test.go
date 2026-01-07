// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
)

func TestPlanActionResponse(t *testing.T) {
	t.Parallel()

	testDeferral := &action.Deferred{
		Reason: action.DeferredReasonAbsentPrereq,
	}

	testProto6Deferred := &tfprotov6.Deferred{
		Reason: tfprotov6.DeferredReasonAbsentPrereq,
	}

	testCases := map[string]struct {
		input    *fwserver.PlanActionResponse
		expected *tfprotov6.PlanActionResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.PlanActionResponse{},
			expected: &tfprotov6.PlanActionResponse{},
		},
		"diagnostics": {
			input: &fwserver.PlanActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov6.PlanActionResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityWarning,
						Summary:  "test warning summary",
						Detail:   "test warning details",
					},
					{
						Severity: tfprotov6.DiagnosticSeverityError,
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
			expected: &tfprotov6.PlanActionResponse{
				Deferred: testProto6Deferred,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.PlanActionResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
