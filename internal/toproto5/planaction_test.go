// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestPlanActionResponse(t *testing.T) {
	t.Parallel()

	testDeferral := &action.Deferred{
		Reason: action.DeferredReasonAbsentPrereq,
	}

	testProto5Deferred := &tfprotov5.Deferred{
		Reason: tfprotov5.DeferredReasonAbsentPrereq,
	}

	testLinkedResourceProto5Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_attribute_one": tftypes.String,
			"test_attribute_two": tftypes.Bool,
		},
	}

	testLinkedResourceProto5Value := tftypes.NewValue(testLinkedResourceProto5Type, map[string]tftypes.Value{
		"test_attribute_one": tftypes.NewValue(tftypes.String, "test-value-1"),
		"test_attribute_two": tftypes.NewValue(tftypes.Bool, true),
	})

	testLinkedResourceProto5DynamicValue, err := tfprotov5.NewDynamicValue(testLinkedResourceProto5Type, testLinkedResourceProto5Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov5.NewDynamicValue(): %s", err)
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

	testLinkedResourceIdentityProto5Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id": tftypes.String,
		},
	}

	testLinkedResourceIdentityProto5Value := tftypes.NewValue(testLinkedResourceIdentityProto5Type, map[string]tftypes.Value{
		"test_id": tftypes.NewValue(tftypes.String, "id-123"),
	})

	testLinkedResourceIdentityProto5DynamicValue, err := tfprotov5.NewDynamicValue(testLinkedResourceIdentityProto5Type, testLinkedResourceIdentityProto5Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov5.NewDynamicValue(): %s", err)
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
		expected *tfprotov5.PlanActionResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input: &fwserver.PlanActionResponse{},
			expected: &tfprotov5.PlanActionResponse{
				LinkedResources: []*tfprotov5.PlannedLinkedResource{},
			},
		},
		"linkedresource": {
			input: &fwserver.PlanActionResponse{
				LinkedResources: []*fwserver.PlanActionResponseLinkedResource{
					{
						PlannedState: &tfsdk.State{
							Raw:    testLinkedResourceProto5Value,
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto5Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
			},
			expected: &tfprotov5.PlanActionResponse{
				LinkedResources: []*tfprotov5.PlannedLinkedResource{
					{
						PlannedState: &testLinkedResourceProto5DynamicValue,
						PlannedIdentity: &tfprotov5.ResourceIdentityData{
							IdentityData: &testLinkedResourceIdentityProto5DynamicValue,
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
							Raw:    testLinkedResourceProto5Value,
							Schema: testLinkedResourceSchema,
						},
						PlannedIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto5Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
					{
						PlannedState: &tfsdk.State{
							Raw:    testLinkedResourceProto5Value,
							Schema: testLinkedResourceSchema,
						},
					},
				},
			},
			expected: &tfprotov5.PlanActionResponse{
				LinkedResources: []*tfprotov5.PlannedLinkedResource{
					{
						PlannedState: &testLinkedResourceProto5DynamicValue,
						PlannedIdentity: &tfprotov5.ResourceIdentityData{
							IdentityData: &testLinkedResourceIdentityProto5DynamicValue,
						},
					},
					{
						PlannedState: &testLinkedResourceProto5DynamicValue,
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
			expected: &tfprotov5.PlanActionResponse{
				LinkedResources: []*tfprotov5.PlannedLinkedResource{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "test warning summary",
						Detail:   "test warning details",
					},
					{
						Severity: tfprotov5.DiagnosticSeverityError,
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
			expected: &tfprotov5.PlanActionResponse{
				Deferred:        testProto5Deferred,
				LinkedResources: []*tfprotov5.PlannedLinkedResource{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.PlanActionResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
