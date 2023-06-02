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

func TestImportResourceStateResponse(t *testing.T) {
	t.Parallel()

	testProto6Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_attribute": tftypes.String,
		},
	}

	testEmptyProto6Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}

	testProto6Value := tftypes.NewValue(testProto6Type, map[string]tftypes.Value{
		"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testEmptyProto6Value := tftypes.NewValue(testEmptyProto6Type, map[string]tftypes.Value{})

	testProto6DynamicValue, err := tfprotov6.NewDynamicValue(testProto6Type, testProto6Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testEmptyProto6DynamicValue, err := tfprotov6.NewDynamicValue(testEmptyProto6Type, testEmptyProto6Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testState := tfsdk.State{
		Raw: testProto6Value,
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"test_attribute": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	testStateInvalid := tfsdk.State{
		Raw: testProto6Value,
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"test_attribute": schema.BoolAttribute{
					Required: true,
				},
			},
		},
	}

	testEmptyState := tfsdk.State{
		Raw: testProto6Value,
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{},
		},
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testCases := map[string]struct {
		input    *fwserver.ImportResourceStateResponse
		expected *tfprotov6.ImportResourceStateResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &fwserver.ImportResourceStateResponse{},
			expected: &tfprotov6.ImportResourceStateResponse{},
		},
		"diagnostics": {
			input: &fwserver.ImportResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
			},
			expected: &tfprotov6.ImportResourceStateResponse{
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
		"diagnostics-invalid-newstate": {
			input: &fwserver.ImportResourceStateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("test warning summary", "test warning details"),
					diag.NewErrorDiagnostic("test error summary", "test error details"),
				},
				ImportedResources: []fwserver.ImportedResource{
					{
						State: testStateInvalid,
					},
				},
			},
			expected: &tfprotov6.ImportResourceStateResponse{
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
		"newstate": {
			input: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State: testState,
					},
				},
			},
			expected: &tfprotov6.ImportResourceStateResponse{
				ImportedResources: []*tfprotov6.ImportedResource{
					{
						State: &testProto6DynamicValue,
					},
				},
			},
		},
		"private": {
			input: &fwserver.ImportResourceStateResponse{
				ImportedResources: []fwserver.ImportedResource{
					{
						State: testEmptyState,
						Private: &privatestate.Data{
							Framework: map[string][]byte{
								".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
							},
							Provider: testProviderData,
						},
					},
				},
			},
			expected: &tfprotov6.ImportResourceStateResponse{
				ImportedResources: []*tfprotov6.ImportedResource{
					{
						State: &testEmptyProto6DynamicValue,
						Private: privatestate.MustMarshalToJson(map[string][]byte{
							".frameworkKey":  []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
							"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
						}),
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.ImportResourceStateResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
