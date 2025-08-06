// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerInvokeAction(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_required": tftypes.String,
		},
	}

	testLinkedResourceSchema := resourceschema.Schema{
		Attributes: map[string]resourceschema.Attribute{
			"test_computed": resourceschema.StringAttribute{
				Computed: true,
			},
			"test_required": resourceschema.StringAttribute{
				Required: true,
			},
		},
	}

	testLinkedResourceSchemaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_computed": tftypes.String,
			"test_required": tftypes.String,
		},
	}

	testLinkedResourceIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}

	testLinkedResourceIdentitySchemaType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_id": tftypes.String,
		},
	}

	testActionConfigDynamicValue := testNewDynamicValue(t, testType, map[string]tftypes.Value{
		"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
	})

	testEmptyDynamicValue := testNewDynamicValue(t, tftypes.Object{}, nil)

	testUnlinkedSchema := actionschema.UnlinkedSchema{
		Attributes: map[string]actionschema.Attribute{
			"test_required": actionschema.StringAttribute{
				Required: true,
			},
		},
	}

	testLifecycleSchema := actionschema.LifecycleSchema{
		Attributes: map[string]actionschema.Attribute{},
		LinkedResource: actionschema.LinkedResource{
			TypeName: "test_linked_resource",
		},
	}

	testLifecycleSchemaRaw := actionschema.LifecycleSchema{
		Attributes: map[string]actionschema.Attribute{},
		LinkedResource: actionschema.RawV5LinkedResource{
			TypeName: "test_linked_resource",
			Schema: func() *tfprotov5.Schema {
				return &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{
						Attributes: []*tfprotov5.SchemaAttribute{
							{
								Name:     "test_computed",
								Type:     tftypes.String,
								Computed: true,
							},
							{
								Name:     "test_required",
								Type:     tftypes.String,
								Required: true,
							},
						},
					},
				}
			},
			IdentitySchema: func() *tfprotov5.ResourceIdentitySchema {
				return &tfprotov5.ResourceIdentitySchema{
					IdentityAttributes: []*tfprotov5.ResourceIdentitySchemaAttribute{
						{
							Name:              "test_id",
							Type:              tftypes.String,
							RequiredForImport: true,
						},
					},
				}
			},
		},
	}

	testLifecycleSchemaRawNoIdentity := actionschema.LifecycleSchema{
		Attributes: map[string]actionschema.Attribute{},
		LinkedResource: actionschema.RawV5LinkedResource{
			TypeName: "test_linked_resource",
			Schema: func() *tfprotov5.Schema {
				return &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{
						Attributes: []*tfprotov5.SchemaAttribute{
							{
								Name:     "test_computed",
								Type:     tftypes.String,
								Computed: true,
							},
							{
								Name:     "test_required",
								Type:     tftypes.String,
								Required: true,
							},
						},
					},
				}
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
											resp.Schema = actionschema.UnlinkedSchema{}
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
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{},
					},
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
											resp.Schema = testUnlinkedSchema
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
				Config:     testActionConfigDynamicValue,
				ActionType: "test_action",
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{},
					},
				},
			},
		},
		"request-linkedresource-no-identity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = testLinkedResourceSchema
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_linked_resource"
										},
									}
								},
							}
						},
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testLifecycleSchema
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
											var linkedResourceData struct {
												TestRequired types.String `tfsdk:"test_required"`
												TestComputed types.String `tfsdk:"test_computed"`
											}

											if len(req.LinkedResources) != 1 {
												resp.Diagnostics.AddError("unexpected req.LinkedResources value", fmt.Sprintf("got %d, expected 1", len(req.LinkedResources)))
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Plan.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if !linkedResourceData.TestComputed.IsUnknown() {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be unknown, got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Config.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if !linkedResourceData.TestComputed.IsNull() {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be null, got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].State.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if linkedResourceData.TestComputed.ValueString() != "test-state-value" {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be \"test-state-value\", got: %s", linkedResourceData.TestComputed),
												)
												return
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
				Config:     testEmptyDynamicValue,
				ActionType: "test_action",
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						PriorState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						PlannedState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						Config: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, nil),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{
							{
								NewState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
									"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
									"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
								}),
							},
						},
					},
				},
			},
		},
		"request-linkedresource-with-identity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithIdentity{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testLinkedResourceSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_linked_resource"
											},
										},
										IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
											resp.IdentitySchema = testLinkedResourceIdentitySchema
										},
									}
								},
							}
						},
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testLifecycleSchema
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
											var linkedResourceData struct {
												TestRequired types.String `tfsdk:"test_required"`
												TestComputed types.String `tfsdk:"test_computed"`
											}
											var linkedResourceIdentityData struct {
												TestID types.String `tfsdk:"test_id"`
											}

											if len(req.LinkedResources) != 1 {
												resp.Diagnostics.AddError("unexpected req.LinkedResources value", fmt.Sprintf("got %d, expected 1", len(req.LinkedResources)))
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Plan.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if !linkedResourceData.TestComputed.IsUnknown() {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be unknown, got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Config.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if !linkedResourceData.TestComputed.IsNull() {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be null, got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].State.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if linkedResourceData.TestComputed.ValueString() != "test-state-value" {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be \"test-state-value\", got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Identity.Get(ctx, &linkedResourceIdentityData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if linkedResourceIdentityData.TestID.ValueString() != "id-123" {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be \"id-123\", got: %s", linkedResourceIdentityData.TestID),
												)
												return
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
				Config:     testEmptyDynamicValue,
				ActionType: "test_action",
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						PriorState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						PlannedState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						Config: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, nil),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						PlannedIdentity: &tfprotov5.ResourceIdentityData{
							IdentityData: testNewDynamicValue(t, testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
						},
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{
							{
								NewState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
									"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
									"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
								}),
								NewIdentity: &tfprotov5.ResourceIdentityData{
									IdentityData: testNewDynamicValue(t, testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
										"test_id": tftypes.NewValue(tftypes.String, "id-123"),
									}),
								},
							},
						},
					},
				},
			},
		},
		"request-raw-linkedresource-no-identity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testLifecycleSchemaRawNoIdentity
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
											var linkedResourceData struct {
												TestRequired types.String `tfsdk:"test_required"`
												TestComputed types.String `tfsdk:"test_computed"`
											}

											if len(req.LinkedResources) != 1 {
												resp.Diagnostics.AddError("unexpected req.LinkedResources value", fmt.Sprintf("got %d, expected 1", len(req.LinkedResources)))
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Plan.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if !linkedResourceData.TestComputed.IsUnknown() {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be unknown, got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Config.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if !linkedResourceData.TestComputed.IsNull() {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be null, got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].State.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if linkedResourceData.TestComputed.ValueString() != "test-state-value" {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be \"test-state-value\", got: %s", linkedResourceData.TestComputed),
												)
												return
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
				Config:     testEmptyDynamicValue,
				ActionType: "test_action",
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						PriorState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						PlannedState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						Config: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, nil),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{
							{
								NewState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
									"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
									"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
								}),
							},
						},
					},
				},
			},
		},
		"request-raw-linkedresource-with-identity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testLifecycleSchemaRaw
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
											var linkedResourceData struct {
												TestRequired types.String `tfsdk:"test_required"`
												TestComputed types.String `tfsdk:"test_computed"`
											}
											var linkedResourceIdentityData struct {
												TestID types.String `tfsdk:"test_id"`
											}

											if len(req.LinkedResources) != 1 {
												resp.Diagnostics.AddError("unexpected req.LinkedResources value", fmt.Sprintf("got %d, expected 1", len(req.LinkedResources)))
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Plan.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if !linkedResourceData.TestComputed.IsUnknown() {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be unknown, got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Config.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if !linkedResourceData.TestComputed.IsNull() {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be null, got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].State.Get(ctx, &linkedResourceData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if linkedResourceData.TestComputed.ValueString() != "test-state-value" {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be \"test-state-value\", got: %s", linkedResourceData.TestComputed),
												)
												return
											}

											resp.Diagnostics.Append(req.LinkedResources[0].Identity.Get(ctx, &linkedResourceIdentityData)...)
											if resp.Diagnostics.HasError() {
												return
											}

											if linkedResourceIdentityData.TestID.ValueString() != "id-123" {
												resp.Diagnostics.AddError(
													"unexpected req.LinkedResources value",
													fmt.Sprintf("expected linked resource data to be \"id-123\", got: %s", linkedResourceIdentityData.TestID),
												)
												return
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
				Config:     testEmptyDynamicValue,
				ActionType: "test_action",
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						PriorState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						PlannedState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						Config: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, nil),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						PlannedIdentity: &tfprotov5.ResourceIdentityData{
							IdentityData: testNewDynamicValue(t, testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
						},
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{
							{
								NewState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
									"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
									"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
								}),
								NewIdentity: &tfprotov5.ResourceIdentityData{
									IdentityData: testNewDynamicValue(t, testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
										"test_id": tftypes.NewValue(tftypes.String, "id-123"),
									}),
								},
							},
						},
					},
				},
			},
		},
		"response-linkedresource-no-identity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = testLinkedResourceSchema
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_linked_resource"
										},
									}
								},
							}
						},
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testLifecycleSchema
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
											// Should be copied over from request
											if len(resp.LinkedResources) != 1 {
												resp.Diagnostics.AddError("unexpected resp.LinkedResources value", fmt.Sprintf("got %d, expected 1", len(req.LinkedResources)))
											}

											resp.Diagnostics.Append(resp.LinkedResources[0].State.SetAttribute(ctx, path.Root("test_computed"), "new-state-value")...)
											if resp.Diagnostics.HasError() {
												return
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
				Config:     testEmptyDynamicValue,
				ActionType: "test_action",
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						PriorState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						PlannedState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						Config: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, nil),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{
							{
								NewState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
									"test_computed": tftypes.NewValue(tftypes.String, "new-state-value"),
									"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
								}),
							},
						},
					},
				},
			},
		},
		"response-linkedresource-with-identity": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithIdentity{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testLinkedResourceSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_linked_resource"
											},
										},
										IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
											resp.IdentitySchema = testLinkedResourceIdentitySchema
										},
									}
								},
							}
						},
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testLifecycleSchema
										},
										MetadataMethod: func(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
											resp.TypeName = "test_action"
										},
										InvokeMethod: func(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
											// Should be copied over from request
											if len(resp.LinkedResources) != 1 {
												resp.Diagnostics.AddError("unexpected resp.LinkedResources value", fmt.Sprintf("got %d, expected 1", len(req.LinkedResources)))
											}

											resp.Diagnostics.Append(resp.LinkedResources[0].State.SetAttribute(ctx, path.Root("test_computed"), "new-state-value")...)
											if resp.Diagnostics.HasError() {
												return
											}

											resp.Diagnostics.Append(resp.LinkedResources[0].Identity.SetAttribute(ctx, path.Root("test_id"), "new-id-123")...)
											if resp.Diagnostics.HasError() {
												return
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
				Config:     testEmptyDynamicValue,
				ActionType: "test_action",
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						PriorState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, "test-state-value"),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						PlannedState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						Config: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
							"test_computed": tftypes.NewValue(tftypes.String, nil),
							"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						}),
						PlannedIdentity: &tfprotov5.ResourceIdentityData{
							IdentityData: testNewDynamicValue(t, testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
								"test_id": tftypes.NewValue(tftypes.String, "id-123"),
							}),
						},
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{
							{
								NewState: testNewDynamicValue(t, testLinkedResourceSchemaType, map[string]tftypes.Value{
									"test_computed": tftypes.NewValue(tftypes.String, "new-state-value"),
									"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
								}),
								NewIdentity: &tfprotov5.ResourceIdentityData{
									IdentityData: testNewDynamicValue(t, testLinkedResourceIdentitySchemaType, map[string]tftypes.Value{
										"test_id": tftypes.NewValue(tftypes.String, "new-id-123"),
									}),
								},
							},
						},
					},
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
											resp.Schema = testUnlinkedSchema
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
				Config:     testActionConfigDynamicValue,
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
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{},
					},
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
											resp.Schema = testUnlinkedSchema
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
				Config:     testActionConfigDynamicValue,
				ActionType: "test_action",
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						LinkedResources: []*tfprotov5.NewLinkedResource{},
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
											resp.Schema = testUnlinkedSchema
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
				Config:     testActionConfigDynamicValue,
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
						LinkedResources: []*tfprotov5.NewLinkedResource{},
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
		"response-linkedresource-not-found": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.ResourceWithIdentity{
										Resource: &testprovider.Resource{
											SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
												resp.Schema = testLinkedResourceSchema
											},
											MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
												resp.TypeName = "test_not_the_right_resource"
											},
										},
										IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
											resp.IdentitySchema = testLinkedResourceIdentitySchema
										},
									}
								},
							}
						},
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = testLifecycleSchema
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
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						// No data setup needed, should error before decoding logic
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						Diagnostics: []*tfprotov5.Diagnostic{
							{
								Severity: tfprotov5.DiagnosticSeverityError,
								Summary:  "Invalid Linked Resource Schema",
								Detail: "An unexpected error was encountered when converting \"test_linked_resource\" linked resource data from the protocol type. " +
									"This is always an issue in the provider code and should be reported to the provider developers.\n\nPlease report this to the provider developer:\n\n" +
									"The \"test_linked_resource\" linked resource was not found on the provider server.",
							},
						},
					},
				},
			},
		},
		"response-raw-linkedresource-invalid-resource-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = actionschema.LifecycleSchema{
												Attributes: map[string]actionschema.Attribute{},
												LinkedResource: actionschema.RawV5LinkedResource{
													TypeName: "test_invalid_linked_resource",
													Schema: func() *tfprotov5.Schema {
														return &tfprotov5.Schema{
															Block: &tfprotov5.SchemaBlock{
																Attributes: []*tfprotov5.SchemaAttribute{
																	// Tuple is not supported in framework
																	{
																		Name:     "test_tuple",
																		Type:     tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.Bool}},
																		Required: true,
																	},
																},
															},
														}
													},
												},
											}
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
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						// No data setup needed, should error before decoding logic
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						Diagnostics: []*tfprotov5.Diagnostic{
							{
								Severity: tfprotov5.DiagnosticSeverityError,
								Summary:  "Invalid Linked Resource Schema",
								Detail: "An unexpected error was encountered when converting \"test_invalid_linked_resource\" linked resource schema from the protocol type. " +
									"This is always an issue in the provider code and should be reported to the provider developers.\n\nPlease report this to the provider developer:\n\n" +
									"no supported attribute for \"test_tuple\", type: tftypes.Tuple",
							},
						},
					},
				},
			},
		},
		"response-raw-linkedresource-invalid-identity-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = actionschema.LifecycleSchema{
												Attributes: map[string]actionschema.Attribute{},
												LinkedResource: actionschema.RawV5LinkedResource{
													TypeName: "test_linked_resource",
													Schema: func() *tfprotov5.Schema {
														return &tfprotov5.Schema{
															Block: &tfprotov5.SchemaBlock{
																Attributes: []*tfprotov5.SchemaAttribute{
																	{
																		Name:     "test_computed",
																		Type:     tftypes.String,
																		Computed: true,
																	},
																	{
																		Name:     "test_required",
																		Type:     tftypes.String,
																		Required: true,
																	},
																},
															},
														}
													},
													IdentitySchema: func() *tfprotov5.ResourceIdentitySchema {
														return &tfprotov5.ResourceIdentitySchema{
															IdentityAttributes: []*tfprotov5.ResourceIdentitySchemaAttribute{
																// Set is not a valid type for resource identity
																{
																	Name:              "test_id",
																	Type:              tftypes.Set{ElementType: tftypes.Bool},
																	RequiredForImport: true,
																},
															},
														}
													},
												},
											}
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
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						// No data setup needed, should error before decoding logic
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						Diagnostics: []*tfprotov5.Diagnostic{
							{
								Severity: tfprotov5.DiagnosticSeverityError,
								Summary:  "Invalid Linked Resource Schema",
								Detail: "An unexpected error was encountered when converting \"test_linked_resource\" linked resource identity schema from the protocol type. " +
									"This is always an issue in the provider code and should be reported to the provider developers.\n\nPlease report this to the provider developer:\n\n" +
									"no supported identity attribute for \"test_id\", type: tftypes.Set",
							},
						},
					},
				},
			},
		},
		"response-raw-linkedresource-v6-resource-schema": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ActionsMethod: func(_ context.Context) []func() action.Action {
							return []func() action.Action{
								func() action.Action {
									return &testprovider.Action{
										SchemaMethod: func(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
											resp.Schema = actionschema.LifecycleSchema{
												Attributes: map[string]actionschema.Attribute{},
												LinkedResource: actionschema.RawV6LinkedResource{
													TypeName: "test_v6_linked_resource",
													Schema: func() *tfprotov6.Schema {
														return &tfprotov6.Schema{
															Block: &tfprotov6.SchemaBlock{
																Attributes: []*tfprotov6.SchemaAttribute{
																	{
																		Name:     "test_computed",
																		Type:     tftypes.String,
																		Computed: true,
																	},
																	{
																		Name:     "test_required",
																		Type:     tftypes.String,
																		Required: true,
																	},
																},
															},
														}
													},
												},
											}
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
				LinkedResources: []*tfprotov5.InvokeLinkedResource{
					{
						// No data setup needed, should error before decoding logic
					},
				},
			},
			expectedEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{
						Diagnostics: []*tfprotov5.Diagnostic{
							{
								Severity: tfprotov5.DiagnosticSeverityError,
								Summary:  "Invalid Linked Resource Schema",
								Detail: "An unexpected error was encountered when converting \"test_v6_linked_resource\" linked resource schema from the protocol type. " +
									"This is always an issue in the provider code and should be reported to the provider developers.\n\nPlease report this to the provider developer:\n\n" +
									"The \"test_v6_linked_resource\" linked resource is a protocol v6 resource but the provider is being served using protocol v5.",
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
