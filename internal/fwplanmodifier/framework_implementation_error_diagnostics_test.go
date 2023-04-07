package fwplanmodifier_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestFrameworkImplementationErrorDiag(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		details  string
		expected diag.Diagnostic
	}{
		"test": {
			path:    path.Root("test"),
			details: "Test details.",
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				"Framework Plan Modifier Implementation Error",
				"A framework-defined plan modifier encountered an unexpected implementation issue which could cause unexpected behavior or panics. "+
					"This is always an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
					"Path: test\n"+
					"Details: Test details.",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwplanmodifier.FrameworkImplementationErrorDiag(testCase.path, testCase.details)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPlanValueTypeAssertionDiag(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path         path.Path
		requestType  attr.Value
		responseType attr.Value
		expected     diag.Diagnostic
	}{
		"test": {
			path:         path.Root("test"),
			requestType:  basetypes.NewStringNull(),
			responseType: basetypes.NewBoolNull(),
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				"Framework Plan Modifier Implementation Error",
				"A framework-defined plan modifier encountered an unexpected implementation issue which could cause unexpected behavior or panics. "+
					"This is always an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
					"Path: test\n"+
					"Details: The shared implementation responded with an unexpected type.\n"+
					"Expected Type: basetypes.StringValue\n"+
					"Response Type: basetypes.BoolValue",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwplanmodifier.PlanValueTypeAssertionDiag(testCase.path, testCase.requestType, testCase.responseType)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
