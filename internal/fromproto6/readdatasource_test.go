package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestReadDataSourceRequest(t *testing.T) {
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

	testFwSchema := &tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test_attribute": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testCases := map[string]struct {
		input               *tfprotov6.ReadDataSourceRequest
		dataSourceSchema    *tfsdk.Schema
		dataSourceType      tfsdk.DataSourceType
		providerMetaSchema  *tfsdk.Schema
		expected            *fwserver.ReadDataSourceRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:    &tfprotov6.ReadDataSourceRequest{},
			expected: &fwserver.ReadDataSourceRequest{},
		},
		"config-missing-schema": {
			input: &tfprotov6.ReadDataSourceRequest{
				Config: &testProto6DynamicValue,
			},
			expected: &fwserver.ReadDataSourceRequest{},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Configuration",
					"An unexpected error was encountered when converting the configuration from the protocol type. "+
						"This is always an issue in the Terraform Provider SDK used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"config": {
			input: &tfprotov6.ReadDataSourceRequest{
				Config: &testProto6DynamicValue,
			},
			dataSourceSchema: testFwSchema,
			expected: &fwserver.ReadDataSourceRequest{
				Config: &tfsdk.Config{
					Raw:    testProto6Value,
					Schema: *testFwSchema,
				},
			},
		},
		"providermeta-missing-data": {
			input:              &tfprotov6.ReadDataSourceRequest{},
			providerMetaSchema: testFwSchema,
			expected: &fwserver.ReadDataSourceRequest{
				ProviderMeta: &tfsdk.Config{
					Raw:    tftypes.NewValue(testProto6Type, nil),
					Schema: *testFwSchema,
				},
			},
		},
		"providermeta-missing-schema": {
			input: &tfprotov6.ReadDataSourceRequest{
				ProviderMeta: &testProto6DynamicValue,
			},
			expected: &fwserver.ReadDataSourceRequest{
				// This intentionally should not include ProviderMeta
			},
		},
		"providermeta": {
			input: &tfprotov6.ReadDataSourceRequest{
				ProviderMeta: &testProto6DynamicValue,
			},
			providerMetaSchema: testFwSchema,
			expected: &fwserver.ReadDataSourceRequest{
				ProviderMeta: &tfsdk.Config{
					Raw:    testProto6Value,
					Schema: *testFwSchema,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.ReadDataSourceRequest(context.Background(), testCase.input, testCase.dataSourceType, testCase.dataSourceSchema, testCase.providerMetaSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
