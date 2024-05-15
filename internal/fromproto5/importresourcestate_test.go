// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestImportResourceStateRequest(t *testing.T) {
	t.Parallel()

	testFwSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_attribute": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testFwEmptyState := tfsdk.State{
		Raw:    tftypes.NewValue(testFwSchema.Type().TerraformType(context.Background()), nil),
		Schema: testFwSchema,
	}

	testCases := map[string]struct {
		input               *tfprotov5.ImportResourceStateRequest
		resourceSchema      fwschema.Schema
		resource            resource.Resource
		expected            *fwserver.ImportResourceStateRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"emptystate": {
			input:          &tfprotov5.ImportResourceStateRequest{},
			resourceSchema: testFwSchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState: testFwEmptyState,
			},
		},
		"emptystate-missing-schema": {
			input:    &tfprotov5.ImportResourceStateRequest{},
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
		"id": {
			input: &tfprotov5.ImportResourceStateRequest{
				ID: "test-id",
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState: testFwEmptyState,
				ID:         "test-id",
			},
		},
		"typename": {
			input: &tfprotov5.ImportResourceStateRequest{
				TypeName: "test_resource",
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState: testFwEmptyState,
				TypeName:   "test_resource",
			},
		},
		"client-capabilities": {
			input: &tfprotov5.ImportResourceStateRequest{
				ID: "test-id",
				ClientCapabilities: &tfprotov5.ImportResourceStateClientCapabilities{
					DeferralAllowed: true,
				},
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState: testFwEmptyState,
				ID:         "test-id",
				ClientCapabilities: resource.ImportStateClientCapabilities{
					DeferralAllowed: true,
				},
			},
		},
		"client-capabilities-unset": {
			input: &tfprotov5.ImportResourceStateRequest{
				ID: "test-id",
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState: testFwEmptyState,
				ID:         "test-id",
				ClientCapabilities: resource.ImportStateClientCapabilities{
					DeferralAllowed: false,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.ImportResourceStateRequest(context.Background(), testCase.input, testCase.resource, testCase.resourceSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
