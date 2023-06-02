// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listdefault_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticValueDefaultList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal types.List
		expected   *defaults.ListResponse
	}{
		"list": {
			defaultVal: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("test-value"),
				},
			),
			expected: &defaults.ListResponse{
				PlanValue: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("test-value"),
					},
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.ListResponse{}

			listdefault.StaticValue(testCase.defaultVal).DefaultList(context.Background(), defaults.ListRequest{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
