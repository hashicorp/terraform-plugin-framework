package types

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPrimitiveTerraformType(t *testing.T) {
	t.Parallel()

	tests := map[primitive]tftypes.Type{
		StringType: tftypes.String,
		NumberType: tftypes.Number,
		BoolType:   tftypes.Bool,
	}
	for prim, expected := range tests {
		prim, expected := prim, expected
		t.Run(prim.String(), func(t *testing.T) {
			t.Parallel()

			got := prim.TerraformType(context.Background())
			if !got.Is(expected) {
				t.Errorf("Expected %s, got %s", expected, got)
			}
		})
	}
}

func TestPrimitiveValueFromTerraform(t *testing.T) {
	t.Parallel()

	t.Run(StringType.String(), func(t *testing.T) {
		t.Parallel()

		testStringValueFromTerraform(t, false)
	})

	t.Run(NumberType.String(), func(t *testing.T) {
		t.Parallel()

		testNumberValueFromTerraform(t, false)
	})

	t.Run(BoolType.String(), func(t *testing.T) {
		t.Parallel()

		testBoolValueFromTerraform(t, false)
	})
}

// testAttributeType is a dummy attribute type to compare against with Equal to
// make sure we can handle unexpected types being passed in.
type testAttributeType struct{}

func (t testAttributeType) TerraformType(_ context.Context) tftypes.Type {
	panic("not implemented")
}

func (t testAttributeType) ValueFromTerraform(_ context.Context, _ tftypes.Value) (attr.Value, error) {
	panic("not implemented")
}

func (t testAttributeType) Equal(_ attr.Type) bool {
	panic("not implemented")
}

func (t testAttributeType) ApplyTerraform5AttributePathStep(_ tftypes.AttributePathStep) (interface{}, error) {
	panic("not implemented")
}

func TestPrimitiveEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		prim      primitive
		candidate attr.Type
		expected  bool
	}
	tests := map[string]testCase{
		"string-string": {
			prim:      StringType,
			candidate: StringType,
			expected:  true,
		},
		"string-number": {
			prim:      StringType,
			candidate: NumberType,
			expected:  false,
		},
		"string-bool": {
			prim:      StringType,
			candidate: BoolType,
			expected:  false,
		},
		"string-unknown": {
			prim:      StringType,
			candidate: primitive(100),
			expected:  false,
		},
		"string-wrongType": {
			prim:      StringType,
			candidate: testAttributeType{},
			expected:  false,
		},
		"number-string": {
			prim:      NumberType,
			candidate: StringType,
			expected:  false,
		},
		"number-number": {
			prim:      NumberType,
			candidate: NumberType,
			expected:  true,
		},
		"number-bool": {
			prim:      NumberType,
			candidate: BoolType,
			expected:  false,
		},
		"number-unknown": {
			prim:      NumberType,
			candidate: primitive(100),
			expected:  false,
		},
		"number-wrongType": {
			prim:      NumberType,
			candidate: testAttributeType{},
			expected:  false,
		},
		"bool-string": {
			prim:      BoolType,
			candidate: StringType,
			expected:  false,
		},
		"bool-number": {
			prim:      BoolType,
			candidate: NumberType,
			expected:  false,
		},
		"bool-bool": {
			prim:      BoolType,
			candidate: BoolType,
			expected:  true,
		},
		"bool-unknown": {
			prim:      BoolType,
			candidate: primitive(100),
			expected:  false,
		},
		"bool-wrongType": {
			prim:      BoolType,
			candidate: testAttributeType{},
			expected:  false,
		},
		"unknown-string": {
			prim:      100,
			candidate: StringType,
			expected:  false,
		},
		"unknown-number": {
			prim:      100,
			candidate: NumberType,
			expected:  false,
		},
		"unknown-bool": {
			prim:      100,
			candidate: BoolType,
			expected:  false,
		},
		"unknown-unknown": {
			prim:      100,
			candidate: primitive(100),
			expected:  false,
		},
		"unknown-wrongType": {
			prim:      100,
			candidate: testAttributeType{},
			expected:  false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.prim.Equal(test.candidate)
			if got != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}
