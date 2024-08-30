// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestRenewEphemeralResourceRequest(t *testing.T) {
	t.Parallel()

	testProto6Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_attribute": tftypes.String,
		},
	}

	testProto6Value := tftypes.NewValue(testProto6Type, map[string]tftypes.Value{
		"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testProto6DynamicValue, err := tfprotov6.NewDynamicValue(testProto6Type, testProto6Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
	}

	testFwSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_attribute": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testCases := map[string]struct {
		input                   *tfprotov6.RenewEphemeralResourceRequest
		ephemeralResourceSchema fwschema.Schema
		ephemeralResource       ephemeral.EphemeralResource
		providerMetaSchema      fwschema.Schema
		expected                *fwserver.RenewEphemeralResourceRequest
		expectedDiagnostics     diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov6.RenewEphemeralResourceRequest{},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing EphemeralResource Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config-missing-schema": {
			input: &tfprotov6.RenewEphemeralResourceRequest{
				Config: &testProto6DynamicValue,
			},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing EphemeralResource Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config": {
			input: &tfprotov6.RenewEphemeralResourceRequest{
				Config: &testProto6DynamicValue,
			},
			ephemeralResourceSchema: testFwSchema,
			expected: &fwserver.RenewEphemeralResourceRequest{
				Config: &tfsdk.Config{
					Raw:    testProto6Value,
					Schema: testFwSchema,
				},
				EphemeralResourceSchema: testFwSchema,
			},
		},
		"private": {
			input: &tfprotov6.RenewEphemeralResourceRequest{
				Private: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey":  []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
			ephemeralResourceSchema: testFwSchema,
			expected: &fwserver.RenewEphemeralResourceRequest{
				Private: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					},
					Provider: testProviderData,
				},
				EphemeralResourceSchema: testFwSchema,
			},
		},
		"state-missing-schema": {
			input: &tfprotov6.RenewEphemeralResourceRequest{
				State: &testProto6DynamicValue,
			},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing EphemeralResource Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"state": {
			input: &tfprotov6.RenewEphemeralResourceRequest{
				State: &testProto6DynamicValue,
			},
			ephemeralResourceSchema: testFwSchema,
			expected: &fwserver.RenewEphemeralResourceRequest{
				State: &tfsdk.EphemeralState{
					Raw:    testProto6Value,
					Schema: testFwSchema,
				},
				EphemeralResourceSchema: testFwSchema,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.RenewEphemeralResourceRequest(context.Background(), testCase.input, testCase.ephemeralResource, testCase.ephemeralResourceSchema)

			if diff := cmp.Diff(got, testCase.expected, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
