// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestReadResourceResponse(t *testing.T) {
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

	testIdentityProto5Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id": tftypes.String,
		},
	}

	testIdentityProto5Value := tftypes.NewValue(testIdentityProto5Type, map[string]tftypes.Value{
		"test_id": tftypes.NewValue(tftypes.String, "id-123"),
	})

	testIdentityProto5DynamicValue, err := tfprotov5.NewDynamicValue(testIdentityProto5Type, testIdentityProto5Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov5.NewDynamicValue(): %s", err)
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testState := &tfsdk.State{
		Raw: testProto5Value,
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"test_attribute": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	testStateInvalid := &tfsdk.State{
		Raw: testProto5Value,
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"test_attribute": schema.BoolAttribute{
					Required: true,
				},
			},
		},
	}

	testIdentity := &tfsdk.ResourceIdentity{
		Raw: testIdentityProto5Value,
		Schema: identityschema.Schema{
			Attributes: map[string]identityschema.Attribute{
				"test_id": identityschema.StringAttribute{
					RequiredForImport: true,
				},
			},
		},
	}

	testIdentityInvalid := &tfsdk.ResourceIdentity{
		Raw: testIdentityProto5Value,
		Schema: identityschema.Schema{
			Attributes: map[string]identityschema.Attribute{
				"test_id": identityschema.BoolAttribute{
					RequiredForImport: true,
				},
			},
		},
	}

	testDeferral := &resource.Deferred{
		Reason: resource.DeferredReasonAbsentPrereq,
	}

	testProto5Deferred := &tfprotov5.Deferred{
		Reason: tfprotov5.DeferredReasonAbsentPrereq,
	}

	testCases := map[string]struct {
		input    *fwserver.ReadResourceResponse
		expected *tfprotov5.ReadResourceResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.ReadResourceResponse{},
			expected: &tfprotov5.ReadResourceResponse{},
		},
		"diagnostics": {
			input: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov5.ReadResourceResponse{
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
		"diagnostics-invalid-newstate": {
			input: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
				NewState: testStateInvalid,
			},
			expected: &tfprotov5.ReadResourceResponse{
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
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Unable to Convert State",
						Detail: "An unexpected error was encountered when converting the state to the protocol type. " +
							"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n" +
							"Please report this to the provider developer:\n\n" +
							"Unable to create DynamicValue: AttributeName(\"test_attribute\"): unexpected value type string, tftypes.Bool values must be of type bool",
					},
				},
			},
		},
		"diagnostics-invalid-newidentity": {
			input: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
				NewIdentity: testIdentityInvalid,
			},
			expected: &tfprotov5.ReadResourceResponse{
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
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Unable to Convert Resource Identity",
						Detail: "An unexpected error was encountered when converting the resource identity to the protocol type. " +
							"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n" +
							"Please report this to the provider developer:\n\n" +
							"Unable to create DynamicValue: AttributeName(\"test_id\"): unexpected value type string, tftypes.Bool values must be of type bool",
					},
				},
			},
		},
		"newstate": {
			input: &fwserver.ReadResourceResponse{
				NewState: testState,
			},
			expected: &tfprotov5.ReadResourceResponse{
				NewState: &testProto5DynamicValue,
			},
		},
		"newidentity": {
			input: &fwserver.ReadResourceResponse{
				NewIdentity: testIdentity,
			},
			expected: &tfprotov5.ReadResourceResponse{
				NewIdentity: &tfprotov5.ResourceIdentityData{
					IdentityData: &testIdentityProto5DynamicValue,
				},
			},
		},
		"private-empty": {
			input: &fwserver.ReadResourceResponse{
				Private: &privatestate.Data{
					Framework: map[string][]byte{},
					Provider:  testEmptyProviderData,
				},
			},
			expected: &tfprotov5.ReadResourceResponse{
				Private: nil,
			},
		},
		"private": {
			input: &fwserver.ReadResourceResponse{
				Private: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`)},
					Provider: testProviderData,
				},
			},
			expected: &tfprotov5.ReadResourceResponse{
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey":  []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
		},
		"deferral": {
			input: &fwserver.ReadResourceResponse{
				Deferred: testDeferral,
			},
			expected: &tfprotov5.ReadResourceResponse{
				Deferred: testProto5Deferred,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.ReadResourceResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
