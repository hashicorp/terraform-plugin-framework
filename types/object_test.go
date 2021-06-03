package types

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestObjectTypeTerraformType_simple(t *testing.T) {
	t.Parallel()
	result := ObjectType{
		AttributeTypes: map[string]attr.Type{
			"foo": StringType,
			"bar": NumberType,
			"baz": BoolType,
		},
	}.TerraformType(context.Background())
	if diff := cmp.Diff(result, tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
			"baz": tftypes.Bool,
		},
	}); diff != "" {
		t.Errorf("unexpected result (+expected, -got): %s", diff)
	}
}

func TestObjectTypeTerraformType_empty(t *testing.T) {
	t.Parallel()
	result := ObjectType{}.TerraformType(context.Background())
	if diff := cmp.Diff(result, tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}); diff != "" {
		t.Errorf("unexpected result (+expected, -got): %s", diff)
	}
}

func TestObjectTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    ObjectType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"basic-object": {
			receiver: ObjectType{
				AttributeTypes: map[string]attr.Type{
					"a": StringType,
					"b": BoolType,
					"c": NumberType,
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Bool,
					"c": tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "red"),
				"b": tftypes.NewValue(tftypes.Bool, true),
				"c": tftypes.NewValue(tftypes.Number, 123),
			}),
			expected: &Object{
				Attributes: map[string]attr.Value{
					"a": &String{Value: "red"},
					"b": &Bool{Value: true},
					"c": &Number{Value: big.NewFloat(123)},
				},
				AttributeTypes: map[string]attr.Type{
					"a": StringType,
					"b": BoolType,
					"c": NumberType,
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := test.receiver.ValueFromTerraform(context.Background(), test.input)
			if err != nil {
				if test.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err.Error())
					return
				}
				if err.Error() != test.expectedErr {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, err.Error())
					return
				}
			}
			if test.expectedErr != "" && err == nil {
				t.Errorf("Expected err to be %q, got nil", test.expectedErr)
				return
			}
			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("unexpected result (-expected, +got): %s", diff)
			}
		})
	}
}

func TestObjectTypeEqual(t *testing.T) {
	t.Parallel()
}

func TestObjectAs(t *testing.T) {
	t.Parallel()
}

func TestObjectToTerraformValue(t *testing.T) {
	t.Parallel()
}

func TestObjectEqual(t *testing.T) {
	t.Parallel()
}

func TestObjectSetTerraformValue(t *testing.T) {
	t.Parallel()
}
