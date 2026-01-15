// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestOpenEphemeralResourceResponse(t *testing.T) {
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

	testDeferral := &ephemeral.Deferred{
		Reason: ephemeral.DeferredReasonAbsentPrereq,
	}

	testProto5Deferred := &tfprotov5.Deferred{
		Reason: tfprotov5.DeferredReasonAbsentPrereq,
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testEphemeralResult := &tfsdk.EphemeralResultData{
		Raw: testProto5Value,
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"test_attribute": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	testEphemeralResultInvalid := &tfsdk.EphemeralResultData{
		Raw: testProto5Value,
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"test_attribute": schema.BoolAttribute{
					Required: true,
				},
			},
		},
	}

	testCases := map[string]struct {
		input    *fwserver.OpenEphemeralResourceResponse
		expected *tfprotov5.OpenEphemeralResourceResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input: &fwserver.OpenEphemeralResourceResponse{},
			expected: &tfprotov5.OpenEphemeralResourceResponse{
				// Time zero
				RenewAt: *new(time.Time),
			},
		},
		"diagnostics": {
			input: &fwserver.OpenEphemeralResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov5.OpenEphemeralResourceResponse{
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
		"diagnostics-invalid-result": {
			input: &fwserver.OpenEphemeralResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
				Result: testEphemeralResultInvalid,
			},
			expected: &tfprotov5.OpenEphemeralResourceResponse{
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
						Summary:  "Unable to Convert Ephemeral Result Data",
						Detail: "An unexpected error was encountered when converting the ephemeral result data to the protocol type. " +
							"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n" +
							"Please report this to the provider developer:\n\n" +
							"Unable to create DynamicValue: AttributeName(\"test_attribute\"): unexpected value type string, tftypes.Bool values must be of type bool",
					},
				},
			},
		},
		"renew-at": {
			input: &fwserver.OpenEphemeralResourceResponse{
				RenewAt: time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC),
			},
			expected: &tfprotov5.OpenEphemeralResourceResponse{
				RenewAt: time.Date(2024, 8, 29, 5, 10, 32, 0, time.UTC),
			},
		},
		"result": {
			input: &fwserver.OpenEphemeralResourceResponse{
				Result: testEphemeralResult,
			},
			expected: &tfprotov5.OpenEphemeralResourceResponse{
				Result: &testProto5DynamicValue,
			},
		},
		"private-empty": {
			input: &fwserver.OpenEphemeralResourceResponse{
				Private: &privatestate.Data{
					Framework: map[string][]byte{},
					Provider:  testEmptyProviderData,
				},
			},
			expected: &tfprotov5.OpenEphemeralResourceResponse{
				Private: nil,
			},
		},
		"private": {
			input: &fwserver.OpenEphemeralResourceResponse{
				Private: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`)},
					Provider: testProviderData,
				},
			},
			expected: &tfprotov5.OpenEphemeralResourceResponse{
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey":  []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
		},
		"deferral": {
			input: &fwserver.OpenEphemeralResourceResponse{
				Deferred: testDeferral,
			},
			expected: &tfprotov5.OpenEphemeralResourceResponse{
				Deferred: testProto5Deferred,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.OpenEphemeralResourceResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
