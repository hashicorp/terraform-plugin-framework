package stringdefault_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticValueDefaultString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal string
		expected   *defaults.StringResponse
	}{
		"string": {
			defaultVal: "test-value",
			expected: &defaults.StringResponse{
				PlanValue: types.StringValue("test-value"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.StringResponse{}

			stringdefault.StaticValue(testCase.defaultVal).DefaultString(context.Background(), defaults.StringRequest{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
