// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPlan(t *testing.T) {
	t.Parallel()

	testProto6Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_attribute": tftypes.String,
		},
	}

	testProto6Value := tftypes.NewValue(testProto6Type, map[string]tftypes.Value{
		"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testProto6DynamicValue, err := tfprotov6.NewDynamicValue(testProto6Type, testProto6Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testFwSchema := testschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"test_attribute": testschema.Attribute{
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testFwSchemaInvalid := testschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"test_attribute": testschema.Attribute{
				Required: true,
				Type:     types.BoolType,
			},
		},
	}

	testCases := map[string]struct {
		input               *tfprotov6.DynamicValue
		schema              fwschema.Schema
		expected            *tfsdk.Plan
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"missing-schema": {
			input:    &testProto6DynamicValue,
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Plan",
					"An unexpected error was encountered when converting the plan from the protocol type. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"invalid-schema": {
			input:    &testProto6DynamicValue,
			schema:   testFwSchemaInvalid,
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Plan",
					"An unexpected error was encountered when converting the plan from the protocol type. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Unable to unmarshal DynamicValue: AttributeName(\"test_attribute\"): couldn't decode bool: msgpack: invalid code=aa decoding bool",
				),
			},
		},
		"valid": {
			input:  &testProto6DynamicValue,
			schema: testFwSchema,
			expected: &tfsdk.Plan{
				Raw:    testProto6Value,
				Schema: testFwSchema,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.Plan(context.Background(), testCase.input, testCase.schema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
