// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stringdefault_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticStringDefaultString(t *testing.T) {
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

			stringdefault.StaticString(testCase.defaultVal).DefaultString(context.Background(), defaults.StringRequest{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
