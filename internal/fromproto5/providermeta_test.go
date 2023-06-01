// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestProviderMeta(t *testing.T) {
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

	testFwSchema := metaschema.Schema{
		Attributes: map[string]metaschema.Attribute{
			"test_attribute": metaschema.StringAttribute{
				Required: true,
			},
		},
	}

	testFwSchemaInvalid := metaschema.Schema{
		Attributes: map[string]metaschema.Attribute{
			"test_attribute": metaschema.BoolAttribute{
				Required: true,
			},
		},
	}

	testCases := map[string]struct {
		input               *tfprotov5.DynamicValue
		schema              fwschema.Schema
		expected            *tfsdk.Config
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"missing-schema": {
			input:    &testProto5DynamicValue,
			expected: nil,
		},
		"invalid-schema": {
			input:    &testProto5DynamicValue,
			schema:   testFwSchemaInvalid,
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Provider Meta Configuration",
					"An unexpected error was encountered when converting the provider meta configuration from the protocol type. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"AttributeName(\"test_attribute\"): couldn't decode bool: msgpack: invalid code=aa decoding bool",
				),
			},
		},
		"schema-and-data": {
			input:  &testProto5DynamicValue,
			schema: testFwSchema,
			expected: &tfsdk.Config{
				Raw:    testProto5Value,
				Schema: testFwSchema,
			},
		},
		"schema-no-data": {
			input:  nil,
			schema: testFwSchema,
			expected: &tfsdk.Config{
				Raw:    tftypes.NewValue(testProto5Type, nil),
				Schema: testFwSchema,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.ProviderMeta(context.Background(), testCase.input, testCase.schema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
