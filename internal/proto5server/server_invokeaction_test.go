// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerInvokeAction(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_required": tftypes.String,
		},
	}

	testConfigDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testEmptyDynamicValue := testNewDynamicValue(t, tftypes.Object{}, nil)

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_required": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testCases := map[string]struct {
		server         *Server
		request        *tfprotov5.InvokeActionRequest
		expectedError  error
		expectedEvents []tfprotov5.InvokeActionEvent
	}{
		"no-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = schema.Schema{}
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.InvokeActionRequest{
				Config:     testEmptyDynamicValue,
				ActionType: "test_action",
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{},
				},
			},
		},
		"request-config": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
											var config struct {
												TestRequired types.String `tfsdk:"test_required"`
											}

											resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

											if config.TestRequired.ValueString() != "test-config-value" {
												resp.Diagnostics.AddError("unexpected req.Config value: %s", config.TestRequired.ValueString())
											}
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.InvokeActionRequest{
				Config:     testConfigDynamicValue,
				ActionType: "test_action",
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{},
				},
			},
		},
		"response-progress-events": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
											resp.SendProgress(action.InvokeProgressEvent{Message: "progress event 1"})
											resp.SendProgress(action.InvokeProgressEvent{Message: "progress event 2"})
											resp.SendProgress(action.InvokeProgressEvent{Message: "progress event 3"})
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.InvokeActionRequest{
				Config:     testConfigDynamicValue,
				ActionType: "test_action",
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.ProgressInvokeActionEventType{
						Message: "progress event 1",
					},
				},
				{
					Type: tfprotov5.ProgressInvokeActionEventType{
						Message: "progress event 2",
					},
				},
				{
					Type: tfprotov5.ProgressInvokeActionEventType{
						Message: "progress event 3",
					},
				},
				{
					Type: tfprotov5.CompletedInvokeActionEventType{},
				},
			},
		},
		"response-diagnostics": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
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
			request: &tfprotov5.InvokeActionRequest{
				Config:     testConfigDynamicValue,
				ActionType: "test_action",
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
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
			},
		},
		"response-diagnostics-with-progress-events": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testSchema
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
											for i := 0; i < 5; i++ {
												resp.SendProgress(action.InvokeProgressEvent{Message: fmt.Sprintf("progress event %d", i+1)})
											}

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
			request: &tfprotov5.InvokeActionRequest{
				Config:     testConfigDynamicValue,
				ActionType: "test_action",
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.ProgressInvokeActionEventType{
						Message: "progress event 1",
					},
				},
				{
					Type: tfprotov5.ProgressInvokeActionEventType{
						Message: "progress event 2",
					},
				},
				{
					Type: tfprotov5.ProgressInvokeActionEventType{
						Message: "progress event 3",
					},
				},
				{
					Type: tfprotov5.ProgressInvokeActionEventType{
						Message: "progress event 4",
					},
				},
				{
					Type: tfprotov5.ProgressInvokeActionEventType{
						Message: "progress event 5",
					},
				},
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
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
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.InvokeAction(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedEvents, slices.Collect(got.Events)); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
