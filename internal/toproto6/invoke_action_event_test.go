// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestProgressInvokeActionEventType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       fwserver.InvokeProgressEvent
		expected tfprotov6.InvokeActionEvent
	}{
		"message": {
			fw: fwserver.InvokeProgressEvent{
				Message: "hello world",
			},
			expected: tfprotov6.InvokeActionEvent{
				Type: tfprotov6.ProgressInvokeActionEventType{
					Message: "hello world",
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.ProgressInvokeActionEventType(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestCompletedInvokeActionEventType(t *testing.T) {
	t.Parallel()

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
		fw       *fwserver.InvokeActionResponse
		expected tfprotov6.InvokeActionEvent
	}{
		"linkedresource": {
			fw: &fwserver.InvokeActionResponse{
				LinkedResources: []*fwserver.InvokeActionResponseLinkedResource{
					{
						NewState: &tfsdk.State{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						NewIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto6Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
				},
			},
			expected: tfprotov6.InvokeActionEvent{
				Type: tfprotov6.CompletedInvokeActionEventType{
					LinkedResources: []*tfprotov6.NewLinkedResource{
						{
							NewState: &testLinkedResourceProto6DynamicValue,
							NewIdentity: &tfprotov6.ResourceIdentityData{
								IdentityData: &testLinkedResourceIdentityProto6DynamicValue,
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
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
						NewIdentity: &tfsdk.ResourceIdentity{
							Raw:    testLinkedResourceIdentityProto6Value,
							Schema: testLinkedResourceIdentitySchema,
						},
					},
					{
						NewState: &tfsdk.State{
							Raw:    testLinkedResourceProto6Value,
							Schema: testLinkedResourceSchema,
						},
					},
				},
			},
			expected: tfprotov6.InvokeActionEvent{
				Type: tfprotov6.CompletedInvokeActionEventType{
					LinkedResources: []*tfprotov6.NewLinkedResource{
						{
							NewState: &testLinkedResourceProto6DynamicValue,
							NewIdentity: &tfprotov6.ResourceIdentityData{
								IdentityData: &testLinkedResourceIdentityProto6DynamicValue,
							},
						},
						{
							NewState: &testLinkedResourceProto6DynamicValue,
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
			expected: tfprotov6.InvokeActionEvent{
				Type: tfprotov6.CompletedInvokeActionEventType{
					LinkedResources: []*tfprotov6.NewLinkedResource{},
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
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.CompletedInvokeActionEventType(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
