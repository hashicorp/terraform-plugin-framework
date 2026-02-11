// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUpgradeResourceIdentityResponse(t *testing.T) {
	t.Parallel()

	testIdentityProto6Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id": tftypes.String,
		},
	}

	testIdentityProto6Value := tftypes.NewValue(testIdentityProto6Type, map[string]tftypes.Value{
		"test_id": tftypes.NewValue(tftypes.String, "id-123"),
	})

	testIdentityProto6DynamicValue, err := tfprotov6.NewDynamicValue(testIdentityProto6Type, testIdentityProto6Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testIdentity := &tfsdk.ResourceIdentity{
		Raw: testIdentityProto6Value,
		Schema: identityschema.Schema{
			Attributes: map[string]identityschema.Attribute{
				"test_id": identityschema.StringAttribute{
					RequiredForImport: true,
				},
			},
		},
	}

	testIdentityInvalid := &tfsdk.ResourceIdentity{
		Raw: testIdentityProto6Value,
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
		expected *tfprotov6.UpgradeResourceIdentityResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.UpgradeResourceIdentityResponse{},
			expected: &tfprotov6.UpgradeResourceIdentityResponse{},
		},
		"diagnostics": {
			input: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov6.UpgradeResourceIdentityResponse{
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
		"diagnostics-invalid-upgradedIdentity": {
			input: &fwserver.UpgradeResourceIdentityResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
				UpgradedIdentity: testIdentityInvalid,
			},
			expected: &tfprotov6.UpgradeResourceIdentityResponse{
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
			expected: &tfprotov6.UpgradeResourceIdentityResponse{
				UpgradedIdentity: &tfprotov6.ResourceIdentityData{
					IdentityData: &testIdentityProto6DynamicValue,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.UpgradeResourceIdentityResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
