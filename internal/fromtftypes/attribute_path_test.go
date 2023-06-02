// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromtftypes_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromtftypes"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAttributePath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tfType        *tftypes.AttributePath
		schema        fwschema.Schema
		expected      path.Path
		expectedDiags diag.Diagnostics
	}{
		"nil": {
			tfType:   nil,
			expected: path.Empty(),
		},
		"empty": {
			tfType:   tftypes.NewAttributePath(),
			expected: path.Empty(),
		},
		"AttributeName": {
			tfType: tftypes.NewAttributePath().WithAttributeName("test"),
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
						Type: types.StringType,
					},
				},
			},
			expected: path.Root("test"),
		},
		"AttributeName-nonexistent-attribute": {
			tfType: tftypes.NewAttributePath().WithAttributeName("test"),
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"not-test": testschema.Attribute{
						Type: testtypes.StringType{},
					},
				},
			},
			expected: path.Empty(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						"Attribute Path: AttributeName(\"test\")\n"+
						"Original Error: AttributeName(\"test\") still remains in the path: could not find attribute or block \"test\" in schema",
				),
			},
		},
		"AttributeName-ElementKeyInt": {
			tfType: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(1),
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
						Type: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: path.Root("test").AtListIndex(1),
		},
		"AttributeName-ElementKeyValue": {
			tfType: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "test-value")),
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			expected: path.Root("test").AtSetValue(types.StringValue("test-value")),
		},
		"AttributeName-ElementKeyValue-value-conversion-error": {
			tfType: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "test-value")),
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
						Type: types.SetType{
							ElemType: testtypes.InvalidType{},
						},
					},
				},
			},
			expected: path.Empty(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is either an error in terraform-plugin-framework or a custom attribute type used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						"Attribute Path: AttributeName(\"test\").ElementKeyValue(tftypes.String<\"test-value\">)\n"+
						"Original Error: unable to create PathStepElementKeyValue from tftypes.Value: unable to convert tftypes.Value (tftypes.String<\"test-value\">) to attr.Value: intentional ValueFromTerraform error",
				),
			},
		},
		"ElementKeyInt": {
			tfType: tftypes.NewAttributePath().WithElementKeyInt(1),
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
						Type: testtypes.StringType{},
					},
				},
			},
			expected: path.Empty(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						"Attribute Path: ElementKeyInt(1)\n"+
						"Original Error: ElementKeyInt(1) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to schema",
				),
			},
		},
		"ElementKeyString": {
			tfType: tftypes.NewAttributePath().WithElementKeyString("test"),
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
						Type: testtypes.StringType{},
					},
				},
			},
			expected: path.Empty(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						"Attribute Path: ElementKeyString(\"test\")\n"+
						"Original Error: ElementKeyString(\"test\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to schema",
				),
			},
		},
		"ElementKeyValue": {
			tfType: tftypes.NewAttributePath().WithElementKeyValue(tftypes.NewValue(tftypes.String, "test-value")),
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
						Type: testtypes.StringType{},
					},
				},
			},
			expected: path.Empty(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						"Attribute Path: ElementKeyValue(tftypes.String<\"test-value\">)\n"+
						"Original Error: ElementKeyValue(tftypes.String<\"test-value\">) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to schema",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromtftypes.AttributePath(context.Background(), testCase.tfType, testCase.schema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				for _, d := range diags {
					t.Logf("diag: %s", d.Detail())
				}
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
