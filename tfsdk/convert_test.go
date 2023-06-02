// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
			val: types.StringValue("hello"),
			typ: testtypes.StringType{},
			expected: testtypes.String{
				InternalString: types.StringValue("hello"),
				CreatedBy:      testtypes.StringType{},
			},
		},
		"testtype-string-to-string": {
			val: testtypes.String{
				InternalString: types.StringValue("hello"),
				CreatedBy:      testtypes.StringType{},
			},
			typ:      types.StringType,
			expected: types.StringValue("hello"),
		},
		"string-to-number": {
			val: types.StringValue("hello"),
			typ: types.NumberType,
			expectedDiags: diag.Diagnostics{diag.NewErrorDiagnostic(
				"Error converting value",
				"An unexpected error was encountered converting a basetypes.StringValue to a basetypes.NumberType. This is always a problem with the provider. Please tell the provider developers that basetypes.NumberType returned the following error when calling ValueFromTerraform: can't unmarshal tftypes.String into *big.Float, expected *big.Float",
			)},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := ConvertValue(context.Background(), tc.val, tc.typ)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				for _, diag := range diags {
					t.Logf("Diag summary: %q, Diag details: %q", diag.Summary(), diag.Detail())
				}
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
