// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
)

func TestRenewEphemeralResourceResponse(t *testing.T) {
	t.Parallel()

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testCases := map[string]struct {
		input    *fwserver.RenewEphemeralResourceResponse
		expected *tfprotov5.RenewEphemeralResourceResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input: &fwserver.RenewEphemeralResourceResponse{},
			expected: &tfprotov5.RenewEphemeralResourceResponse{
				// Time zero
				RenewAt: *new(time.Time),
			},
		},
		"diagnostics": {
			input: &fwserver.RenewEphemeralResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov5.RenewEphemeralResourceResponse{
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
		"renew-at": {
			input: &fwserver.RenewEphemeralResourceResponse{
				RenewAt: time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC),
			},
			expected: &tfprotov5.RenewEphemeralResourceResponse{
				RenewAt: time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC),
			},
		},
		"private-empty": {
			input: &fwserver.RenewEphemeralResourceResponse{
				Private: &privatestate.Data{
					Framework: map[string][]byte{},
					Provider:  testEmptyProviderData,
				},
			},
			expected: &tfprotov5.RenewEphemeralResourceResponse{
				Private: nil,
			},
		},
		"private": {
			input: &fwserver.RenewEphemeralResourceResponse{
				Private: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`)},
					Provider: testProviderData,
				},
			},
			expected: &tfprotov5.RenewEphemeralResourceResponse{
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey":  []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.RenewEphemeralResourceResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
