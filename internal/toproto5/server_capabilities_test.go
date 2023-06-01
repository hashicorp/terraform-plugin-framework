// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func TestServerCapabilities(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       *fwserver.ServerCapabilities
		expected *tfprotov5.ServerCapabilities
	}{
		"nil": {
			fw:       nil,
			expected: nil,
		},
		"PlanDestroy": {
			fw: &fwserver.ServerCapabilities{
				PlanDestroy: true,
			},
			expected: &tfprotov5.ServerCapabilities{
				PlanDestroy: true,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.ServerCapabilities(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
