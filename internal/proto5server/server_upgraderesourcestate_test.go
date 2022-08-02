package proto5server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerUpgradeResourceState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"optional_attribute": {
				Type:     types.StringType,
				Optional: true,
			},
			"required_attribute": {
				Type:     types.StringType,
				Required: true,
			},
		},
		Version: 1, // Must be above 0
	}
	schemaType := schema.TerraformType(ctx)

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.UpgradeResourceStateRequest
		expectedResponse *tfprotov5.UpgradeResourceStateResponse
		expectedError    error
	}{
		"nil": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request:          nil,
			expectedResponse: &tfprotov5.UpgradeResourceStateResponse{},
		},
		"request-RawState": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
							return map[string]provider.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return schema, nil
									},
									NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
										return &testprovider.ResourceWithUpgradeState{
											Resource: &testprovider.Resource{},
											UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
												return map[int64]resource.StateUpgrader{
													0: {
														StateUpgrader: func(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
															expectedRawState := testNewTfprotov6RawState(t, map[string]interface{}{
																"id":                 "test-id-value",
																"required_attribute": true,
															})

															if diff := cmp.Diff(req.RawState, expectedRawState); diff != "" {
																resp.Diagnostics.AddError("Unexpected req.RawState difference", diff)
															}

															// Prevent Missing Upgraded Resource State error
															resp.State = tfsdk.State{
																Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
																	"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
																	"optional_attribute": tftypes.NewValue(tftypes.String, nil),
																	"required_attribute": tftypes.NewValue(tftypes.String, "true"),
																}),
																Schema: schema,
															}
														},
													},
												}
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov5.UpgradeResourceStateRequest{
				RawState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_resource",
				Version:  0,
			},
			expectedResponse: &tfprotov5.UpgradeResourceStateResponse{
				UpgradedState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"request-TypeName-missing": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request: &tfprotov5.UpgradeResourceStateRequest{},
			expectedResponse: &tfprotov5.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Resource Type Not Found",
						Detail:   "No resource type named \"\" was found in the provider.",
					},
				},
			},
		},
		"request-TypeName-unknown": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{},
				},
			},
			request: &tfprotov5.UpgradeResourceStateRequest{
				TypeName: "unknown",
			},
			expectedResponse: &tfprotov5.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Resource Type Not Found",
						Detail:   "No resource type named \"unknown\" was found in the provider.",
					},
				},
			},
		},
		"response-Diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
							return map[string]provider.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return schema, nil
									},
									NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
										return &testprovider.ResourceWithUpgradeState{
											Resource: &testprovider.Resource{},
											UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
												return map[int64]resource.StateUpgrader{
													0: {
														StateUpgrader: func(_ context.Context, _ resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
															resp.Diagnostics.AddWarning("warning summary", "warning detail")
															resp.Diagnostics.AddError("error summary", "error detail")
														},
													},
												}
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov5.UpgradeResourceStateRequest{
				RawState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_resource",
				Version:  0,
			},
			expectedResponse: &tfprotov5.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
				},
			},
		},
		"response-UpgradedState": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
							return map[string]provider.ResourceType{
								"test_resource": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return schema, nil
									},
									NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
										return &testprovider.ResourceWithUpgradeState{
											Resource: &testprovider.Resource{},
											UpgradeStateMethod: func(ctx context.Context) map[int64]resource.StateUpgrader {
												return map[int64]resource.StateUpgrader{
													0: {
														StateUpgrader: func(_ context.Context, _ resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
															resp.State = tfsdk.State{
																Raw: tftypes.NewValue(schemaType, map[string]tftypes.Value{
																	"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
																	"optional_attribute": tftypes.NewValue(tftypes.String, nil),
																	"required_attribute": tftypes.NewValue(tftypes.String, "true"),
																}),
																Schema: schema,
															}
														},
													},
												}
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov5.UpgradeResourceStateRequest{
				RawState: testNewTfprotov5RawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_resource",
				Version:  0,
			},
			expectedResponse: &tfprotov5.UpgradeResourceStateResponse{
				UpgradedState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.UpgradeResourceState(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
