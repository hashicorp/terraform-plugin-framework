package fromproto5_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReadResourceRequest(t *testing.T) {
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

	testFwSchema := &tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test_attribute": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testCases := map[string]struct {
		input               *tfprotov5.ReadResourceRequest
		resourceSchema      fwschema.Schema
		resourceType        provider.ResourceType
		providerMetaSchema  fwschema.Schema
		expected            *fwserver.ReadResourceRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov5.ReadResourceRequest{},
			expected: &fwserver.ReadResourceRequest{},
		},
		"currentstate-missing-schema": {
			input: &tfprotov5.ReadResourceRequest{
				CurrentState: &testProto5DynamicValue,
			},
			expected: &fwserver.ReadResourceRequest{},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert State",
					"An unexpected error was encountered when converting the state from the protocol type. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"currentstate": {
			input: &tfprotov5.ReadResourceRequest{
				CurrentState: &testProto5DynamicValue,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ReadResourceRequest{
				CurrentState: &tfsdk.State{
					Raw:    testProto5Value,
					Schema: *testFwSchema,
				},
			},
		},
		"private-malformed-json": {
			input: &tfprotov5.ReadResourceRequest{
				Private: []byte(`{`),
			},
			resourceSchema: testFwSchema,
			expected:       &fwserver.ReadResourceRequest{},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when decoding private state: unexpected end of JSON input.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"private-empty-json": {
			input: &tfprotov5.ReadResourceRequest{
				Private: []byte("{}"),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ReadResourceRequest{
				Private: &privatestate.Data{
					Framework: map[string][]byte{},
					Provider:  map[string][]byte{},
				},
			},
		},
		"private": {
			input: &tfprotov5.ReadResourceRequest{
				Private: marshalToJson(map[string][]byte{
					".frameworkKey": []byte("framework value"),
					"providerKey":   []byte("provider value"),
				}),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.ReadResourceRequest{
				Private: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`framework value`),
					},
					Provider: map[string][]byte{
						"providerKey": []byte(`provider value`),
					},
				},
			},
		},
		"providermeta-missing-data": {
			input:              &tfprotov5.ReadResourceRequest{},
			providerMetaSchema: testFwSchema,
			expected: &fwserver.ReadResourceRequest{
				ProviderMeta: &tfsdk.Config{
					Raw:    tftypes.NewValue(testProto5Type, nil),
					Schema: *testFwSchema,
				},
			},
		},
		"providermeta-missing-schema": {
			input: &tfprotov5.ReadResourceRequest{
				ProviderMeta: &testProto5DynamicValue,
			},
			expected: &fwserver.ReadResourceRequest{
				// This intentionally should not include ProviderMeta
			},
		},
		"providermeta": {
			input: &tfprotov5.ReadResourceRequest{
				ProviderMeta: &testProto5DynamicValue,
			},
			providerMetaSchema: testFwSchema,
			expected: &fwserver.ReadResourceRequest{
				ProviderMeta: &tfsdk.Config{
					Raw:    testProto5Value,
					Schema: *testFwSchema,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.ReadResourceRequest(context.Background(), testCase.input, testCase.resourceType, testCase.resourceSchema, testCase.providerMetaSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func marshalToJson(input map[string][]byte) []byte {
	output, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}

	return output
}
