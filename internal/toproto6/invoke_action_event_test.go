// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestProgressInvokeActionEventType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       fwserver.InvokeProgressEvent
		expected tfprotov6.InvokeActionEvent
	}{
		"message": {
			fw: fwserver.InvokeProgressEvent{
				Message: "hello world",
			},
			expected: tfprotov6.InvokeActionEvent{
				Type: tfprotov6.ProgressInvokeActionEventType{
					Message: "hello world",
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.ProgressInvokeActionEventType(context.Background(), testCase.fw)

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
		expected tfprotov6.InvokeActionEvent
	}{
		"diagnostics": {
			fw: &fwserver.InvokeActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: tfprotov6.InvokeActionEvent{
				Type: tfprotov6.CompletedInvokeActionEventType{
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
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.CompletedInvokeActionEventType(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
