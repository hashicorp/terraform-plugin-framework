// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDataSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		data          fwschemadata.Data
		val           any
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"write": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "oldvalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"overwrite": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.Value{},
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"multiple-attributes": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.Value{},
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"one": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
						"two": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
						"three": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				One   types.String `tfsdk:"one"`
				Two   *string      `tfsdk:"two"`
				Three string       `tfsdk:"three"`
			}{
				One:   types.StringUnknown(),
				Two:   nil,
				Three: "value3",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"one":   tftypes.String,
					"two":   tftypes.String,
					"three": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"one":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"two":   tftypes.NewValue(tftypes.String, nil),
				"three": tftypes.NewValue(tftypes.String, "value3"),
			}),
		},
		"AttrTypeWithValidateError": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.Value{},
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected:      tftypes.Value{},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(path.Root("name"))},
		},
		"AttrTypeWithValidateWarning": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.Value{},
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(path.Root("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.data.Set(context.Background(), tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.data.TerraformValue, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
