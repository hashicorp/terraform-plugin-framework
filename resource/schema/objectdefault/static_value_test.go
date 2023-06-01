// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package objectdefault_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticValueDefaultObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal types.Object
		expected   *defaults.ObjectResponse
	}{
		"object": {
			defaultVal: types.ObjectValueMust(
				map[string]attr.Type{
					"string": types.StringType,
				},
				map[string]attr.Value{
					"string": types.StringValue("test-value"),
				},
			),
			expected: &defaults.ObjectResponse{
				PlanValue: types.ObjectValueMust(
					map[string]attr.Type{
						"string": types.StringType,
					},
					map[string]attr.Value{
						"string": types.StringValue("test-value"),
					},
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.ObjectResponse{}

			objectdefault.StaticValue(testCase.defaultVal).DefaultObject(context.Background(), defaults.ObjectRequest{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
