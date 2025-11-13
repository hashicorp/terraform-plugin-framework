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
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-framework/statestore/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestConfigureStateStoreRequest(t *testing.T) {
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

	testCases := map[string]struct {
		input               *tfprotov6.ConfigureStateStoreRequest
		statestoreSchema    fwschema.Schema
		expected            *statestore.ConfigureStateStoreRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov6.ConfigureStateStoreRequest{},
			expected: &statestore.ConfigureStateStoreRequest{},
		},
		"config-missing-schema": {
			input: &tfprotov6.ConfigureStateStoreRequest{
				Config: &testProto6DynamicValue,
			},
			expected: &statestore.ConfigureStateStoreRequest{},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Configuration",
					"An unexpected error was encountered when converting the configuration from the protocol type. "+
						"This is always an issue in terraform-plugin-framework used to implement the statestore and should be reported to the statestore developers.\n\n"+
						"Please report this to the statestore developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config": {
			input: &tfprotov6.ConfigureStateStoreRequest{
				Config: &testProto6DynamicValue,
			},
			statestoreSchema: testFwSchema,
			expected: &statestore.ConfigureStateStoreRequest{
				Config: tfsdk.Config{
					Raw:    testProto6Value,
					Schema: testFwSchema,
				},
			},
		},
		"typename": {
			input: &tfprotov6.ConfigureStateStoreRequest{
				TypeName: "foo",
			},
			expected: &statestore.ConfigureStateStoreRequest{
				TypeName: "foo",
			},
		},
		"client-capabilities": {
			input: &tfprotov6.ConfigureStateStoreRequest{
				Capabilities: &tfprotov6.ConfigureStateStoreClientCapabilities{
					DeferralAllowed: true,
				},
			},
			expected: &statestore.ConfigureStateStoreRequest{
				Capabilities: statestore.ConfigureStateStoreClientCapabilities{
					DeferralAllowed: true,
				},
			},
		},
		"client-capabilities-unset": {
			input: &tfprotov6.ConfigureStateStoreRequest{},
			expected: &statestore.ConfigureStateStoreRequest{
				Capabilities: statestore.ConfigureStateStoreClientCapabilities{
					DeferralAllowed: false,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.ConfigureStateStoreRequest(context.Background(), testCase.input, testCase.statestoreSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
