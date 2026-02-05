// Copyright IBM Corp. 2021, 2026
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

func TestEphemeralResourceMetadata(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       fwserver.EphemeralResourceMetadata
		expected tfprotov6.EphemeralResourceMetadata
	}{
		"TypeName": {
			fw: fwserver.EphemeralResourceMetadata{
				TypeName: "test",
			},
			expected: tfprotov6.EphemeralResourceMetadata{
				TypeName: "test",
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.EphemeralResourceMetadata(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
