// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func TestListResourceResult(t *testing.T) {
	t.Parallel()

	testListResultData := &fwserver.ListResult{
		Identity: nil,
		Resource: &tfsdk.Resource{
			Schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
		},
		DisplayName: "test-display-name",
		Diagnostics: nil,
	}

	testCases := map[string]struct {
		input    *fwserver.ListResult
		expected tfprotov5.ListResourceResult
	}{
		"nil": {
			input: &fwserver.ListResult{
				Identity:    nil,
				Resource:    nil,
				DisplayName: "",
				Diagnostics: nil,
			},
			expected: tfprotov5.ListResourceResult{
				Identity:    nil,
				Resource:    nil,
				DisplayName: "",
				Diagnostics: nil,
			},
		},
		"valid": {
			input: testListResultData,
			expected: tfprotov5.ListResourceResult{
				DisplayName: "test-display-name",
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.ListResourceResult(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
