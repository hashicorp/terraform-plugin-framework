// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int64default_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticInt64DefaultInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal int64
		expected   *defaults.Int64Response
	}{
		"int64": {
			defaultVal: 12345,
			expected: &defaults.Int64Response{
				PlanValue: types.Int64Value(12345),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.Int64Response{}

			int64default.StaticInt64(testCase.defaultVal).DefaultInt64(context.Background(), defaults.Int64Request{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
