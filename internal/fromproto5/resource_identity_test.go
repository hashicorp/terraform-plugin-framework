// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestResourceIdentity(t *testing.T) {
	t.Parallel()

	testProto5Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_attribute": tftypes.String,
		},
	}

	testProto5Value := tftypes.NewValue(testProto5Type, map[string]tftypes.Value{
		"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testProto5DynamicValue, err := tfprotov5.NewDynamicValue(testProto5Type, testProto5Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov5.NewDynamicValue(): %s", err)
	}

	testFwSchema := testschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"test_attribute": testschema.Attribute{
				RequiredForImport: true,
				Type:              types.StringType,
			},
		},
	}

	testFwSchemaInvalid := testschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"test_attribute": testschema.Attribute{
				RequiredForImport: true,
				Type:              types.BoolType,
			},
		},
	}

	testCases := map[string]struct {
		input               *tfprotov5.ResourceIdentityData
		schema              fwschema.Schema
		expected            *tfsdk.ResourceIdentity
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov5.ResourceIdentityData{},
			expected: nil,
		},
		"missing-schema": {
			input: &tfprotov5.ResourceIdentityData{
				IdentityData: &testProto5DynamicValue,
			},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Resource Identity",
					"An unexpected error was encountered when converting the resource identity from the protocol type. "+
						"Identity data was sent in the protocol to a resource that doesn't support identity.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"invalid-schema": {
			input: &tfprotov5.ResourceIdentityData{
				IdentityData: &testProto5DynamicValue,
			},
			schema:   testFwSchemaInvalid,
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Resource Identity",
					"An unexpected error was encountered when converting the resource identity from the protocol type. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Unable to unmarshal DynamicValue: AttributeName(\"test_attribute\"): couldn't decode bool: msgpack: invalid code=aa decoding bool",
				),
			},
		},
		"valid": {
			input: &tfprotov5.ResourceIdentityData{
				IdentityData: &testProto5DynamicValue,
			},
			schema: testFwSchema,
			expected: &tfsdk.ResourceIdentity{
				Raw:    testProto5Value,
				Schema: testFwSchema,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.ResourceIdentity(context.Background(), testCase.input, testCase.schema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
