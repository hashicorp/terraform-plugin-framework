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

func TestServerImportResourceState(t *testing.T) {
	t.Parallel()

	type testCase struct {
		req *tfprotov6.ImportResourceStateRequest

		impl func(context.Context, tfsdk.ImportResourceStateRequest, *tfsdk.ImportResourceStateResponse)

		resp *tfprotov6.ImportResourceStateResponse
	}

	tests := map[string]testCase{
		"Set": {
			req: &tfprotov6.ImportResourceStateRequest{
				ID:       "test",
				TypeName: "test_import_state",
			},

			impl: func(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
				state := testServeResourceImportStateData{
					Id: req.ID,
				}

				diags := resp.State.Set(ctx, state)
				resp.Diagnostics.Append(diags...)
			},
			resp: &tfprotov6.ImportResourceStateResponse{
				ImportedResources: []*tfprotov6.ImportedResource{
					{
						State: func() *tfprotov6.DynamicValue {
							val, err := tfprotov6.NewDynamicValue(
								testServeResourceTypeImportStateTftype,
								tftypes.NewValue(
									testServeResourceTypeImportStateTftype,
									map[string]tftypes.Value{
										"id":              tftypes.NewValue(tftypes.String, "test"),
										"optional_string": tftypes.NewValue(tftypes.String, nil),
										"required_string": tftypes.NewValue(tftypes.String, ""),
									},
								),
							)
							if err != nil {
								panic(err)
							}
							return &val
						}(),
						TypeName: "test_import_state",
					},
				},
			},
		},
		"SetAttribute": {
			req: &tfprotov6.ImportResourceStateRequest{
				ID:       "test",
				TypeName: "test_import_state",
			},

			impl: func(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
				tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
			},

			resp: &tfprotov6.ImportResourceStateResponse{
				ImportedResources: []*tfprotov6.ImportedResource{
					{
						State: func() *tfprotov6.DynamicValue {
							val, err := tfprotov6.NewDynamicValue(
								testServeResourceTypeImportStateTftype,
								tftypes.NewValue(
									testServeResourceTypeImportStateTftype,
									map[string]tftypes.Value{
										"id":              tftypes.NewValue(tftypes.String, "test"),
										"optional_string": tftypes.NewValue(tftypes.String, nil),
										"required_string": tftypes.NewValue(tftypes.String, nil),
									},
								),
							)
							if err != nil {
								panic(err)
							}
							return &val
						}(),
						TypeName: "test_import_state",
					},
				},
			},
		},
		"imported_resource_conversion_error": {
			req: &tfprotov6.ImportResourceStateRequest{
				ID:       "test",
				TypeName: "test_import_state",
			},

			impl: func(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
				resp.State.Raw = tftypes.NewValue(tftypes.String, "this should never work")
			},

			resp: &tfprotov6.ImportResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Summary:  "Unable to Convert State",
						Severity: tfprotov6.DiagnosticSeverityError,
						Detail: "An unexpected error was encountered when converting the state to the protocol type. This is always an issue in the Terraform Provider SDK used to implement the provider and should be reported to the provider developers.\n\n" +
							"Please report this to the provider developer:\n\n" +
							`unexpected value type string, tftypes.Object["id":tftypes.String, "optional_string":tftypes.String, "required_string":tftypes.String] values must be of type map[string]tftypes.Value`,
					},
				},
			},
		},
		"no_state": {
			req: &tfprotov6.ImportResourceStateRequest{
				ID:       "test",
				TypeName: "test_import_state",
			},

			impl: func(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
			},

			resp: &tfprotov6.ImportResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Summary:  "Missing Resource Import State",
						Severity: tfprotov6.DiagnosticSeverityError,
						Detail: "An unexpected error was encountered when importing the resource. This is always a problem with the provider. Please give the following information to the provider developer:\n\n" +
							"Resource ImportState method returned no State in response. If import is intentionally not supported, remove the Resource type ImportState method or return an error.",
					},
				},
			},
		},
		"TypeName-ImportState-not-implemented": {
			req: &tfprotov6.ImportResourceStateRequest{
				ID:       "test",
				TypeName: "test_import_state_not_implemented",
			},
			resp: &tfprotov6.ImportResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Summary:  "Resource Import Not Implemented",
						Severity: tfprotov6.DiagnosticSeverityError,
						Detail:   "This resource does not support import. Please contact the provider developer for additional information.",
					},
				},
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := &testServeProvider{
				importStateFunc: tc.impl,
			}
			testServer := &Server{
				FrameworkServer: fwserver.Server{
					Provider: s,
				},
			}

			got, err := testServer.ImportResourceState(context.Background(), tc.req)

			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			if tc.req.TypeName == "test_import_state" && s.importResourceStateCalledResourceType != tc.req.TypeName {
				t.Errorf("Called wrong resource. Expected to call %q, actually called %q", tc.req.TypeName, s.importResourceStateCalledResourceType)
				return
			}

			if diff := cmp.Diff(got, tc.resp); diff != "" {
				t.Errorf("Unexpected diff in response (+wanted, -got): %s", diff)
			}
		})
	}
}
