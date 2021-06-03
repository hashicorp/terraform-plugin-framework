package types

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestListTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    ListType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"list-of-strings": {
			input: ListType{
				ElemType: StringType,
			},
			expected: tftypes.List{
				ElementType: tftypes.String,
			},
		},
		"list-of-list-of-strings": {
			input: ListType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
			expected: tftypes.List{
				ElementType: tftypes.List{
					ElementType: tftypes.String,
				},
			},
		},
		"list-of-list-of-list-of-strings": {
			input: ListType{
				ElemType: ListType{
					ElemType: ListType{
						ElemType: StringType,
					},
				},
			},
			expected: tftypes.List{
				ElementType: tftypes.List{
					ElementType: tftypes.List{
						ElementType: tftypes.String,
					},
				},
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := test.input.TerraformType(context.Background())
			if !got.Is(test.expected) {
				t.Errorf("Expected %s, got %s", test.expected, got)
			}
		})
	}
}

func TestListTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    ListType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"list-of-strings": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			expected: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
		},
		"unknown-list": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, tftypes.UnknownValue),
			expected: List{
				ElemType: StringType,
				Unknown:  true,
			},
		},
		"partially-unknown-list": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			expected: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
		},
		"null-list": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, nil),
			expected: List{
				ElemType: StringType,
				Null:     true,
			},
		},
		"partially-null-list": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, nil),
			}),
			expected: List{
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

func TestListTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver ListType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: ListType{ElemType: StringType},
			input:    ListType{ElemType: StringType},
			expected: true,
		},
		"diff": {
			receiver: ListType{ElemType: StringType},
			input:    ListType{ElemType: NumberType},
			expected: false,
		},
		"wrongType": {
			receiver: ListType{ElemType: StringType},
			input:    NumberType,
			expected: false,
		},
		"nil": {
			receiver: ListType{ElemType: StringType},
			input:    nil,
			expected: false,
		},
		"nil-elem": {
			receiver: ListType{},
			input:    ListType{},
			// ListTypes with nil ElemTypes are invalid, and
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

func TestListElementsAs_stringSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []string
	expected := []string{"hello", "world"}

	err := (List{
		ElemType: StringType,
		Elems: []attr.Value{
			String{Value: "hello"},
			String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestListElementsAs_attributeValueSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []String
	expected := []String{
		{Value: "hello"},
		{Value: "world"},
	}

	err := (List{
		ElemType: StringType,
		Elems: []attr.Value{
			String{Value: "hello"},
			String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestListToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       List
		expectation interface{}
	}
	tests := map[string]testCase{
		"value": {
			input: List{
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
			input:       List{Unknown: true},
			expectation: tftypes.UnknownValue,
		},
		"null": {
			input:       List{Null: true},
			expectation: nil,
		},
		"partial-unknown": {
			input: List{
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
			input: List{
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

func TestListEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver List
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"list-value-list-value": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: true,
		},
		"list-value-diff": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "goodnight"},
					String{Value: "moon"},
				},
			},
			expected: false,
		},
		"list-value-count-diff": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
					String{Value: "test"},
				},
			},
			expected: false,
		},
		"list-value-type-diff": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: List{
				ElemType: BoolType,
				Elems: []attr.Value{
					Bool{Value: false},
					Bool{Value: true},
				},
			},
			expected: false,
		},
		"list-value-unknown": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    List{Unknown: true},
			expected: false,
		},
		"list-value-null": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    List{Null: true},
			expected: false,
		},
		"list-value-wrongType": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"list-value-nil": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    nil,
			expected: false,
		},
		"partially-known-list-value-list-value": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			expected: true,
		},
		"partially-known-list-value-diff": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false,
		},
		"partially-known-list-value-unknown": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input:    List{Unknown: true},
			expected: false,
		},
		"partially-known-list-value-null": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input:    List{Null: true},
			expected: false,
		},
		"partially-known-list-value-wrongType": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"partially-known-list-value-nil": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input:    nil,
			expected: false,
		},
		"partially-null-list-value-list-value": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			expected: true,
		},
		"partially-null-list-value-diff": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false,
		},
		"partially-null-list-value-unknown": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: List{
				Unknown: true,
			},
			expected: false,
		},
		"partially-null-list-value-null": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: List{
				Null: true,
			},
			expected: false,
		},
		"partially-null-list-value-wrongType": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"partially-null-list-value-nil": {
			receiver: List{
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
