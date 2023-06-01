// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package totftypes_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/internal/totftypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAttributePath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw            path.Path
		expected      *tftypes.AttributePath
		expectedDiags diag.Diagnostics
	}{
		"empty": {
			fw:       path.Empty(),
			expected: tftypes.NewAttributePath(),
		},
		"one": {
			fw:       path.Root("test"),
			expected: tftypes.NewAttributePath().WithAttributeName("test"),
		},
		"two": {
			fw:       path.Root("test").AtListIndex(1),
			expected: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(1),
		},
		"step-error": {
			fw:       path.Root("test").AtSetValue(testtypes.Invalid{}),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is either an error in terraform-plugin-framework or a custom attribute type used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						"Attribute Path: test[Value(<invalid>)]\n"+
						"Original Error: unable to convert attr.Value (<invalid>) to tftypes.Value: intentional ToTerraformValue error",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := totftypes.AttributePath(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
