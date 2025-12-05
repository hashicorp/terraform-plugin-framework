// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestApplyResourceChangeRequest(t *testing.T) {
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

	testFwSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_attribute": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testIdentityProto5Type := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_identity_attribute": tftypes.String,
		},
	}

	testIdentityProto5Value := tftypes.NewValue(testIdentityProto5Type, map[string]tftypes.Value{
		"test_identity_attribute": tftypes.NewValue(tftypes.String, "id-123"),
	})

	testIdentityProto5DynamicValue, err := tfprotov5.NewDynamicValue(testIdentityProto5Type, testIdentityProto5Value)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov5.NewDynamicValue(): %s", err)
	}

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_identity_attribute": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testCases := map[string]struct {
		input               *tfprotov5.ApplyResourceChangeRequest
		resourceBehavior    resource.ResourceBehavior
		resourceSchema      fwschema.Schema
		resource            resource.Resource
		providerMetaSchema  fwschema.Schema
		identitySchema      fwschema.Schema
		expected            *fwserver.ApplyResourceChangeRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov5.ApplyResourceChangeRequest{},
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
			input: &tfprotov5.ApplyResourceChangeRequest{
				Config: &testProto5DynamicValue,
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
			input: &tfprotov5.ApplyResourceChangeRequest{
				Config: &testProto5DynamicValue,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				Config: &tfsdk.Config{
					Raw:    testProto5Value,
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"plannedstate-missing-schema": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				PlannedState: &testProto5DynamicValue,
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
		"plannedstate": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				PlannedState: &testProto5DynamicValue,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				PlannedState: &tfsdk.Plan{
					Raw:    testProto5Value,
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"plannedidentity-missing-schema": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				PlannedIdentity: &tfprotov5.ResourceIdentityData{
					IdentityData: &testIdentityProto5DynamicValue,
				},
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				ResourceSchema: testFwSchema,
			},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Resource Identity",
					"An unexpected error was encountered when converting the resource identity from the protocol type. "+
						"Identity data was sent in the protocol to a resource that doesn't support identity.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"plannedidentity": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				PlannedIdentity: &tfprotov5.ResourceIdentityData{
					IdentityData: &testIdentityProto5DynamicValue,
				},
			},
			identitySchema: testIdentitySchema,
			resourceSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				IdentitySchema: testIdentitySchema,
				PlannedIdentity: &tfsdk.ResourceIdentity{
					Raw:    testIdentityProto5Value,
					Schema: testIdentitySchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"plannedprivate-malformed-json": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				PlannedPrivate: []byte(`{`),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				ResourceSchema: testFwSchema,
			}, expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when decoding private state: unexpected end of JSON input.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"plannedprivate-empty-json": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				PlannedPrivate: []byte("{}"),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				ResourceSchema: testFwSchema,
				PlannedPrivate: &privatestate.Data{
					Framework: map[string][]byte{},
					Provider:  testEmptyProviderData,
				},
			},
		},
		"plannedprivate": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				PlannedPrivate: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey":  []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				ResourceSchema: testFwSchema,
				PlannedPrivate: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					},
					Provider: testProviderData,
				},
			},
		},
		"priorstate-missing-schema": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				PriorState: &testProto5DynamicValue,
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
			input: &tfprotov5.ApplyResourceChangeRequest{
				PriorState: &testProto5DynamicValue,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				PriorState: &tfsdk.State{
					Raw:    testProto5Value,
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"providermeta-missing-data": {
			input:              &tfprotov5.ApplyResourceChangeRequest{},
			resourceSchema:     testFwSchema,
			providerMetaSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				ProviderMeta: &tfsdk.Config{
					Raw:    tftypes.NewValue(testProto5Type, nil),
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"providermeta-missing-schema": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				ProviderMeta: &testProto5DynamicValue,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				// This intentionally should not include ProviderMeta
				ResourceSchema: testFwSchema,
			},
		},
		"providermeta": {
			input: &tfprotov5.ApplyResourceChangeRequest{
				ProviderMeta: &testProto5DynamicValue,
			},
			resourceSchema:     testFwSchema,
			providerMetaSchema: testFwSchema,
			expected: &fwserver.ApplyResourceChangeRequest{
				ProviderMeta: &tfsdk.Config{
					Raw:    testProto5Value,
					Schema: testFwSchema,
				},
				ResourceSchema: testFwSchema,
			},
		},
		"resource-behavior": {
			input:          &tfprotov5.ApplyResourceChangeRequest{},
			resourceSchema: testFwSchema,
			resourceBehavior: resource.ResourceBehavior{
				MutableIdentity: true,
			},
			expected: &fwserver.ApplyResourceChangeRequest{
				ResourceBehavior: resource.ResourceBehavior{
					MutableIdentity: true,
				},
				ResourceSchema: testFwSchema,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.ApplyResourceChangeRequest(context.Background(), testCase.input, testCase.resource, testCase.resourceSchema, testCase.providerMetaSchema, testCase.resourceBehavior, testCase.identitySchema)

			if diff := cmp.Diff(got, testCase.expected, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
