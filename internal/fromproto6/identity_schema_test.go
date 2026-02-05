// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestIdentitySchema(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input       *tfprotov6.ResourceIdentitySchema
		expected    *identityschema.Schema
		expectedErr string
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"no-attrs": {
			input: &tfprotov6.ResourceIdentitySchema{},
			expected: &identityschema.Schema{
				Attributes: make(map[string]identityschema.Attribute, 0),
			},
		},
		"primitives-attrs": {
			input: &tfprotov6.ResourceIdentitySchema{
				IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
					{
						Name:              "bool",
						Type:              tftypes.Bool,
						RequiredForImport: true,
					},
					{
						Name:              "number",
						Type:              tftypes.Number,
						OptionalForImport: true,
					},
					{
						Name:              "string",
						Type:              tftypes.String,
						OptionalForImport: true,
					},
				},
			},
			expected: &identityschema.Schema{
				Attributes: map[string]identityschema.Attribute{
					"bool": identityschema.BoolAttribute{
						RequiredForImport: true,
					},
					"number": identityschema.NumberAttribute{
						OptionalForImport: true,
					},
					"string": identityschema.StringAttribute{
						OptionalForImport: true,
					},
				},
			},
		},
		"list-attr": {
			input: &tfprotov6.ResourceIdentitySchema{
				IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
					{
						Name:              "list_of_bools",
						Type:              tftypes.List{ElementType: tftypes.Bool},
						RequiredForImport: true,
					},
				},
			},
			expected: &identityschema.Schema{
				Attributes: map[string]identityschema.Attribute{
					"list_of_bools": identityschema.ListAttribute{
						ElementType:       basetypes.BoolType{},
						RequiredForImport: true,
					},
				},
			},
		},
		"map-error": {
			input: &tfprotov6.ResourceIdentitySchema{
				IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
					{
						Name:              "map_of_strings",
						Type:              tftypes.Map{ElementType: tftypes.String},
						OptionalForImport: true,
					},
				},
			},
			expectedErr: `no supported identity attribute for "map_of_strings", type: tftypes.Map`,
		},
		"set-error": {
			input: &tfprotov6.ResourceIdentitySchema{
				IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
					{
						Name:              "set_of_strings",
						Type:              tftypes.Set{ElementType: tftypes.String},
						OptionalForImport: true,
					},
				},
			},
			expectedErr: `no supported identity attribute for "set_of_strings", type: tftypes.Set`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := fromproto6.IdentitySchema(context.Background(), tc.input)
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
