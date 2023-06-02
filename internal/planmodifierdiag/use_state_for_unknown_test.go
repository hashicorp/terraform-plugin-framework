// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifierdiag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/planmodifierdiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func TestUseStateForUnknownUnderListOrSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		expected diag.Diagnostic
	}{
		"test": {
			path: path.Root("test"),
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				"Invalid Attribute Schema",
				"Attributes under a list or set cannot use the UseStateForUnknown() plan modifier. "+
					// TODO: Implement MatchElementStateForUnknown plan modifiers.
					// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/717
					// "Use the MatchElementStateForUnknown() plan modifier instead. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"Path: test\n",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := planmodifierdiag.UseStateForUnknownUnderListOrSet(testCase.path)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
