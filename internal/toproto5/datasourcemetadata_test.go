// Copyright IBM Corp. 2021, 2025
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

func TestDataSourceMetadata(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       fwserver.DataSourceMetadata
		expected tfprotov5.DataSourceMetadata
	}{
		"TypeName": {
			fw: fwserver.DataSourceMetadata{
				TypeName: "test",
			},
			expected: tfprotov5.DataSourceMetadata{
				TypeName: "test",
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.DataSourceMetadata(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
