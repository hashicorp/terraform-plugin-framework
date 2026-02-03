// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestIdentitySchemaAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		attr        fwschema.Attribute
		path        *tftypes.AttributePath
		expected    *tfprotov6.ResourceIdentitySchemaAttribute
		expectedErr string
	}

	tests := map[string]testCase{
		"description": {
			name: "string",
			attr: testschema.Attribute{
				Type:              types.StringType,
				RequiredForImport: true,
				Description:       "A string attribute",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.ResourceIdentitySchemaAttribute{
				Name:              "string",
				Type:              tftypes.String,
				RequiredForImport: true,
				Description:       "A string attribute",
			},
		},
		"attr-string": {
			name: "string",
			attr: testschema.Attribute{
				Type:              types.StringType,
				RequiredForImport: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.ResourceIdentitySchemaAttribute{
				Name:              "string",
				Type:              tftypes.String,
				RequiredForImport: true,
			},
		},
		"attr-bool": {
			name: "bool",
			attr: testschema.Attribute{
				Type:              types.BoolType,
				RequiredForImport: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.ResourceIdentitySchemaAttribute{
				Name:              "bool",
				Type:              tftypes.Bool,
				RequiredForImport: true,
			},
		},
		"attr-number": {
			name: "number",
			attr: testschema.Attribute{
				Type:              types.NumberType,
				RequiredForImport: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.ResourceIdentitySchemaAttribute{
				Name:              "number",
				Type:              tftypes.Number,
				RequiredForImport: true,
			},
		},
		"attr-list": {
			name: "list",
			attr: testschema.Attribute{
				Type:              types.ListType{ElemType: types.NumberType},
				RequiredForImport: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.ResourceIdentitySchemaAttribute{
				Name:              "list",
				Type:              tftypes.List{ElementType: tftypes.Number},
				RequiredForImport: true,
			},
		},
		"requiredforimport": {
			name: "string",
			attr: testschema.Attribute{
				Type:              types.StringType,
				RequiredForImport: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.ResourceIdentitySchemaAttribute{
				Name:              "string",
				Type:              tftypes.String,
				RequiredForImport: true,
			},
		},
		"optionalforimport": {
			name: "string",
			attr: testschema.Attribute{
				Type:              types.StringType,
				OptionalForImport: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.ResourceIdentitySchemaAttribute{
				Name:              "string",
				Type:              tftypes.String,
				OptionalForImport: true,
			},
		},
		"nested-attr-single-error": {
			name: "single_nested",
			attr: testschema.NestedAttribute{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
						"computed": testschema.Attribute{
							Type:      types.NumberType,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
				NestingMode: fwschema.NestingModeSingle,
				Optional:    true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "identity schemas don't support NestedAttribute",
		},
		"nested-attr-list-error": {
			name: "list_nested",
			attr: testschema.NestedAttribute{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
						"computed": testschema.Attribute{
							Type:      types.NumberType,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
				NestingMode: fwschema.NestingModeList,
				Optional:    true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "identity schemas don't support NestedAttribute",
		},
		"nested-attr-map-error": {
			name: "map_nested",
			attr: testschema.NestedAttribute{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
						"computed": testschema.Attribute{
							Type:      types.NumberType,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
				NestingMode: fwschema.NestingModeMap,
				Optional:    true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "identity schemas don't support NestedAttribute",
		},
		"nested-attr-set-error": {
			name: "set_nested",
			attr: testschema.NestedAttribute{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
						"computed": testschema.Attribute{
							Type:      types.NumberType,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
				NestingMode: fwschema.NestingModeSet,
				Optional:    true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "identity schemas don't support NestedAttribute",
		},
		"attr-unset": {
			name: "whoops",
			attr: testschema.Attribute{
				Optional: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have Type set",
		},
		"missing-requiredforimport-and-optionalforimport": {
			name: "whoops",
			attr: testschema.Attribute{
				Type: types.StringType,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have RequiredForImport or OptionalForImport set",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := toproto6.IdentitySchemaAttribute(context.Background(), tc.name, tc.path, tc.attr)
			if err != nil {
				if tc.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if err.Error() != tc.expectedErr {
					t.Errorf("Expected error to be %q, got %q", tc.expectedErr, err.Error())
					return
				}
				// got expected error
				return
			}
			if tc.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", tc.expectedErr)
				return
			}
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
				return
			}
		})
	}
}
