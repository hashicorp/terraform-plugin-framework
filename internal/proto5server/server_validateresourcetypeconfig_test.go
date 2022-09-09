package proto5server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerValidateResourceTypeConfig(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test": tftypes.String,
		},
	}

	testValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testDynamicValue, err := tfprotov5.NewDynamicValue(testType, testValue)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov5.NewDynamicValue(): %s", err)
	}

	testSchema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.ValidateResourceTypeConfigRequest
		expectedError    error
		expectedResponse *tfprotov5.ValidateResourceTypeConfigResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return tfsdk.Schema{}, nil
										},
										TypeNameMethod: func(_ context.Context, _ resource.TypeNameRequest, resp *resource.TypeNameResponse) {
											resp.TypeName = "test_resource"
										},
										Resource: &testprovider.Resource{},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ValidateResourceTypeConfigRequest{
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ValidateResourceTypeConfigResponse{},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithGetSchemaAndTypeName{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return testSchema, nil
										},
										TypeNameMethod: func(_ context.Context, _ resource.TypeNameRequest, resp *resource.TypeNameResponse) {
											resp.TypeName = "test_resource"
										},
										Resource: &testprovider.Resource{},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ValidateResourceTypeConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ValidateResourceTypeConfigResponse{},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithGetSchemaAndTypeNameAndValidateConfig{
										GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
											return testSchema, nil
										},
										TypeNameMethod: func(_ context.Context, _ resource.TypeNameRequest, resp *resource.TypeNameResponse) {
											resp.TypeName = "test_resource"
										},
										Resource: &testprovider.Resource{},
										ValidateConfigMethod: func(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
											resp.Diagnostics.AddWarning("warning summary", "warning detail")
											resp.Diagnostics.AddError("error summary", "error detail")
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.ValidateResourceTypeConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_resource",
			},
			expectedResponse: &tfprotov5.ValidateResourceTypeConfigResponse{
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
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.ValidateResourceTypeConfig(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
