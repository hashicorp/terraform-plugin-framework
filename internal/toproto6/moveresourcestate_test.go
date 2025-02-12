// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestMoveResourceStateResponse(t *testing.T) {
	t.Parallel()

	testProto6Value := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"test_attribute": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
		},
	)

	testCases := map[string]struct {
		input    *fwserver.MoveResourceStateResponse
		expected *tfprotov6.MoveResourceStateResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.MoveResourceStateResponse{},
			expected: &tfprotov6.MoveResourceStateResponse{},
		},
		"Diagnostics": {
			input: &fwserver.MoveResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov6.MoveResourceStateResponse{
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
			expected: &tfprotov6.MoveResourceStateResponse{
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
			expected: &tfprotov6.MoveResourceStateResponse{
				TargetPrivate: nil,
			},
		},
		"TargetState": {
			input: &fwserver.MoveResourceStateResponse{
				TargetState: &tfsdk.State{
					Raw: testProto6Value,
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.MoveResourceStateResponse{
				TargetState: DynamicValueMust(testProto6Value),
			},
		},
		"TargetState-invalid": {
			input: &fwserver.MoveResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
				TargetState: &tfsdk.State{
					Raw: testProto6Value,
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.BoolAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.MoveResourceStateResponse{
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
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Unable to Convert State",
						Detail: "An unexpected error was encountered when converting the state to the protocol type. " +
							"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n" +
							"Please report this to the provider developer:\n\n" +
							"Unable to create DynamicValue: AttributeName(\"test_attribute\"): unexpected value type string, tftypes.Bool values must be of type bool",
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.MoveResourceStateResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
