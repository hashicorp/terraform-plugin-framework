// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package float32default_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32default"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStaticFloat32DefaultFloat32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultVal float32
		expected   *defaults.Float32Response
	}{
		"float32": {
			defaultVal: 1.2345,
			expected: &defaults.Float32Response{
				PlanValue: types.Float32Value(1.2345),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &defaults.Float32Response{}

			float32default.StaticFloat32(testCase.defaultVal).DefaultFloat32(context.Background(), defaults.Float32Request{}, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
