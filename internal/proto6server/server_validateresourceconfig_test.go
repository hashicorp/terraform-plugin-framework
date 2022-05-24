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

func TestServerValidateResourceConfig(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		config       tftypes.Value
		resource     string
		resourceType tftypes.Type

		impl func(context.Context, tfsdk.ValidateResourceConfigRequest, *tfsdk.ValidateResourceConfigResponse)

		// response expectations
		expectedDiags []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"no_validation": {
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, ""),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, ""),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,
		},
		"config_validators_no_diags": {
			config: tftypes.NewValue(testServeResourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_config_validators",
			resourceType: testServeResourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
			},
		},
		"config_validators_one_diag": {
			config: tftypes.NewValue(testServeResourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_config_validators",
			resourceType: testServeResourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
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
			config: tftypes.NewValue(testServeResourceTypeConfigValidatorsType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_config_validators",
			resourceType: testServeResourceTypeConfigValidatorsType,

			impl: func(_ context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
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
			config: tftypes.NewValue(testServeResourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_validate_config",
			resourceType: testServeResourceTypeValidateConfigType,

			impl: func(_ context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
			},
		},
		"validate_config_one_diag": {
			config: tftypes.NewValue(testServeResourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_validate_config",
			resourceType: testServeResourceTypeValidateConfigType,

			impl: func(_ context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
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
			config: tftypes.NewValue(testServeResourceTypeValidateConfigType, map[string]tftypes.Value{
				"string": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_validate_config",
			resourceType: testServeResourceTypeValidateConfigType,

			impl: func(_ context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
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
				validateResourceConfigImpl: tc.impl,
			}
			testServer := &Server{
				FrameworkServer: fwserver.Server{
					Provider: s,
				},
			}

			dv, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ValidateResourceConfigRequest{
				TypeName: tc.resource,
				Config:   &dv,
			}
			got, err := testServer.ValidateResourceConfig(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.validateResourceConfigCalledResourceType != tc.resource && !(tc.resource == "test_one" && s.validateResourceConfigCalledResourceType == "") {
				t.Errorf("Called wrong resource. Expected to call %q, actually called %q", tc.resource, s.readDataSourceCalledDataSourceType)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}
