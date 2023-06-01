// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mapdefault_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticValueDefaultMap(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal types.Map
		expected   *defaults.MapResponse
	}{
		"map": {
			defaultVal: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"one": types.StringValue("test-value"),
				},
			),
			expected: &defaults.MapResponse{
				PlanValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"one": types.StringValue("test-value"),
					},
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.MapResponse{}

			mapdefault.StaticValue(testCase.defaultVal).DefaultMap(context.Background(), defaults.MapRequest{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
