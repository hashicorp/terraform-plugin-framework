// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestMoveResourceStateResponse(t *testing.T) {
	t.Parallel()

	testProto5Value := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"test_attribute": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
		},
	)

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

	testCases := map[string]struct {
		input    *fwserver.MoveResourceStateResponse
		expected *tfprotov5.MoveResourceStateResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.MoveResourceStateResponse{},
			expected: &tfprotov5.MoveResourceStateResponse{},
		},
		"Diagnostics": {
			input: &fwserver.MoveResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov5.MoveResourceStateResponse{
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
		"TargetPrivate": {
			input: &fwserver.MoveResourceStateResponse{
				TargetPrivate: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`)},
					Provider: privatestate.MustProviderData(context.Background(), privatestate.MustMarshalToJson(map[string][]byte{
						"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
					})),
				},
			},
			expected: &tfprotov5.MoveResourceStateResponse{
				TargetPrivate: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey":  []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
		},
		"TargetPrivate-empty": {
			input: &fwserver.MoveResourceStateResponse{
				TargetPrivate: &privatestate.Data{
					Framework: map[string][]byte{},
					Provider:  privatestate.EmptyProviderData(context.Background()),
				},
			},
			expected: &tfprotov5.MoveResourceStateResponse{
				TargetPrivate: nil,
			},
		},
		"TargetState": {
			input: &fwserver.MoveResourceStateResponse{
				TargetState: &tfsdk.State{
					Raw: testProto5Value,
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov5.MoveResourceStateResponse{
				TargetState: DynamicValueMust(testProto5Value),
			},
		},
		"TargetState-invalid": {
			input: &fwserver.MoveResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
				TargetState: &tfsdk.State{
					Raw: testProto5Value,
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.BoolAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov5.MoveResourceStateResponse{
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
		"TargetIdentity": {
			input: &fwserver.MoveResourceStateResponse{
				TargetIdentity: testIdentity,
			},
			expected: &tfprotov5.MoveResourceStateResponse{
				TargetIdentity: &tfprotov5.ResourceIdentityData{
					IdentityData: &testIdentityProto5DynamicValue,
				},
			},
		},
		"TargetIdentity-invalid": {
			input: &fwserver.MoveResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
				TargetIdentity: testIdentityInvalid,
			},
			expected: &tfprotov5.MoveResourceStateResponse{
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
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.MoveResourceStateResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
