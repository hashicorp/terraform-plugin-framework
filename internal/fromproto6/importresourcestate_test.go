package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestImportResourceStateRequest(t *testing.T) {
	t.Parallel()

	testFwSchema := &tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test_attribute": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testFwEmptyState := tfsdk.State{
		Raw:    tftypes.NewValue(testFwSchema.TerraformType(context.Background()), nil),
		Schema: *testFwSchema,
	}

	testCases := map[string]struct {
		input               *tfprotov6.ImportResourceStateRequest
		resourceSchema      *tfsdk.Schema
		resourceType        tfsdk.ResourceType
		expected            *fwserver.ImportResourceStateRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"emptystate": {
			input:          &tfprotov6.ImportResourceStateRequest{},
			resourceSchema: testFwSchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState: testFwEmptyState,
			},
		},
		"emptystate-missing-schema": {
			input:    &tfprotov6.ImportResourceStateRequest{},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Create Empty State",
					"An unexpected error was encountered when creating the empty state. "+
						"This is always an issue in the Terraform Provider SDK used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"id": {
			input: &tfprotov6.ImportResourceStateRequest{
				ID: "test-id",
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState: testFwEmptyState,
				ID:         "test-id",
			},
		},
		"typename": {
			input: &tfprotov6.ImportResourceStateRequest{
				TypeName: "test_resource",
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState: testFwEmptyState,
				TypeName:   "test_resource",
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.ImportResourceStateRequest(context.Background(), testCase.input, testCase.resourceType, testCase.resourceSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}