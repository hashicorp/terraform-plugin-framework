package types

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSetTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    SetType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"set-of-strings": {
			input: SetType{
				ElemType: StringType,
			},
			expected: tftypes.Set{
				ElementType: tftypes.String,
			},
		},
		"set-of-set-of-strings": {
			input: SetType{
				ElemType: SetType{
					ElemType: StringType,
				},
			},
			expected: tftypes.Set{
				ElementType: tftypes.Set{
					ElementType: tftypes.String,
				},
			},
		},
		"set-of-set-of-set-of-strings": {
			input: SetType{
				ElemType: SetType{
					ElemType: SetType{
						ElemType: StringType,
					},
				},
			},
			expected: tftypes.Set{
				ElementType: tftypes.Set{
					ElementType: tftypes.Set{
						ElementType: tftypes.String,
					},
				},
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.TerraformType(context.Background())
			if !got.Is(test.expected) {
				t.Errorf("Expected %s, got %s", test.expected, got)
			}
		})
	}
}

func TestSetTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    SetType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"set-of-strings": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
		},
		"unknown-set": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, tftypes.UnknownValue),
			expected: Set{
				ElemType: StringType,
				Unknown:  true,
			},
		},
		"partially-unknown-set": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
		},
		"null-set": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, nil),
			expected: Set{
				ElemType: StringType,
				Null:     true,
			},
		},
		"partially-null-set": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, nil),
			}),
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := test.receiver.ValueFromTerraform(context.Background(), test.input)
			if gotErr != nil {
				if test.expectedErr != "" {
					if gotErr.Error() != test.expectedErr {
						t.Errorf("Expected error to be %q, got %q", test.expectedErr, gotErr.Error())
						return
					}
				}
				t.Errorf("Unexpected error: %s", gotErr.Error())
				return
			}
			if gotErr == nil && test.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", test.expectedErr)
				return
			}
			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("Unexpected diff (-expected, +got): %s", diff)
			}
		})
	}
}

func TestSetTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver SetType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: SetType{ElemType: StringType},
			input:    SetType{ElemType: StringType},
			expected: true,
		},
		"diff": {
			receiver: SetType{ElemType: StringType},
			input:    SetType{ElemType: NumberType},
			expected: false,
		},
		"wrongType": {
			receiver: SetType{ElemType: StringType},
			input:    NumberType,
			expected: false,
		},
		"nil": {
			receiver: SetType{ElemType: StringType},
			input:    nil,
			expected: false,
		},
		"nil-elem": {
			receiver: SetType{},
			input:    SetType{},
			// SetTypes with nil ElemTypes are invalid, and
			// aren't equal to anything
			expected: false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.input)
			if test.expected != got {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestSetElementsAs_stringSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []string
	expected := []string{"hello", "world"}

	diags := (Set{
		ElemType: StringType,
		Elems: []attr.Value{
			String{Value: "hello"},
			String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %s", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestSetElementsAs_attributeValueSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []String
	expected := []String{
		{Value: "hello"},
		{Value: "world"},
	}

	diags := (Set{
		ElemType: StringType,
		Elems: []attr.Value{
			String{Value: "hello"},
			String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %s", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestSetToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Set
		expectation interface{}
	}
	tests := map[string]testCase{
		"value": {
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expectation: []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			},
		},
		"unknown": {
			input:       Set{Unknown: true},
			expectation: tftypes.UnknownValue,
		},
		"null": {
			input:       Set{Null: true},
			expectation: nil,
		},
		"partial-unknown": {
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Unknown: true},
					String{Value: "hello, world"},
				},
			},
			expectation: []tftypes.Value{
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "hello, world"),
			},
		},
		"partial-null": {
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Null: true},
					String{Value: "hello, world"},
				},
			},
			expectation: []tftypes.Value{
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.String, "hello, world"),
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := test.input.ToTerraformValue(context.Background())
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(got, test.expectation); diff != "" {
				t.Errorf("Unexpected result (+got, -expected): %s", diff)
			}
		})
	}
}

func TestSetEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver Set
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"set-value-set-value": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: true,
		},
		"set-value-diff": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "goodnight"},
					String{Value: "moon"},
				},
			},
			expected: false,
		},
		"set-value-count-diff": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
					String{Value: "test"},
				},
			},
			expected: false,
		},
		"set-value-type-diff": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: Set{
				ElemType: BoolType,
				Elems: []attr.Value{
					Bool{Value: false},
					Bool{Value: true},
				},
			},
			expected: false,
		},
		"set-value-unknown": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    Set{Unknown: true},
			expected: false,
		},
		"set-value-null": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    Set{Null: true},
			expected: false,
		},
		"set-value-wrongType": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"set-value-nil": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    nil,
			expected: false,
		},
		"partially-known-set-value-set-value": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			expected: true,
		},
		"partially-known-set-value-diff": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false,
		},
		"partially-known-set-value-unknown": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input:    Set{Unknown: true},
			expected: false,
		},
		"partially-known-set-value-null": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input:    Set{Null: true},
			expected: false,
		},
		"partially-known-set-value-wrongType": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"partially-known-set-value-nil": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input:    nil,
			expected: false,
		},
		"partially-null-set-value-set-value": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			expected: true,
		},
		"partially-null-set-value-diff": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false,
		},
		"partially-null-set-value-unknown": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: Set{
				Unknown: true,
			},
			expected: false,
		},
		"partially-null-set-value-null": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: Set{
				Null: true,
			},
			expected: false,
		},
		"partially-null-set-value-wrongType": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"partially-null-set-value-nil": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input:    nil,
			expected: false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.input)
			if got != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}
