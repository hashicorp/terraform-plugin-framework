package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\nunhandled null value",
				),
			},
		},
		"unhandled-unknown": {
			tfValue: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\nunhandled unknown value",
				),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var s string
			_, diags := refl.BuildValue(context.Background(), types.StringType, tc.tfValue, reflect.ValueOf(s), refl.Options{}, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}
