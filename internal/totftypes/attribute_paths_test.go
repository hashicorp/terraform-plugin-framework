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

func TestAttributePaths(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		paths         path.Paths
		expected      []*tftypes.AttributePath
		expectedDiags diag.Diagnostics
	}{
		"nil": {
			paths:    nil,
			expected: nil,
		},
		"empty": {
			paths:    path.Paths{},
			expected: []*tftypes.AttributePath{},
		},
		"one": {
			paths: path.Paths{
				path.Root("test"),
			},
			expected: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("test"),
			},
		},
		"one-diagnostics": {
			paths: path.Paths{
				path.Root("test").AtSetValue(testtypes.Invalid{}),
			},
			expected: []*tftypes.AttributePath{},
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
		"two": {
			paths: path.Paths{
				path.Root("test1").AtListIndex(1).AtName("test1_nested"),
				path.Root("test2").AtMapKey("test-key2"),
			},
			expected: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("test1").WithElementKeyInt(1).WithAttributeName("test1_nested"),
				tftypes.NewAttributePath().WithAttributeName("test2").WithElementKeyString("test-key2"),
			},
		},
		"two-diagnostics": {
			paths: path.Paths{
				path.Root("test1"),
				path.Root("test2").AtSetValue(testtypes.Invalid{}),
			},
			expected: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("test1"),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is either an error in terraform-plugin-framework or a custom attribute type used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						"Attribute Path: test2[Value(<invalid>)]\n"+
						"Original Error: unable to convert attr.Value (<invalid>) to tftypes.Value: intentional ToTerraformValue error",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := totftypes.AttributePaths(context.Background(), testCase.paths)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
