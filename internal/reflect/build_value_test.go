// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tfValue       tftypes.Value
		expectedDiags diag.Diagnostics
	}{
		"unhandled-null": {
			tfValue: tftypes.NewValue(tftypes.String, nil),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("id"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received null value, however the target type cannot handle null values. Use the corresponding `types` package type, a pointer type or a custom type that handles null values.\n\n"+
						"Path: id\nTarget Type: string\nSuggested `types` Type: basetypes.StringValue\nSuggested Pointer Type: *string",
				),
			},
		},
		"unhandled-unknown": {
			tfValue: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("id"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: id\nTarget Type: string\nSuggested Type: basetypes.StringValue",
				),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var s string
			_, diags := refl.BuildValue(context.Background(), types.StringType, tc.tfValue, reflect.ValueOf(s), refl.Options{}, path.Root("id"))

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}
