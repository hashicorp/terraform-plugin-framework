package numberdefault_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticValueDefaultNumber(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal types.Number
		expected   *defaults.NumberResponse
	}{
		"number": {
			defaultVal: types.NumberValue(big.NewFloat(1.2345)),
			expected: &defaults.NumberResponse{
				PlanValue: types.NumberValue(big.NewFloat(1.2345)),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.NumberResponse{}

			numberdefault.StaticValue(testCase.defaultVal).DefaultNumber(context.Background(), defaults.NumberRequest{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}