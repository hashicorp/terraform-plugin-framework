// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestIdentitySchema(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       fwschema.Schema
		expected    *tfprotov5.ResourceIdentitySchema
		expectedErr string
	}

	tests := map[string]testCase{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty-val": {
			input: testschema.Schema{},
			expected: &tfprotov5.ResourceIdentitySchema{
				IdentityAttributes: []*tfprotov5.ResourceIdentitySchemaAttribute{},
				Version:            0,
			},
		},
		"basic-attrs": {
			input: testschema.Schema{
				Version: 1,
				Attributes: map[string]fwschema.Attribute{
					"string": testschema.Attribute{
						Type:              types.StringType,
						RequiredForImport: true,
					},
					"number": testschema.Attribute{
						Type:              types.NumberType,
						OptionalForImport: true,
					},
					"bool": testschema.Attribute{
						Type:              types.BoolType,
						OptionalForImport: true,
					},
				},
			},
			expected: &tfprotov5.ResourceIdentitySchema{
				Version: 1,
				IdentityAttributes: []*tfprotov5.ResourceIdentitySchemaAttribute{
					{
						Name:              "bool",
						Type:              tftypes.Bool,
						OptionalForImport: true,
					},
					{
						Name:              "number",
						Type:              tftypes.Number,
						OptionalForImport: true,
					},
					{
						Name:              "string",
						Type:              tftypes.String,
						RequiredForImport: true,
					},
				},
			},
		},
		"complex-attrs": {
			input: testschema.Schema{
				Version: 2,
				Attributes: map[string]fwschema.Attribute{
					"list_of_string": testschema.Attribute{
						Type:              types.ListType{ElemType: types.StringType},
						RequiredForImport: true,
					},
					"list_of_bool": testschema.Attribute{
						Type:              types.ListType{ElemType: types.BoolType},
						RequiredForImport: true,
					},
				},
			},
			expected: &tfprotov5.ResourceIdentitySchema{
				Version: 2,
				IdentityAttributes: []*tfprotov5.ResourceIdentitySchemaAttribute{
					{
						Name:              "list_of_bool",
						Type:              tftypes.List{ElementType: tftypes.Bool},
						RequiredForImport: true,
					},
					{
						Name:              "list_of_string",
						Type:              tftypes.List{ElementType: tftypes.String},
						RequiredForImport: true,
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := toproto5.IdentitySchema(context.Background(), tc.input)
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
