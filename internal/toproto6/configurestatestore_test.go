// Copyright IBM Corp. 2021, 2025
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

func TestConfigureStateStoreResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.ConfigureStateStoreResponse
		expected *tfprotov6.ConfigureStateStoreResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.ConfigureStateStoreResponse{},
			expected: &tfprotov6.ConfigureStateStoreResponse{},
		},
		"diagnostics": {
			input: &fwserver.ConfigureStateStoreResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov6.ConfigureStateStoreResponse{
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
		"server-capabilities": {
			input: &fwserver.ConfigureStateStoreResponse{
				ServerCapabilities: &fwserver.StateStoreServerCapabilities{
					ChunkSize: 4 << 20,
				},
			},
			expected: &tfprotov6.ConfigureStateStoreResponse{
				Capabilities: &tfprotov6.StateStoreServerCapabilities{
					ChunkSize: 4 << 20,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.ConfigureStateStoreResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
