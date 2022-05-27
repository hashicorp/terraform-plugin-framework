package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerValidateResourceConfig(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test": tftypes.String,
		},
	}

	testValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test": tftypes.NewValue(tftypes.String, "test-value"),
	})

	testDynamicValue, err := tfprotov6.NewDynamicValue(testType, testValue)

	if err != nil {
		t.Fatalf("unexpected error calling tfprotov6.NewDynamicValue(): %s", err)
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
		request          *tfprotov6.ValidateResourceConfigRequest
		expectedError    error
		expectedResponse *tfprotov6.ValidateResourceConfigResponse
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_data_source": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return tfsdk.Schema{}, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ValidateResourceConfigRequest{
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ValidateResourceConfigResponse{},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_data_source": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.Resource{}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ValidateResourceConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ValidateResourceConfigResponse{},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						GetResourcesMethod: func(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
							return map[string]tfsdk.ResourceType{
								"test_data_source": &testprovider.ResourceType{
									GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
										return testSchema, nil
									},
									NewResourceMethod: func(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
										return &testprovider.ResourceWithValidateConfig{
											Resource: &testprovider.Resource{},
											ValidateConfigMethod: func(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
												resp.Diagnostics.AddWarning("warning summary", "warning detail")
												resp.Diagnostics.AddError("error summary", "error detail")
											},
										}, nil
									},
								},
							}, nil
						},
					},
				},
			},
			request: &tfprotov6.ValidateResourceConfigRequest{
				Config:   &testDynamicValue,
				TypeName: "test_data_source",
			},
			expectedResponse: &tfprotov6.ValidateResourceConfigResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov6.DiagnosticSeverityError,
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

			got, err := testCase.server.ValidateResourceConfig(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
