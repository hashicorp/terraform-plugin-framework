// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int32default_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticInt32DefaultInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal int32
		expected   *defaults.Int32Response
	}{
		"int32": {
			defaultVal: 12345,
			expected: &defaults.Int32Response{
				PlanValue: types.Int32Value(12345),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.Int32Response{}

			int32default.StaticInt32(testCase.defaultVal).DefaultInt32(context.Background(), defaults.Int32Request{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
