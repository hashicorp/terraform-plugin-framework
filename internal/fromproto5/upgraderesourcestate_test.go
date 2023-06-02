// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func TestUpgradeResourceStateRequest(t *testing.T) {
	t.Parallel()

	testFwSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_attribute": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testCases := map[string]struct {
		input               *tfprotov5.UpgradeResourceStateRequest
		resourceSchema      fwschema.Schema
		resource            resource.Resource
		expected            *fwserver.UpgradeResourceStateRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"rawstate": {
			input: &tfprotov5.UpgradeResourceStateRequest{
				RawState: testNewTfprotov5RawState(t, map[string]interface{}{
					"test_attribute": "test-value",
				}),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.UpgradeResourceStateRequest{
				RawState: testNewTfprotov6RawState(t, map[string]interface{}{
					"test_attribute": "test-value",
				}),
				ResourceSchema: testFwSchema,
			},
		},
		"resourceschema": {
			input:          &tfprotov5.UpgradeResourceStateRequest{},
			resourceSchema: testFwSchema,
			expected: &fwserver.UpgradeResourceStateRequest{
				ResourceSchema: testFwSchema,
			},
		},
		"resourceschema-missing": {
			input:    &tfprotov5.UpgradeResourceStateRequest{},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Create Empty State",
					"An unexpected error was encountered when creating the empty state. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"version": {
			input: &tfprotov5.UpgradeResourceStateRequest{
				Version: 123,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.UpgradeResourceStateRequest{
				ResourceSchema: testFwSchema,
				Version:        123,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.UpgradeResourceStateRequest(context.Background(), testCase.input, testCase.resource, testCase.resourceSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
