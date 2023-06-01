// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float64default_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticFloat64DefaultFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal float64
		expected   *defaults.Float64Response
	}{
		"float64": {
			defaultVal: 1.2345,
			expected: &defaults.Float64Response{
				PlanValue: types.Float64Value(1.2345),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.Float64Response{}

			float64default.StaticFloat64(testCase.defaultVal).DefaultFloat64(context.Background(), defaults.Float64Request{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
