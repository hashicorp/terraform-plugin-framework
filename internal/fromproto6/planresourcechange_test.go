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
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestPlanResourceChangeRequest(t *testing.T) {
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
		input               *tfprotov6.PlanResourceChangeRequest
		resourceSchema      fwschema.Schema
		resource            resource.Resource
		providerMetaSchema  fwschema.Schema
		expected            *fwserver.PlanResourceChangeRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov6.PlanResourceChangeRequest{},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing Resource Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config-missing-schema": {
			input: &tfprotov6.PlanResourceChangeRequest{
				Config: &testProto6DynamicValue,
			},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing Resource Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config": {
			input: &tfprotov6.PlanResourceChangeRequest{
				Config: &testProto6DynamicValue,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.PlanResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw:    testProto6Value,
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"priorprivate": {
			input: &tfprotov6.PlanResourceChangeRequest{
				PriorPrivate: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey":  []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.PlanResourceChangeRequest{
				PriorPrivate: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					},
					Provider: testProviderData,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"priorstate-missing-schema": {
			input: &tfprotov6.PlanResourceChangeRequest{
				PriorState: &testProto6DynamicValue,
			},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing Resource Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"priorstate": {
			input: &tfprotov6.PlanResourceChangeRequest{
				PriorState: &testProto6DynamicValue,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.PlanResourceChangeRequest{
				PriorState: &tfsdk.State{
					Raw:    testProto6Value,
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"proposednewstate-missing-schema": {
			input: &tfprotov6.PlanResourceChangeRequest{
				ProposedNewState: &testProto6DynamicValue,
			},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing Resource Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"proposednewstate": {
			input: &tfprotov6.PlanResourceChangeRequest{
				ProposedNewState: &testProto6DynamicValue,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.PlanResourceChangeRequest{
				ProposedNewState: &tfsdk.Plan{
					Raw:    testProto6Value,
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"providermeta-missing-data": {
			input:              &tfprotov6.PlanResourceChangeRequest{},
			resourceSchema:     testFwSchema,
			providerMetaSchema: testFwSchema,
			expected: &fwserver.PlanResourceChangeRequest{
				ProviderMeta: &tfsdk.Config{
					Raw:    tftypes.NewValue(testProto6Type, nil),
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"providermeta-missing-schema": {
			input: &tfprotov6.PlanResourceChangeRequest{
				ProviderMeta: &testProto6DynamicValue,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.PlanResourceChangeRequest{
				// This intentionally should not include ProviderMeta
				ResourceSchema: testFwSchema,
			},
		},
		"providermeta": {
			input: &tfprotov6.PlanResourceChangeRequest{
				ProviderMeta: &testProto6DynamicValue,
			},
			resourceSchema:     testFwSchema,
			providerMetaSchema: testFwSchema,
			expected: &fwserver.PlanResourceChangeRequest{
				ProviderMeta: &tfsdk.Config{
					Raw:    testProto6Value,
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.PlanResourceChangeRequest(context.Background(), testCase.input, testCase.resource, testCase.resourceSchema, testCase.providerMetaSchema)

			if diff := cmp.Diff(got, testCase.expected, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
