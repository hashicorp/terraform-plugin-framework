// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestPlanActionResponse(t *testing.T) {
	t.Parallel()

	testDeferral := &action.Deferred{
		Reason: action.DeferredReasonAbsentPrereq,
	}

	testProto6Deferred := &tfprotov6.Deferred{
		Reason: tfprotov6.DeferredReasonAbsentPrereq,
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

	testCases := map[string]struct {
		input    *fwserver.PlanActionResponse
		expected *tfprotov6.PlanActionResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input: &fwserver.PlanActionResponse{},
			expected: &tfprotov6.PlanActionResponse{
				LinkedResources: []*tfprotov6.PlannedLinkedResource{},
			},
		},
		"linkedresource": {
			input: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionResponseLinkedResource{
					{
						PlannedState: &tfsdk.State{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto6Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
			},
			expected: &tfprotov6.PlanActionResponse{
				LinkedResources: []*tfprotov6.PlannedLinkedResource{
					{
						PlannedState: &testLinkedResourceProto6DynamicValue,
						PlannedIdentity: &tfprotov6.ResourceIdentityData{
							IdentityData: &testLinkedResourceIdentityProto6DynamicValue,
						},
					},
				},
			},
		},
		"linkedresources": {
			input: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionResponseLinkedResource{
					{
						PlannedState: &tfsdk.State{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto6Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
					{
						PlannedState: &tfsdk.State{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
					},
				},
			},
			expected: &tfprotov6.PlanActionResponse{
				LinkedResources: []*tfprotov6.PlannedLinkedResource{
					{
						PlannedState: &testLinkedResourceProto6DynamicValue,
						PlannedIdentity: &tfprotov6.ResourceIdentityData{
							IdentityData: &testLinkedResourceIdentityProto6DynamicValue,
						},
					},
					{
						PlannedState: &testLinkedResourceProto6DynamicValue,
					},
				},
			},
		},
		"diagnostics": {
			input: &fwserver.PlanActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov6.PlanActionResponse{
				LinkedResources: []*tfprotov6.PlannedLinkedResource{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityWarning,
						Summary:  "test warning summary",
						Detail:   "test warning details",
					},
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "test error summary",
						Detail:   "test error details",
					},
				},
			},
		},
		"deferral": {
			input: &fwserver.PlanActionResponse{
				Deferred: testDeferral,
			},
			expected: &tfprotov6.PlanActionResponse{
				Deferred:        testProto6Deferred,
				LinkedResources: []*tfprotov6.PlannedLinkedResource{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.PlanActionResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
