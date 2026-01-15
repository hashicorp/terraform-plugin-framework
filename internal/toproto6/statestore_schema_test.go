// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	statestoreschema "github.com/hashicorp/terraform-plugin-framework/statestore/schema"
)

func TestStateStoreSchema(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       statestoreschema.Schema
		expected    *tfprotov6.StateStoreSchema
		expectedErr string
	}

	tests := map[string]testCase{
		"empty": {
			input: statestoreschema.Schema{},
			expected: &tfprotov6.StateStoreSchema{
				Schema: &tfprotov6.Schema{Block: &tfprotov6.SchemaBlock{}},
			},
		},
		"valid": {
			input: statestoreschema.Schema{
				Attributes: map[string]statestoreschema.Attribute{
					"bool": statestoreschema.BoolAttribute{
						Optional: true,
					},
					"string": statestoreschema.StringAttribute{
						Required: true,
					},
				},
			},
			expected: &tfprotov6.StateStoreSchema{
				Schema: &tfprotov6.Schema{
					Version: 0,
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "bool",
								Type:     tftypes.Bool,
								Optional: true,
							},
							{
								Name:     "string",
								Type:     tftypes.String,
								Required: true,
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := toproto6.StateStoreSchema(context.Background(), tc.input)
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
