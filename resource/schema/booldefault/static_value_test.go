// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package booldefault_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticBoolDefaultBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal bool
		expected   *defaults.BoolResponse
	}{
		"bool": {
			defaultVal: true,
			expected: &defaults.BoolResponse{
				PlanValue: types.BoolValue(true),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.BoolResponse{}

			booldefault.StaticBool(testCase.defaultVal).DefaultBool(context.Background(), defaults.BoolRequest{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
