// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUpgradeResourceIdentityResponse(t *testing.T) {
	t.Parallel()

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
		input    *fwserver.UpgradeResourceIdentityResponse
		expected *tfprotov5.UpgradeResourceIdentityResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.UpgradeResourceIdentityResponse{},
			expected: &tfprotov5.UpgradeResourceIdentityResponse{},
		},
		"diagnostics": {
			input: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov5.UpgradeResourceIdentityResponse{
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
		"diagnostics-invalid-upgradedIdentity": {
			input: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
				UpgradedIdentity: testIdentityInvalid,
			},
			expected: &tfprotov5.UpgradeResourceIdentityResponse{
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
		"upgradedIdentity": {
			input: &fwserver.UpgradeResourceIdentityResponse{
				UpgradedIdentity: testIdentity,
			},
			expected: &tfprotov5.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfprotov5.ResourceIdentityData{
					IdentityData: &testIdentityProto5DynamicValue,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.UpgradeResourceIdentityResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
