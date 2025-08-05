// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestProgressInvokeActionEventType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       fwserver.InvokeProgressEvent
		expected tfprotov5.InvokeActionEvent
	}{
		"message": {
			fw: fwserver.InvokeProgressEvent{
				Message: "hello world",
			},
			expected: tfprotov5.InvokeActionEvent{
				Type: tfprotov5.ProgressInvokeActionEventType{
					Message: "hello world",
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.ProgressInvokeActionEventType(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestCompletedInvokeActionEventType(t *testing.T) {
	t.Parallel()

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
		fw       *fwserver.InvokeActionResponse
		expected tfprotov5.InvokeActionEvent
	}{
		"linkedresource": {
			fw: &fwserver.InvokeActionResponse{
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{
					{
						NewState: &tfsdk.State{
							Raw:    testLinkedResourceProto5Value,
							Schema: testLinkedResourceSchema,
						},
						NewIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto5Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
			},
			expected: tfprotov5.InvokeActionEvent{
				Type: tfprotov5.CompletedInvokeActionEventType{
					LinkedResources: []*tfprotov5.NewLinkedResource{
						{
							NewState: &testLinkedResourceProto5DynamicValue,
							NewIdentity: &tfprotov5.ResourceIdentityData{
								IdentityData: &testLinkedResourceIdentityProto5DynamicValue,
							},
						},
					},
				},
			},
		},
		"linkedresources": {
			fw: &fwserver.InvokeActionResponse{
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{
					{
						NewState: &tfsdk.State{
							Raw:    testLinkedResourceProto5Value,
							Schema: testLinkedResourceSchema,
						},
						NewIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto5Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
					{
						NewState: &tfsdk.State{
							Raw:    testLinkedResourceProto5Value,
							Schema: testLinkedResourceSchema,
						},
					},
				},
			},
			expected: tfprotov5.InvokeActionEvent{
				Type: tfprotov5.CompletedInvokeActionEventType{
					LinkedResources: []*tfprotov5.NewLinkedResource{
						{
							NewState: &testLinkedResourceProto5DynamicValue,
							NewIdentity: &tfprotov5.ResourceIdentityData{
								IdentityData: &testLinkedResourceIdentityProto5DynamicValue,
							},
						},
						{
							NewState: &testLinkedResourceProto5DynamicValue,
						},
					},
				},
			},
		},
		"diagnostics": {
			fw: &fwserver.InvokeActionResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: tfprotov5.InvokeActionEvent{
				Type: tfprotov5.CompletedInvokeActionEventType{
					LinkedResources: []*tfprotov5.NewLinkedResource{},
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
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.CompletedInvokeActionEventType(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
