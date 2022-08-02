package fwserver_test

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
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerReadResource(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testCurrentStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, nil),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testNewStateValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"test_computed": tftypes.NewValue(tftypes.String, "test-newstate-value"),
		"test_required": tftypes.NewValue(tftypes.String, "test-currentstate-value"),
	})

	testSchema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"test_computed": {
				Computed: true,
				Type:     types.StringType,
			},
			"test_required": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}

	testConfig := &tfsdk.Config{
		Raw:    testCurrentStateValue,
		Schema: testSchema,
	}

	testCurrentState := &tfsdk.State{
		Raw:    testCurrentStateValue,
		Schema: testSchema,
	}

	testNewState := &tfsdk.State{
		Raw:    testNewStateValue,
		Schema: testSchema,
	}

	testNewStateRemoved := &tfsdk.State{
		Raw:    tftypes.NewValue(testType, nil),
		Schema: testSchema,
	}

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.ReadResourceRequest
		expectedResponse *fwserver.ReadResourceResponse
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.ReadResourceResponse{},
		},
		"request-currentstate-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{},
			expectedResponse: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Read Request",
						"An unexpected error was encountered when reading the resource. The current state was missing.\n\n"+
							"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
					),
				},
			},
		},
		"request-currentstate": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
								var data struct {
									TestComputed types.String `tfsdk:"test_computed"`
									TestRequired types.String `tfsdk:"test_required"`
								}

								resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

								if data.TestRequired.Value != "test-currentstate-value" {
									resp.Diagnostics.AddError("unexpected req.State value: %s", data.TestRequired.Value)
								}
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
			},
		},
		"request-providermeta": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
								var config struct {
									TestComputed types.String `tfsdk:"test_computed"`
									TestRequired types.String `tfsdk:"test_required"`
								}

								resp.Diagnostics.Append(req.ProviderMeta.Get(ctx, &config)...)

								if config.TestRequired.Value != "test-currentstate-value" {
									resp.Diagnostics.AddError("unexpected req.ProviderMeta value: %s", config.TestRequired.Value)
								}
							},
						}, nil
					},
				},
				ProviderMeta: testConfig,
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testCurrentState,
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
								resp.Diagnostics.AddWarning("warning summary", "warning detail")
								resp.Diagnostics.AddError("error summary", "error detail")
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic(
						"warning summary",
						"warning detail",
					),
					diag.NewErrorDiagnostic(
						"error summary",
						"error detail",
					),
				},
				NewState: testCurrentState,
			},
		},
		"response-state": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
								var data struct {
									TestComputed types.String `tfsdk:"test_computed"`
									TestRequired types.String `tfsdk:"test_required"`
								}

								resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

								data.TestComputed = types.String{Value: "test-newstate-value"}

								resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testNewState,
			},
		},
		"response-state-removeresource": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.ReadResourceRequest{
				CurrentState: testCurrentState,
				ResourceType: &testprovider.ResourceType{
					GetSchemaMethod: func(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
						return testSchema, nil
					},
					NewResourceMethod: func(_ context.Context, _ provider.Provider) (resource.Resource, diag.Diagnostics) {
						return &testprovider.Resource{
							ReadMethod: func(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
								resp.State.RemoveResource(ctx)
							},
						}, nil
					},
				},
			},
			expectedResponse: &fwserver.ReadResourceResponse{
				NewState: testNewStateRemoved,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.ReadResourceResponse{}
			testCase.server.ReadResource(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
