// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package setdefault_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticValueDefaultSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal types.Set
		expected   *defaults.SetResponse
	}{
		"set": {
			defaultVal: types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("test-value"),
				},
			),
			expected: &defaults.SetResponse{
				PlanValue: types.SetValueMust(
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

			resp := &defaults.SetResponse{}

			setdefault.StaticValue(testCase.defaultVal).DefaultSet(context.Background(), defaults.SetRequest{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
