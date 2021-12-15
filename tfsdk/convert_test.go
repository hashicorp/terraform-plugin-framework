package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val           attr.Value
		typ           attr.Type
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}

	tests := map[string]testCase{
		"string-to-testtype-string": {
			val: types.String{Value: "hello"},
			typ: testtypes.StringType{},
			expected: testtypes.String{
				Str:       types.String{Value: "hello"},
				CreatedBy: testtypes.StringType{},
			},
		},
		"testtype-string-to-string": {
			val: testtypes.String{
				Str:       types.String{Value: "hello"},
				CreatedBy: testtypes.StringType{},
			},
			typ:      types.StringType,
			expected: types.String{Value: "hello"},
		},
		"string-to-number": {
			val: types.String{Value: "hello"},
			typ: types.NumberType,
			expectedDiags: diag.Diagnostics{diag.NewErrorDiagnostic(
				"Error converting value",
				"An unexpected error was encountered converting a types.String to a types.NumberType. This is always a problem with the provider. Please tell the provider developers that types.String is not compatible with types.NumberType.",
			)},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := ConvertValue(context.Background(), tc.val, tc.typ)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Fatalf("Unexpected diff in diags (-wanted, +got): %s", diff)
			}

			if diags.HasError() {
				return
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Fatalf("Unexpected diff in result (-wanted, +got): %s", diff)
			}
		})
	}
}
