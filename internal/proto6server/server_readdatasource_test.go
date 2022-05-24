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

func TestServerReadDataSource(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		config         tftypes.Value
		providerMeta   tftypes.Value
		dataSource     string
		dataSourceType tftypes.Type

		impl func(context.Context, tfsdk.ReadDataSourceRequest, *tfsdk.ReadDataSourceResponse)

		// response expectations
		expectedNewState tftypes.Value
		expectedDiags    []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"one_basic": {
			config: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, nil),
				"current_time": tftypes.NewValue(tftypes.String, nil),
				"is_dst":       tftypes.NewValue(tftypes.Bool, nil),
			}),
			dataSource:     "test_one",
			dataSourceType: testServeDataSourceTypeOneType,

			impl: func(_ context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
					"current_date": tftypes.NewValue(tftypes.String, "today"),
					"current_time": tftypes.NewValue(tftypes.String, "now"),
					"is_dst":       tftypes.NewValue(tftypes.Bool, true),
				})
			},

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, "today"),
				"current_time": tftypes.NewValue(tftypes.String, "now"),
				"is_dst":       tftypes.NewValue(tftypes.Bool, true),
			}),
		},
		"one_provider_meta": {
			config: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, nil),
				"current_time": tftypes.NewValue(tftypes.String, nil),
				"is_dst":       tftypes.NewValue(tftypes.Bool, nil),
			}),
			dataSource:     "test_one",
			dataSourceType: testServeDataSourceTypeOneType,

			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),

			impl: func(_ context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
					"current_date": tftypes.NewValue(tftypes.String, "today"),
					"current_time": tftypes.NewValue(tftypes.String, "now"),
					"is_dst":       tftypes.NewValue(tftypes.Bool, true),
				})
			},

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, "today"),
				"current_time": tftypes.NewValue(tftypes.String, "now"),
				"is_dst":       tftypes.NewValue(tftypes.Bool, true),
			}),
		},
		"one_remove": {
			config: tftypes.NewValue(testServeDataSourceTypeOneType, map[string]tftypes.Value{
				"current_date": tftypes.NewValue(tftypes.String, nil),
				"current_time": tftypes.NewValue(tftypes.String, nil),
				"is_dst":       tftypes.NewValue(tftypes.Bool, nil),
			}),
			dataSource:     "test_one",
			dataSourceType: testServeDataSourceTypeOneType,

			impl: func(_ context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeOneType, nil)
			},

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeOneType, nil),
		},
		"two_basic": {
			config: tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
				"family": tftypes.NewValue(tftypes.String, "123foo"),
				"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
				"id":     tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_two",
			dataSourceType: testServeDataSourceTypeTwoType,

			impl: func(_ context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
					"family": tftypes.NewValue(tftypes.String, "123foo"),
					"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
					"id":     tftypes.NewValue(tftypes.String, "a random id or something I dunno"),
				})
			},

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
				"family": tftypes.NewValue(tftypes.String, "123foo"),
				"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
				"id":     tftypes.NewValue(tftypes.String, "a random id or something I dunno"),
			}),
		},
		"two_diags": {
			config: tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
				"family": tftypes.NewValue(tftypes.String, "123foo"),
				"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
				"id":     tftypes.NewValue(tftypes.String, nil),
			}),
			dataSource:     "test_two",
			dataSourceType: testServeDataSourceTypeTwoType,

			impl: func(_ context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
					"family": tftypes.NewValue(tftypes.String, "123foo"),
					"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
					"id":     tftypes.NewValue(tftypes.String, "a random id or something I dunno"),
				})
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

			expectedNewState: tftypes.NewValue(testServeDataSourceTypeTwoType, map[string]tftypes.Value{
				"family": tftypes.NewValue(tftypes.String, "123foo"),
				"name":   tftypes.NewValue(tftypes.String, "123foo-askjgsio"),
				"id":     tftypes.NewValue(tftypes.String, "a random id or something I dunno"),
			}),

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
				readDataSourceImpl: tc.impl,
			}
			testServer := &Server{
				FrameworkServer: fwserver.Server{
					Provider: s,
				},
			}
			var pmSchema tfsdk.Schema
			if tc.providerMeta.Type() != nil {
				testServer.FrameworkServer.Provider = &testServeProviderWithMetaSchema{s}
				schema, diags := testServer.FrameworkServer.ProviderMetaSchema(context.Background())
				if len(diags) > 0 {
					t.Errorf("Unexpected diags: %+v", diags)
					return
				}
				pmSchema = *schema
			}

			rt, diags := testServer.FrameworkServer.DataSourceType(context.Background(), tc.dataSource)
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}
			schema, diags := rt.GetSchema(context.Background())
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}

			dv, err := tfprotov6.NewDynamicValue(tc.dataSourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ReadDataSourceRequest{
				TypeName: tc.dataSource,
				Config:   &dv,
			}
			if tc.providerMeta.Type() != nil {
				providerMeta, err := tfprotov6.NewDynamicValue(testServeProviderMetaType, tc.providerMeta)
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				req.ProviderMeta = &providerMeta
			}
			got, err := testServer.ReadDataSource(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.readDataSourceCalledDataSourceType != tc.dataSource {
				t.Errorf("Called wrong dataSource. Expected to call %q, actually called %q", tc.dataSource, s.readDataSourceCalledDataSourceType)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			if diff := cmp.Diff(s.readDataSourceConfigValue, tc.config); diff != "" {
				t.Errorf("Unexpected diff in config (+wanted, -got): %s", diff)
				return
			}
			if diff := cmp.Diff(s.readDataSourceConfigSchema, schema); diff != "" {
				t.Errorf("Unexpected diff in config schema (+wanted, -got): %s", diff)
				return
			}
			if tc.providerMeta.Type() != nil {
				if diff := cmp.Diff(s.readDataSourceProviderMetaValue, tc.providerMeta); diff != "" {
					t.Errorf("Unexpected diff in provider meta (+wanted, -got): %s", diff)
					return
				}
				if diff := cmp.Diff(s.readDataSourceProviderMetaSchema, pmSchema); diff != "" {
					t.Errorf("Unexpected diff in provider meta schema (+wanted, -got): %s", diff)
					return
				}
			}
			gotNewState, err := got.State.Unmarshal(tc.dataSourceType)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(gotNewState, tc.expectedNewState); diff != "" {
				t.Errorf("Unexpected diff in new state (+wanted, -got): %s", diff)
				return
			}
		})
	}
}
