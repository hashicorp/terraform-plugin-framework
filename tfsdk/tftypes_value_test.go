package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestCreateParentValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parentType    tftypes.Type
		parentValue   tftypes.Value
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}{
		"Bool-null": {
			parentType:  tftypes.Bool,
			parentValue: tftypes.NewValue(tftypes.Bool, nil),
			expected:    tftypes.NewValue(tftypes.Bool, nil),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Unknown parent type tftypes.primitive to create value.",
				),
			},
		},
		"List-null": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, nil),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
		},
		"List-unknown": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, tftypes.UnknownValue),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
		},
		"List-value": {
			parentType: tftypes.List{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
				tftypes.NewValue(tftypes.String, "two"),
			}),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
				tftypes.NewValue(tftypes.String, "two"),
			}),
		},
		"Map-null": {
			parentType: tftypes.Map{
				AttributeType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, nil),
			expected: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{}),
		},
		"Map-unknown": {
			parentType: tftypes.Map{
				AttributeType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, tftypes.UnknownValue),
			expected: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{}),
		},
		"Map-value": {
			parentType: tftypes.Map{
				AttributeType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{
				"keyone": tftypes.NewValue(tftypes.String, "valueone"),
				"keytwo": tftypes.NewValue(tftypes.String, "valuetwo"),
			}),
			expected: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{
				"keyone": tftypes.NewValue(tftypes.String, "valueone"),
				"keytwo": tftypes.NewValue(tftypes.String, "valuetwo"),
			}),
		},
		"Object-null": {
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
			}, nil),
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
			parentValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			}, tftypes.UnknownValue),
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
		"Object-value": {
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
				"attrone": tftypes.NewValue(tftypes.String, "one"),
				"attrtwo": tftypes.NewValue(tftypes.String, "two"),
			}),
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"attrone": tftypes.String,
					"attrtwo": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"attrone": tftypes.NewValue(tftypes.String, "one"),
				"attrtwo": tftypes.NewValue(tftypes.String, "two"),
			}),
		},
		"Set-null": {
			parentType: tftypes.Set{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, nil),
			expected: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
		},
		"Set-unknown": {
			parentType: tftypes.Set{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, tftypes.UnknownValue),
			expected: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
		},
		"Set-value": {
			parentType: tftypes.Set{
				ElementType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
				tftypes.NewValue(tftypes.String, "two"),
			}),
			expected: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
				tftypes.NewValue(tftypes.String, "two"),
			}),
		},
		"Tuple-null": {
			parentType: tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			},
			parentValue: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			}, nil),
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
			parentValue: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			}, tftypes.UnknownValue),
			expected: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		},
		"Tuple-value": {
			parentType: tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			},
			parentValue: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
				tftypes.NewValue(tftypes.String, "two"),
			}),
			expected: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
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

			got, diags := createParentValue(
				context.Background(),
				tftypes.NewAttributePath().WithAttributeName("test"),
				tc.parentType,
				tc.parentValue,
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

func TestUpsertChildValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parentType    tftypes.Type
		parentValue   tftypes.Value
		childStep     tftypes.AttributePathStep
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
			childStep:  tftypes.ElementKeyInt(0),
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
			childStep:  tftypes.ElementKeyInt(1),
			childValue: tftypes.NewValue(tftypes.String, "two"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test"),
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
			childStep:  tftypes.ElementKeyInt(0),
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
			childStep:  tftypes.ElementKeyInt(1),
			childValue: tftypes.NewValue(tftypes.String, "two"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, nil),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test"),
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
			childStep:  tftypes.ElementKeyInt(0),
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
			childStep:  tftypes.ElementKeyInt(1),
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
			childStep:  tftypes.ElementKeyInt(2),
			childValue: tftypes.NewValue(tftypes.String, "three"),
			expected: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 3 as list currently has 1 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"Map-empty": {
			parentType: tftypes.Map{
				AttributeType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{}),
			childStep:  tftypes.ElementKeyString("key"),
			childValue: tftypes.NewValue(tftypes.String, "value"),
			expected: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{
				"key": tftypes.NewValue(tftypes.String, "value"),
			}),
		},
		"Map-null": {
			parentType: tftypes.Map{
				AttributeType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, nil),
			childStep:  tftypes.ElementKeyString("key"),
			childValue: tftypes.NewValue(tftypes.String, "value"),
			expected: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{
				"key": tftypes.NewValue(tftypes.String, "value"),
			}),
		},
		"Map-value-overwrite": {
			parentType: tftypes.Map{
				AttributeType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{
				"key": tftypes.NewValue(tftypes.String, "oldvalue"),
			}),
			childStep:  tftypes.ElementKeyString("key"),
			childValue: tftypes.NewValue(tftypes.String, "newvalue"),
			expected: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{
				"key": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"Map-value-write": {
			parentType: tftypes.Map{
				AttributeType: tftypes.String,
			},
			parentValue: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
			}, map[string]tftypes.Value{
				"keyone": tftypes.NewValue(tftypes.String, "valueone"),
			}),
			childStep:  tftypes.ElementKeyString("keytwo"),
			childValue: tftypes.NewValue(tftypes.String, "valuetwo"),
			expected: tftypes.NewValue(tftypes.Map{
				AttributeType: tftypes.String,
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
			childStep:  tftypes.AttributeName("attrone"),
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
			childStep:  tftypes.AttributeName("attrone"),
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
			childStep:  tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "one")),
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
			childStep:  tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "one")),
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
			childStep:  tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "two")),
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

			got, diags := upsertChildValue(
				context.Background(),
				tftypes.NewAttributePath().WithAttributeName("test"),
				tc.parentType,
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
