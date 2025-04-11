// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
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

	testIdentityProto6Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_identity_attribute": tftypes.String,
		},
	}

	testIdentityProto6Value := tftypes.NewValue(testIdentityProto6Type, map[string]tftypes.Value{
		"test_identity_attribute": tftypes.NewValue(tftypes.String, "id-123"),
	})

	testIdentityProto6DynamicValue, err := tfprotov6.NewDynamicValue(testIdentityProto6Type, testIdentityProto6Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_identity_attribute": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}

	testFwEmptyState := tfsdk.State{
		Raw:    tftypes.NewValue(testFwSchema.Type().TerraformType(context.Background()), nil),
		Schema: testFwSchema,
	}

	testCases := map[string]struct {
		input               *tfprotov6.ImportResourceStateRequest
		resourceSchema      fwschema.Schema
		identitySchema      fwschema.Schema
		resource            resource.Resource
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
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"identity-missing-schema": {
			input: &tfprotov6.ImportResourceStateRequest{
				Identity: &tfprotov6.ResourceIdentityData{
					IdentityData: &testIdentityProto6DynamicValue,
				},
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState: testFwEmptyState,
			},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Resource Identity",
					"An unexpected error was encountered when converting the resource identity from the protocol type. "+
						"Identity data was sent in the protocol to a resource that doesn't support identity.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"identity": {
			input: &tfprotov6.ImportResourceStateRequest{
				Identity: &tfprotov6.ResourceIdentityData{
					IdentityData: &testIdentityProto6DynamicValue,
				},
			},
			resourceSchema: testFwSchema,
			identitySchema: testIdentitySchema,
			expected: &fwserver.ImportResourceStateRequest{
				EmptyState:     testFwEmptyState,
				IdentitySchema: testIdentitySchema,
				Identity: &tfsdk.ResourceIdentity{
					Raw:    testIdentityProto6Value,
					Schema: testIdentitySchema,
				},
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
		"client-capabilities": {
			input: &tfprotov6.ImportResourceStateRequest{
				ID: "test-id",
				ClientCapabilities: &tfprotov6.ImportResourceStateClientCapabilities{
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
			input: &tfprotov6.ImportResourceStateRequest{
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.ImportResourceStateRequest(context.Background(), testCase.input, testCase.resource, testCase.resourceSchema, testCase.identitySchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
