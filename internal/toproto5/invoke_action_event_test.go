// Copyright IBM Corp. 2021, 2025
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

func TestProgressInvokeActionEventType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       fwserver.InvokeProgressEvent
		expected tfprotov5.InvokeActionEvent
	}{
		"message": {
			fw: fwserver.InvokeProgressEvent{
				Message: "hello world",
			},
			expected: tfprotov5.InvokeActionEvent{
				Type: tfprotov5.ProgressInvokeActionEventType{
					Message: "hello world",
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.ProgressInvokeActionEventType(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestCompletedInvokeActionEventType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       *fwserver.InvokeActionResponse
		expected tfprotov5.InvokeActionEvent
	}{
		"diagnostics": {
			fw: &fwserver.InvokeActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: tfprotov5.InvokeActionEvent{
				Type: tfprotov5.CompletedInvokeActionEventType{
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
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.CompletedInvokeActionEventType(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
