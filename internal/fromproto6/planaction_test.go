// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestPlanActionRequest(t *testing.T) {
	t.Parallel()

	testEmptyProto6Value := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{})

	testEmptyProto6DynamicValue, err := tfprotov6.NewDynamicValue(tftypes.Object{}, testEmptyProto6Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

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

	testLinkedResourceProto6Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_attribute_one": tftypes.String,
			"test_attribute_two": tftypes.Bool,
		},
	}

	testLinkedResourceProto6Value := tftypes.NewValue(testLinkedResourceProto6Type, map[string]tftypes.Value{
		"test_attribute_one": tftypes.NewValue(tftypes.String, "test-value-1"),
		"test_attribute_two": tftypes.NewValue(tftypes.Bool, true),
	})

	testLinkedResourceProto6DynamicValue, err := tfprotov6.NewDynamicValue(testLinkedResourceProto6Type, testLinkedResourceProto6Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testLinkedResourceSchema := resourceschema.Schema{
		Attributes: map[string]resourceschema.Attribute{
			"test_attribute_one": resourceschema.StringAttribute{
				Required: true,
			},
			"test_attribute_two": resourceschema.BoolAttribute{
				Required: true,
			},
		},
	}

	testLinkedResourceIdentityProto6Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id": tftypes.String,
		},
	}

	testLinkedResourceIdentityProto6Value := tftypes.NewValue(testLinkedResourceIdentityProto6Type, map[string]tftypes.Value{
		"test_id": tftypes.NewValue(tftypes.String, "id-123"),
	})

	testLinkedResourceIdentityProto6DynamicValue, err := tfprotov6.NewDynamicValue(testLinkedResourceIdentityProto6Type, testLinkedResourceIdentityProto6Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testLinkedResourceIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}

	testUnlinkedSchema := actionschema.UnlinkedSchema{
		Attributes: map[string]actionschema.Attribute{
			"test_attribute": actionschema.StringAttribute{
				Required: true,
			},
		},
	}

	testLifecycleSchemaLinked := actionschema.LifecycleSchema{
		Attributes: map[string]actionschema.Attribute{},
		LinkedResource: actionschema.LinkedResource{
			TypeName: "test_linked_resource",
		},
	}

	testCases := map[string]struct {
		input                         *tfprotov6.PlanActionRequest
		actionSchema                  fwschema.Schema
		actionImpl                    action.Action
		linkedResourceSchemas         []fwschema.Schema
		linkedResourceIdentitySchemas []fwschema.Schema
		providerMetaSchema            fwschema.Schema
		expected                      *fwserver.PlanActionRequest
		expectedDiagnostics           diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov6.PlanActionRequest{},
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
			input: &tfprotov6.PlanActionRequest{
				Config: &testProto6DynamicValue,
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
			input: &tfprotov6.PlanActionRequest{
				Config: &testProto6DynamicValue,
			},
			actionSchema: testUnlinkedSchema,
			expected: &fwserver.PlanActionRequest{
				Config: &tfsdk.Config{
					Raw:    testProto6Value,
					Schema: testUnlinkedSchema,
				},
				ActionSchema: testUnlinkedSchema,
			},
		},
		"client-capabilities": {
			input: &tfprotov6.PlanActionRequest{
				ClientCapabilities: &tfprotov6.PlanActionClientCapabilities{
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
			input:        &tfprotov6.PlanActionRequest{},
			actionSchema: testUnlinkedSchema,
			expected: &fwserver.PlanActionRequest{
				ActionSchema: testUnlinkedSchema,
				ClientCapabilities: action.ModifyPlanClientCapabilities{
					DeferralAllowed: false,
				},
			},
		},
		"linkedresource": {
			input: &tfprotov6.PlanActionRequest{
				Config: &testEmptyProto6DynamicValue,
				LinkedResources: []*tfprotov6.ProposedLinkedResource{
					{
						PriorState:   &testLinkedResourceProto6DynamicValue,
						PlannedState: &testLinkedResourceProto6DynamicValue,
						Config:       &testLinkedResourceProto6DynamicValue,
						PriorIdentity: &tfprotov6.ResourceIdentityData{
							IdentityData: &testLinkedResourceIdentityProto6DynamicValue,
						},
					},
				},
			},
			linkedResourceSchemas: []fwschema.Schema{
				testLinkedResourceSchema,
			},
			linkedResourceIdentitySchemas: []fwschema.Schema{
				testLinkedResourceIdentitySchema,
			},
			actionSchema: testLifecycleSchemaLinked,
			expected: &fwserver.PlanActionRequest{
				ActionSchema: testLifecycleSchemaLinked,
				Config: &tfsdk.Config{
					Raw:    testEmptyProto6Value,
					Schema: testLifecycleSchemaLinked,
				},
				LinkedResources: []*fwserver.PlanActionRequestLinkedResource{
					{
						Config: &tfsdk.Config{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PriorIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto6Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
			},
		},
		"linkedresources": {
			input: &tfprotov6.PlanActionRequest{
				Config: &testEmptyProto6DynamicValue,
				LinkedResources: []*tfprotov6.ProposedLinkedResource{
					{
						PriorState:   &testLinkedResourceProto6DynamicValue,
						PlannedState: &testLinkedResourceProto6DynamicValue,
						Config:       &testLinkedResourceProto6DynamicValue,
						PriorIdentity: &tfprotov6.ResourceIdentityData{
							IdentityData: &testLinkedResourceIdentityProto6DynamicValue,
						},
					},
					{
						PriorState:   &testLinkedResourceProto6DynamicValue,
						PlannedState: &testLinkedResourceProto6DynamicValue,
						Config:       &testLinkedResourceProto6DynamicValue,
					},
				},
			},
			linkedResourceSchemas: []fwschema.Schema{
				testLinkedResourceSchema,
				testLinkedResourceSchema,
			},
			linkedResourceIdentitySchemas: []fwschema.Schema{
				testLinkedResourceIdentitySchema,
				nil, // Second resource doesn't have an identity
			},
			actionSchema: testLifecycleSchemaLinked,
			expected: &fwserver.PlanActionRequest{
				ActionSchema: testLifecycleSchemaLinked,
				Config: &tfsdk.Config{
					Raw:    testEmptyProto6Value,
					Schema: testLifecycleSchemaLinked,
				},
				LinkedResources: []*fwserver.PlanActionRequestLinkedResource{
					{
						Config: &tfsdk.Config{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PriorIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto6Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
					{
						Config: &tfsdk.Config{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PlannedState: &tfsdk.Plan{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PriorState: &tfsdk.State{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
					},
				},
			},
		},
		"linkedresources-mismatched-number-of-schemas": {
			input: &tfprotov6.PlanActionRequest{
				Config: &testEmptyProto6DynamicValue,
				LinkedResources: []*tfprotov6.ProposedLinkedResource{
					{
						PriorState:   &testLinkedResourceProto6DynamicValue,
						PlannedState: &testLinkedResourceProto6DynamicValue,
						Config:       &testLinkedResourceProto6DynamicValue,
						PriorIdentity: &tfprotov6.ResourceIdentityData{
							IdentityData: &testLinkedResourceIdentityProto6DynamicValue,
						},
					},
					{
						PriorState:   &testLinkedResourceProto6DynamicValue,
						PlannedState: &testLinkedResourceProto6DynamicValue,
						Config:       &testLinkedResourceProto6DynamicValue,
					},
				},
			},
			linkedResourceSchemas: []fwschema.Schema{
				testLinkedResourceSchema,
			},
			linkedResourceIdentitySchemas: []fwschema.Schema{
				testLinkedResourceIdentitySchema,
				nil, // Second resource doesn't have an identity
			},
			actionSchema: testLifecycleSchemaLinked,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Mismatched Linked Resources in PlanAction Request",
					"An unexpected error was encountered when handling the request. "+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.\n\n"+
						"Received 2 linked resource(s), but the provider was expecting 1 linked resource(s).",
				),
			},
		},
		"linkedresources-mismatched-number-of-identity-schemas": {
			input: &tfprotov6.PlanActionRequest{
				Config: &testEmptyProto6DynamicValue,
				LinkedResources: []*tfprotov6.ProposedLinkedResource{
					{
						PriorState:   &testLinkedResourceProto6DynamicValue,
						PlannedState: &testLinkedResourceProto6DynamicValue,
						Config:       &testLinkedResourceProto6DynamicValue,
						PriorIdentity: &tfprotov6.ResourceIdentityData{
							IdentityData: &testLinkedResourceIdentityProto6DynamicValue,
						},
					},
					{
						PriorState:   &testLinkedResourceProto6DynamicValue,
						PlannedState: &testLinkedResourceProto6DynamicValue,
						Config:       &testLinkedResourceProto6DynamicValue,
					},
				},
			},
			linkedResourceSchemas: []fwschema.Schema{
				testLinkedResourceSchema,
				testLinkedResourceSchema,
			},
			linkedResourceIdentitySchemas: []fwschema.Schema{
				testLinkedResourceIdentitySchema,
			},
			actionSchema: testLifecycleSchemaLinked,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Mismatched Linked Resources in PlanAction Request",
					"An unexpected error was encountered when handling the request. "+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.\n\n"+
						"Received 2 linked resource(s), but the provider was expecting 1 linked resource(s).",
				),
			},
		},
		"linkedresources-no-identity-schema": {
			input: &tfprotov6.PlanActionRequest{
				Config: &testEmptyProto6DynamicValue,
				LinkedResources: []*tfprotov6.ProposedLinkedResource{
					{
						PriorState:   &testLinkedResourceProto6DynamicValue,
						PlannedState: &testLinkedResourceProto6DynamicValue,
						Config:       &testLinkedResourceProto6DynamicValue,
						PriorIdentity: &tfprotov6.ResourceIdentityData{
							IdentityData: &testLinkedResourceIdentityProto6DynamicValue,
						},
					},
				},
			},
			linkedResourceSchemas: []fwschema.Schema{
				testLinkedResourceSchema,
			},
			linkedResourceIdentitySchemas: []fwschema.Schema{
				nil,
			},
			actionSchema: testLifecycleSchemaLinked,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Linked Resource Identity",
					"An unexpected error was encountered when converting a linked resource identity from the protocol type. "+
						"Linked resource (at index 0) contained identity data, but the resource doesn't support identity.\n\n"+
						"This is always a problem with the provider and should be reported to the provider developer.",
				),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.PlanActionRequest(
				context.Background(),
				testCase.input,
				testCase.actionImpl,
				testCase.actionSchema,
				testCase.linkedResourceSchemas,
				testCase.linkedResourceIdentitySchemas,
			)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
