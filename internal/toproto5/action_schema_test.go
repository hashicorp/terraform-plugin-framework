// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestActionSchema(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       actionschema.SchemaType
		expected    *tfprotov5.ActionSchema
		expectedErr string
	}

	tests := map[string]testCase{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"unlinked": {
			input: actionschema.UnlinkedSchema{
				Attributes: map[string]actionschema.Attribute{
					"bool": actionschema.BoolAttribute{
						Optional: true,
					},
					"string": actionschema.StringAttribute{
						Required: true,
					},
				},
				Blocks: map[string]actionschema.Block{
					"single_block": actionschema.SingleNestedBlock{
						Attributes: map[string]actionschema.Attribute{
							"bool": actionschema.BoolAttribute{
								Required: true,
							},
							"string": actionschema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov5.ActionSchema{
				Type: tfprotov5.UnlinkedActionSchemaType{},
				Schema: &tfprotov5.Schema{
					Version: 0,
					Block: &tfprotov5.SchemaBlock{
						Attributes: []*tfprotov5.SchemaAttribute{
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
						BlockTypes: []*tfprotov5.SchemaNestedBlock{
							{
								TypeName: "single_block",
								Block: &tfprotov5.SchemaBlock{
									Attributes: []*tfprotov5.SchemaAttribute{
										{
											Name:     "bool",
											Type:     tftypes.Bool,
											Required: true,
										},
										{
											Name:     "string",
											Type:     tftypes.String,
											Optional: true,
										},
									},
								},
								Nesting: tfprotov5.SchemaNestedBlockNestingModeSingle,
							},
						},
					},
				},
			},
		},
		"lifecycle": {
			input: actionschema.LifecycleSchema{
				ExecutionOrder: actionschema.ExecutionOrderAfter,
				LinkedResource: actionschema.LinkedResource{
					TypeName:    "test_linked_resource",
					Description: "A linked resource for this action",
				},
				Attributes: map[string]actionschema.Attribute{
					"bool": actionschema.BoolAttribute{
						Optional: true,
					},
					"string": actionschema.StringAttribute{
						Required: true,
					},
				},
				Blocks: map[string]actionschema.Block{
					"single_block": actionschema.SingleNestedBlock{
						Attributes: map[string]actionschema.Attribute{
							"bool": actionschema.BoolAttribute{
								Required: true,
							},
							"string": actionschema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov5.ActionSchema{
				Type: tfprotov5.LifecycleActionSchemaType{
					Executes: tfprotov5.LifecycleExecutionOrderAfter,
					LinkedResource: &tfprotov5.LinkedResourceSchema{
						TypeName:    "test_linked_resource",
						Description: "A linked resource for this action",
					},
				},
				Schema: &tfprotov5.Schema{
					Version: 0,
					Block: &tfprotov5.SchemaBlock{
						Attributes: []*tfprotov5.SchemaAttribute{
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
						BlockTypes: []*tfprotov5.SchemaNestedBlock{
							{
								TypeName: "single_block",
								Block: &tfprotov5.SchemaBlock{
									Attributes: []*tfprotov5.SchemaAttribute{
										{
											Name:     "bool",
											Type:     tftypes.Bool,
											Required: true,
										},
										{
											Name:     "string",
											Type:     tftypes.String,
											Optional: true,
										},
									},
								},
								Nesting: tfprotov5.SchemaNestedBlockNestingModeSingle,
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

			got, err := toproto5.ActionSchema(context.Background(), tc.input)
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
