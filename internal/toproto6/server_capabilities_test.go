// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestServerCapabilities(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       *fwserver.ServerCapabilities
		expected *tfprotov6.ServerCapabilities
	}{
		"nil": {
			fw:       nil,
			expected: nil,
		},
		"GetProviderSchemaOptional": {
			fw: &fwserver.ServerCapabilities{
				GetProviderSchemaOptional: true,
			},
			expected: &tfprotov6.ServerCapabilities{
				GetProviderSchemaOptional: true,
			},
		},
		"MoveResourceState": {
			fw: &fwserver.ServerCapabilities{
				MoveResourceState: true,
			},
			expected: &tfprotov6.ServerCapabilities{
				MoveResourceState: true,
			},
		},
		"PlanDestroy": {
			fw: &fwserver.ServerCapabilities{
				PlanDestroy: true,
			},
			expected: &tfprotov6.ServerCapabilities{
				PlanDestroy: true,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.ServerCapabilities(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
