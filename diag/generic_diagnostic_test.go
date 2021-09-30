package diag_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestGenericDiagnosticEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diag     *diag.GenericDiagnostic
		other    diag.Diagnostic
		expected bool
	}{
		"matching-attribute-error": {
			diag:     diag.NewAttributeErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("test"), "test summary", "test detail"),
			other:    diag.NewAttributeErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("test"), "test summary", "test detail"),
			expected: true,
		},
		"matching-attribute-warning": {
			diag:     diag.NewAttributeWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test"), "test summary", "test detail"),
			other:    diag.NewAttributeWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test"), "test summary", "test detail"),
			expected: true,
		},
		"matching-error": {
			diag:     diag.NewErrorDiagnostic("test summary", "test detail"),
			other:    diag.NewErrorDiagnostic("test summary", "test detail"),
			expected: true,
		},
		"matching-warning": {
			diag:     diag.NewWarningDiagnostic("test summary", "test detail"),
			other:    diag.NewWarningDiagnostic("test summary", "test detail"),
			expected: true,
		},
		"nil": {
			diag:     diag.NewWarningDiagnostic("test summary", "test detail"),
			other:    nil,
			expected: false,
		},
		"no-path": {
			diag:     diag.NewAttributeWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test"), "test summary", "test detail"),
			other:    diag.NewWarningDiagnostic("test summary", "test detail"),
			expected: false,
		},
		"different-detail": {
			diag:     diag.NewWarningDiagnostic("test summary", "test detail"),
			other:    diag.NewWarningDiagnostic("test summary", "different detail"),
			expected: false,
		},
		"different-path": {
			diag:     diag.NewAttributeWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test"), "test summary", "test detail"),
			other:    diag.NewAttributeWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("other"), "test summary", "test detail"),
			expected: false,
		},
		"different-summary": {
			diag:     diag.NewWarningDiagnostic("test summary", "test detail"),
			other:    diag.NewWarningDiagnostic("different summary", "test detail"),
			expected: false,
		},
		"different-type": {
			diag:     diag.NewWarningDiagnostic("test summary", "test detail"),
			other:    diag.NewErrorDiagnostic("test summary", "test detail"),
			expected: false,
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.diag.Equal(tc.other)

			if got != tc.expected {
				t.Errorf("Unexpected response: got: %t, wanted: %t", got, tc.expected)
			}
		})
	}
}
