// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamicdefault_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticValueDefaultDynamic(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal types.Dynamic
		expected   *defaults.DynamicResponse
	}{
		"dynamic": {
			defaultVal: types.DynamicValue(types.StringValue("test value")),
			expected: &defaults.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringValue("test value")),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.DynamicResponse{}

			dynamicdefault.StaticValue(testCase.defaultVal).DefaultDynamic(context.Background(), defaults.DynamicRequest{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
