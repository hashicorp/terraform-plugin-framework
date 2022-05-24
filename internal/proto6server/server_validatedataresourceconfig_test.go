package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerValidateDataResourceConfig(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		config         tftypes.Value
		dataSource     string
		dataSourceType tftypes.Type

		impl func(context.Context, tfsdk.ValidateDataSourceConfigRequest, *tfsdk.ValidateDataSourceConfigResponse)

		// response expectations
		expectedDiags []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"no_validation": {
			config: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, nil),
				"current_time": tftypes.NewValue(tftypes.String, nil),
				"is_dst":       tftypes.NewValue(tftypes.Bool, nil),
			}),
			dataSource:     "test_one",
			dataSourceType: testServeDataSourceTypeOneType,
		},
		"config_validators_no_diags": {
			config: tftypes.NewValue(testServeDataSourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_config_validators",
			dataSourceType: testServeDataSourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
			},
		},
		"config_validators_one_diag": {
			config: tftypes.NewValue(testServeDataSourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_config_validators",
			dataSourceType: testServeDataSourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
				if len(resp.Diagnostics) == 0 {
					resp.Diagnostics.AddError(
						"This is an error",
						"Oops.",
					)
				} else {
					resp.Diagnostics.AddError(
						"This is another error",
						"Oops again.",
					)
				}
			},

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Summary:  "This is an error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops.",
				},
				// ConfigValidators includes multiple calls
				{
					Summary:  "This is another error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops again.",
				},
			},
		},
		"config_validators_two_diags": {
			config: tftypes.NewValue(testServeDataSourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_config_validators",
			dataSourceType: testServeDataSourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
				if len(resp.Diagnostics) == 0 {
					resp.Diagnostics.AddAttributeWarning(
						tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
						"This is a warning",
						"This is your final warning",
					)
					resp.Diagnostics.AddError(
						"This is an error",
						"Oops.",
					)
				} else {
					resp.Diagnostics.AddAttributeWarning(
						tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
						"This is another warning",
						"This is really your final warning",
					)
					resp.Diagnostics.AddError(
						"This is another error",
						"Oops again.",
					)
				}
			},

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Summary:   "This is a warning",
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Detail:    "This is your final warning",
					Attribute: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
				},
				{
					Summary:  "This is an error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops.",
				},
				// ConfigValidators includes multiple calls
				{
					Summary:   "This is another warning",
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Detail:    "This is really your final warning",
					Attribute: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
				},
				{
					Summary:  "This is another error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops again.",
				},
			},
		},
		"validate_config_no_diags": {
			config: tftypes.NewValue(testServeDataSourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_validate_config",
			dataSourceType: testServeDataSourceTypeValidateConfigType,

			impl: func(_ context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
			},
		},
		"validate_config_one_diag": {
			config: tftypes.NewValue(testServeDataSourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_validate_config",
			dataSourceType: testServeDataSourceTypeValidateConfigType,

			impl: func(_ context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
				resp.Diagnostics.AddError(
					"This is an error",
					"Oops.",
				)
			},

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Summary:  "This is an error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops.",
				},
			},
		},
		"validate_config_two_diags": {
			config: tftypes.NewValue(testServeDataSourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_validate_config",
			dataSourceType: testServeDataSourceTypeValidateConfigType,

			impl: func(_ context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
				resp.Diagnostics.AddAttributeWarning(
					tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
					"This is a warning",
					"This is your final warning",
				)
				resp.Diagnostics.AddError(
					"This is an error",
					"Oops.",
				)
			},

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Summary:   "This is a warning",
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Detail:    "This is your final warning",
					Attribute: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
				},
				{
					Summary:  "This is an error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops.",
				},
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := &testServeProvider{
				validateDataSourceConfigImpl: tc.impl,
			}
			testServer := &Server{
				FrameworkServer: fwserver.Server{
					Provider: s,
				},
			}

			dv, err := tfprotov6.NewDynamicValue(tc.dataSourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ValidateDataResourceConfigRequest{
				TypeName: tc.dataSource,
				Config:   &dv,
			}
			got, err := testServer.ValidateDataResourceConfig(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.validateDataSourceConfigCalledDataSourceType != tc.dataSource && !(tc.dataSource == "test_one" && s.validateDataSourceConfigCalledDataSourceType == "") {
				t.Errorf("Called wrong data source. Expected to call %q, actually called %q", tc.dataSource, s.readDataSourceCalledDataSourceType)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}
