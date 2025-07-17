// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestPlanActionRequest(t *testing.T) {
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

	testUnlinkedSchema := schema.UnlinkedSchema{
		Attributes: map[string]schema.Attribute{
			"test_attribute": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testCases := map[string]struct {
		input               *tfprotov5.PlanActionRequest
		actionSchema        fwschema.Schema
		actionImpl          action.Action
		providerMetaSchema  fwschema.Schema
		expected            *fwserver.PlanActionRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov5.PlanActionRequest{},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing Action Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config-missing-schema": {
			input: &tfprotov5.PlanActionRequest{
				Config: &testProto5DynamicValue,
			},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing Action Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config": {
			input: &tfprotov5.PlanActionRequest{
				Config: &testProto5DynamicValue,
			},
			actionSchema: testUnlinkedSchema,
			expected: &fwserver.PlanActionRequest{
				Config: &tfsdk.Config{
					Raw:    testProto5Value,
					Schema: testUnlinkedSchema,
				},
				ActionSchema: testUnlinkedSchema,
			},
		},
		"client-capabilities": {
			input: &tfprotov5.PlanActionRequest{
				ClientCapabilities: &tfprotov5.PlanActionClientCapabilities{
					DeferralAllowed: true,
				},
			},
			actionSchema: testUnlinkedSchema,
			expected: &fwserver.PlanActionRequest{
				ActionSchema: testUnlinkedSchema,
				ClientCapabilities: action.ModifyPlanClientCapabilities{
					DeferralAllowed: true,
				},
			},
		},
		"client-capabilities-unset": {
			input:        &tfprotov5.PlanActionRequest{},
			actionSchema: testUnlinkedSchema,
			expected: &fwserver.PlanActionRequest{
				ActionSchema: testUnlinkedSchema,
				ClientCapabilities: action.ModifyPlanClientCapabilities{
					DeferralAllowed: false,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.PlanActionRequest(context.Background(), testCase.input, testCase.actionImpl, testCase.actionSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
