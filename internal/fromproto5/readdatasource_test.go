// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestReadDataSourceRequest(t *testing.T) {
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

	testCases := map[string]struct {
		input               *tfprotov5.ReadDataSourceRequest
		dataSourceSchema    fwschema.Schema
		dataSource          datasource.DataSource
		providerMetaSchema  fwschema.Schema
		expected            *fwserver.ReadDataSourceRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov5.ReadDataSourceRequest{},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing DataSource Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config-missing-schema": {
			input: &tfprotov5.ReadDataSourceRequest{
				Config: &testProto5DynamicValue,
			},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing DataSource Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config": {
			input: &tfprotov5.ReadDataSourceRequest{
				Config: &testProto5DynamicValue,
			},
			dataSourceSchema: testFwSchema,
			expected: &fwserver.ReadDataSourceRequest{
				Config: &tfsdk.Config{
					Raw:    testProto5Value,
					Schema: testFwSchema,
				},
				DataSourceSchema: testFwSchema,
			},
		},
		"providermeta-missing-data": {
			input:              &tfprotov5.ReadDataSourceRequest{},
			dataSourceSchema:   testFwSchema,
			providerMetaSchema: testFwSchema,
			expected: &fwserver.ReadDataSourceRequest{
				DataSourceSchema: testFwSchema,
				ProviderMeta: &tfsdk.Config{
					Raw:    tftypes.NewValue(testProto5Type, nil),
					Schema: testFwSchema,
				},
			},
		},
		"providermeta-missing-schema": {
			input: &tfprotov5.ReadDataSourceRequest{
				ProviderMeta: &testProto5DynamicValue,
			},
			dataSourceSchema: testFwSchema,
			expected: &fwserver.ReadDataSourceRequest{
				DataSourceSchema: testFwSchema,
				// This intentionally should not include ProviderMeta
			},
		},
		"providermeta": {
			input: &tfprotov5.ReadDataSourceRequest{
				ProviderMeta: &testProto5DynamicValue,
			},
			dataSourceSchema:   testFwSchema,
			providerMetaSchema: testFwSchema,
			expected: &fwserver.ReadDataSourceRequest{
				DataSourceSchema: testFwSchema,
				ProviderMeta: &tfsdk.Config{
					Raw:    testProto5Value,
					Schema: testFwSchema,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.ReadDataSourceRequest(context.Background(), testCase.input, testCase.dataSource, testCase.dataSourceSchema, testCase.providerMetaSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
