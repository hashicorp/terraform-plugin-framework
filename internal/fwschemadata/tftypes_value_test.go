// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestCreateParentTerraformValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parentType    tftypes.Type
		childValue    interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}{
		"Bool-null": {
			parentType: tftypes.Bool,
			childValue: nil,
			expected:   tftypes.Value{},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Unknown parent type tftypes.Bool to create value.",
				),
			},
		},
		"List-null": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			childValue: nil,
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
		},
		"List-unknown": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			childValue: tftypes.UnknownValue,
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
		},
		"Map-null": {
			parentType: tftypes.Map{
				ElementType: tftypes.String,
			},
			childValue: nil,
			expected: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{}),
		},
		"Map-unknown": {
			parentType: tftypes.Map{
				ElementType: tftypes.String,
			},
			childValue: tftypes.UnknownValue,
			expected: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{}),
		},
		"Object-null": {
			parentType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			},
			childValue: nil,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attrone": tftypes.NewValue(tftypes.String, nil),
				"attrtwo": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"Object-unknown": {
			parentType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			},
			childValue: tftypes.UnknownValue,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attrone": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"attrtwo": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		},
		"Set-null": {
			parentType: tftypes.Set{
				ElementType: tftypes.String,
			},
			childValue: nil,
			expected: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
		},
		"Set-unknown": {
			parentType: tftypes.Set{
				ElementType: tftypes.String,
			},
			childValue: tftypes.UnknownValue,
			expected: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
		},
		"Tuple-null": {
			parentType: tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			},
			childValue: nil,
			expected: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"Tuple-unknown": {
			parentType: tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			},
			childValue: tftypes.UnknownValue,
			expected: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fwschemadata.CreateParentTerraformValue(
				context.Background(),
				path.Root("test"),
				tc.parentType,
				tc.childValue,
			)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestUpsertChildTerraformValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parentType    tftypes.Type
		parentValue   tftypes.Value
		childStep     path.PathStep
		childValue    tftypes.Value
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}{
		"List-empty-write-first": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
			childStep:  path.PathStepElementKeyInt(0),
			childValue: tftypes.NewValue(tftypes.String, "one"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
		},
		"List-empty-write-length-error": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
			childStep:  path.PathStepElementKeyInt(1),
			childValue: tftypes.NewValue(tftypes.String, "two"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 2 as list currently has 0 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"List-null-write-first": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, nil),
			childStep:  path.PathStepElementKeyInt(0),
			childValue: tftypes.NewValue(tftypes.String, "one"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
		},
		"List-null-write-length-error": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, nil),
			childStep:  path.PathStepElementKeyInt(1),
			childValue: tftypes.NewValue(tftypes.String, "two"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, nil),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 2 as list currently has 0 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"List-value-overwrite": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
			childStep:  path.PathStepElementKeyInt(0),
			childValue: tftypes.NewValue(tftypes.String, "new"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "new"),
			}),
		},
		"List-value-write-next": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
			childStep:  path.PathStepElementKeyInt(1),
			childValue: tftypes.NewValue(tftypes.String, "two"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
				tftypes.NewValue(tftypes.String, "two"),
			}),
		},
		"List-value-write-length-error": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
			childStep:  path.PathStepElementKeyInt(2),
			childValue: tftypes.NewValue(tftypes.String, "three"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 3 as list currently has 1 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"Map-empty": {
			parentType: tftypes.Map{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{}),
			childStep:  path.PathStepElementKeyString("key"),
			childValue: tftypes.NewValue(tftypes.String, "value"),
			expected: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"key": tftypes.NewValue(tftypes.String, "value"),
			}),
		},
		"Map-null": {
			parentType: tftypes.Map{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, nil),
			childStep:  path.PathStepElementKeyString("key"),
			childValue: tftypes.NewValue(tftypes.String, "value"),
			expected: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"key": tftypes.NewValue(tftypes.String, "value"),
			}),
		},
		"Map-value-overwrite": {
			parentType: tftypes.Map{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"key": tftypes.NewValue(tftypes.String, "oldvalue"),
			}),
			childStep:  path.PathStepElementKeyString("key"),
			childValue: tftypes.NewValue(tftypes.String, "newvalue"),
			expected: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"key": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"Map-value-write": {
			parentType: tftypes.Map{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"keyone": tftypes.NewValue(tftypes.String, "valueone"),
			}),
			childStep:  path.PathStepElementKeyString("keytwo"),
			childValue: tftypes.NewValue(tftypes.String, "valuetwo"),
			expected: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"keyone": tftypes.NewValue(tftypes.String, "valueone"),
				"keytwo": tftypes.NewValue(tftypes.String, "valuetwo"),
			}),
		},
		"Object-overwrite": {
			parentType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			},
			parentValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attrone": tftypes.NewValue(tftypes.String, "oldvalue"),
				"attrtwo": tftypes.NewValue(tftypes.String, nil),
			}),
			childStep:  path.PathStepAttributeName("attrone"),
			childValue: tftypes.NewValue(tftypes.String, "newvalue"),
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attrone": tftypes.NewValue(tftypes.String, "newvalue"),
				"attrtwo": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"Object-write": {
			parentType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			},
			parentValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attrone": tftypes.NewValue(tftypes.String, nil),
				"attrtwo": tftypes.NewValue(tftypes.String, nil),
			}),
			childStep:  path.PathStepAttributeName("attrone"),
			childValue: tftypes.NewValue(tftypes.String, "attronevalue"),
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attrone": tftypes.NewValue(tftypes.String, "attronevalue"),
				"attrtwo": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"Set-empty": {
			parentType: tftypes.Set{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, nil),
			childStep:  path.PathStepElementKeyValue{Value: types.StringValue("one")},
			childValue: tftypes.NewValue(tftypes.String, "one"),
			expected: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
		},
		"Set-overwrite-value": {
			parentType: tftypes.Set{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
			childStep:  path.PathStepElementKeyValue{Value: types.StringValue("one")},
			childValue: tftypes.NewValue(tftypes.String, "one"),
			expected: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
		},
		"Set-write-value": {
			parentType: tftypes.Set{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
			childStep:  path.PathStepElementKeyValue{Value: types.StringValue("two")},
			childValue: tftypes.NewValue(tftypes.String, "two"),
			expected: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
				tftypes.NewValue(tftypes.String, "two"),
			}),
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fwschemadata.UpsertChildTerraformValue(
				context.Background(),
				path.Root("test"),
				tc.parentValue,
				tc.childStep,
				tc.childValue,
			)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}
