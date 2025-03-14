// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestGetResourceIdentitySchemasRequest(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *tfprotov6.GetResourceIdentitySchemasRequest
		expected *fwserver.GetResourceIdentitySchemasRequest
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov6.GetResourceIdentitySchemasRequest{},
			expected: &fwserver.GetResourceIdentitySchemasRequest{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fromproto6.GetResourceIdentitySchemasRequest(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
