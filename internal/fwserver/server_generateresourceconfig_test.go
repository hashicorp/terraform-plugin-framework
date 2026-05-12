// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	schemavalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerGenerateResourceConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.GenerateResourceConfigRequest
		expectedResponse     *fwserver.GenerateResourceConfigResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{},
		},
		"request-state-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Generate Config Request",
						"An unexpected error was encountered when generating resource configuration. The current state was missing.\n\n"+
							"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
					),
				},
			},
		},
		"response-default-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"test_computed": schema.StringAttribute{
							Computed: true,
						},
						"test_optional": schema.StringAttribute{
							Optional: true,
						},
						"test_required": schema.StringAttribute{
							Required: true,
						},
						"test_deprecated": schema.ListAttribute{
							ElementType:        types.StringType,
							Optional:           true,
							DeprecationMessage: "deprecated",
						},
						"test_false_bool": schema.BoolAttribute{
							Optional: true,
						},
						"test_empty_string": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"test_deprecated_block": schema.ListNestedBlock{
							DeprecationMessage: "deprecated",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"test_nested_block_attr": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"test_nested_block": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"test_nested_nested_block": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"test_computed": schema.StringAttribute{
													Computed: true,
												},
												"test_optional": schema.StringAttribute{
													Optional: true,
												},
												"test_required": schema.StringAttribute{
													Required: true,
												},
												"test_deprecated": schema.ListAttribute{
													ElementType:        types.StringType,
													Optional:           true,
													DeprecationMessage: "deprecated",
												},
											},
										},
									},
								},
							},
						},
						"test_nested_deprecated_block": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"test_nested_nested_block": schema.ListNestedBlock{
										DeprecationMessage: "deprecated",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"test_nested_nested_block_attr": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                tftypes.String,
							"test_computed":     tftypes.String,
							"test_optional":     tftypes.String,
							"test_required":     tftypes.String,
							"test_deprecated":   tftypes.List{ElementType: tftypes.String},
							"test_false_bool":   tftypes.Bool,
							"test_empty_string": tftypes.String,
							"test_deprecated_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_block_attr": tftypes.String,
								},
							}},
							"test_nested_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_computed":   tftypes.String,
											"test_optional":   tftypes.String,
											"test_required":   tftypes.String,
											"test_deprecated": tftypes.List{ElementType: tftypes.String},
										},
									}},
								},
							}},
							"test_nested_deprecated_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_nested_nested_block_attr": tftypes.String,
										},
									}},
								},
							}},
						},
					}, map[string]tftypes.Value{
						"id":            tftypes.NewValue(tftypes.String, "test-id-val"),
						"test_computed": tftypes.NewValue(tftypes.String, "test-computed-val"),
						"test_optional": tftypes.NewValue(tftypes.String, "test-optional-val"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_deprecated": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-deprecated-a"),
							tftypes.NewValue(tftypes.String, "test-deprecated-b"),
						}),
						"test_false_bool":   tftypes.NewValue(tftypes.Bool, false),
						"test_empty_string": tftypes.NewValue(tftypes.String, ""),
						"test_deprecated_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_nested_block_attr": tftypes.String,
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_block_attr": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"test_nested_block_attr": tftypes.NewValue(tftypes.String, "test-nested-block-val-a"),
							}),
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_block_attr": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"test_nested_block_attr": tftypes.NewValue(tftypes.String, "test-nested-block-val-b"),
							}),
						}),
						"test_nested_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_computed":   tftypes.String,
										"test_optional":   tftypes.String,
										"test_required":   tftypes.String,
										"test_deprecated": tftypes.List{ElementType: tftypes.String},
									},
								}},
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_computed":   tftypes.String,
											"test_optional":   tftypes.String,
											"test_required":   tftypes.String,
											"test_deprecated": tftypes.List{ElementType: tftypes.String},
										},
									}},
								},
							}, map[string]tftypes.Value{
								"test_nested_nested_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_computed":   tftypes.String,
										"test_optional":   tftypes.String,
										"test_required":   tftypes.String,
										"test_deprecated": tftypes.List{ElementType: tftypes.String},
									},
								}}, []tftypes.Value{
									tftypes.NewValue(tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_computed":   tftypes.String,
											"test_optional":   tftypes.String,
											"test_required":   tftypes.String,
											"test_deprecated": tftypes.List{ElementType: tftypes.String},
										},
									}, map[string]tftypes.Value{
										"test_computed":   tftypes.NewValue(tftypes.String, "computed-val-a"),
										"test_optional":   tftypes.NewValue(tftypes.String, "optional-val-a"),
										"test_required":   tftypes.NewValue(tftypes.String, "required-val-a"),
										"test_deprecated": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "hello-a"), tftypes.NewValue(tftypes.String, "world-a")}),
									}),
									tftypes.NewValue(tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_computed":   tftypes.String,
											"test_optional":   tftypes.String,
											"test_required":   tftypes.String,
											"test_deprecated": tftypes.List{ElementType: tftypes.String},
										},
									}, map[string]tftypes.Value{
										"test_computed":   tftypes.NewValue(tftypes.String, "computed-val-b"),
										"test_optional":   tftypes.NewValue(tftypes.String, "optional-val-b"),
										"test_required":   tftypes.NewValue(tftypes.String, "required-val-b"),
										"test_deprecated": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "hello-b"), tftypes.NewValue(tftypes.String, "world-b")}),
									}),
								}),
							}),
						}),
						"test_nested_deprecated_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_nested_nested_block_attr": tftypes.String,
									},
								}},
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_nested_nested_block_attr": tftypes.String,
										},
									}},
								},
							}, map[string]tftypes.Value{
								"test_nested_nested_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_nested_nested_block_attr": tftypes.String,
									},
								}}, []tftypes.Value{
									tftypes.NewValue(tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_nested_nested_block_attr": tftypes.String,
										},
									}, map[string]tftypes.Value{
										"test_nested_nested_block_attr": tftypes.NewValue(tftypes.String, "val-a"),
									}),
									tftypes.NewValue(tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_nested_nested_block_attr": tftypes.String,
										},
									}, map[string]tftypes.Value{
										"test_nested_nested_block_attr": tftypes.NewValue(tftypes.String, "val-b"),
									}),
								}),
							}),
						}),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							"test_computed": schema.StringAttribute{
								Computed: true,
							},
							"test_optional": schema.StringAttribute{
								Optional: true,
							},
							"test_required": schema.StringAttribute{
								Required: true,
							},
							"test_deprecated": schema.ListAttribute{
								ElementType:        types.StringType,
								Optional:           true,
								DeprecationMessage: "deprecated",
							},
							"test_false_bool": schema.BoolAttribute{
								Optional: true,
							},
							"test_empty_string": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"test_deprecated_block": schema.ListNestedBlock{
								DeprecationMessage: "deprecated",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"test_nested_block_attr": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
							"test_nested_block": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"test_nested_nested_block": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"test_computed": schema.StringAttribute{
														Computed: true,
													},
													"test_optional": schema.StringAttribute{
														Optional: true,
													},
													"test_required": schema.StringAttribute{
														Required: true,
													},
													"test_deprecated": schema.ListAttribute{
														ElementType:        types.StringType,
														Optional:           true,
														DeprecationMessage: "deprecated",
													},
												},
											},
										},
									},
								},
							},
							"test_nested_deprecated_block": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"test_nested_nested_block": schema.ListNestedBlock{
											DeprecationMessage: "deprecated",
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"test_nested_nested_block_attr": schema.StringAttribute{
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                tftypes.String,
							"test_computed":     tftypes.String,
							"test_optional":     tftypes.String,
							"test_required":     tftypes.String,
							"test_deprecated":   tftypes.List{ElementType: tftypes.String},
							"test_false_bool":   tftypes.Bool,
							"test_empty_string": tftypes.String,
							"test_deprecated_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_block_attr": tftypes.String,
								},
							}},
							"test_nested_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_computed":   tftypes.String,
											"test_optional":   tftypes.String,
											"test_required":   tftypes.String,
											"test_deprecated": tftypes.List{ElementType: tftypes.String},
										},
									}},
								},
							}},
							"test_nested_deprecated_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_nested_nested_block_attr": tftypes.String,
										},
									}},
								},
							}},
						},
					}, map[string]tftypes.Value{
						"id":                tftypes.NewValue(tftypes.String, "test-id-val"), // this should stay for framework as id is not special here
						"test_computed":     tftypes.NewValue(tftypes.String, nil),
						"test_optional":     tftypes.NewValue(tftypes.String, "test-optional-val"),
						"test_required":     tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_deprecated":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_false_bool":   tftypes.NewValue(tftypes.Bool, false),
						"test_empty_string": tftypes.NewValue(tftypes.String, ""), // for framework an empty string and null are not the same, so it should stay
						"test_deprecated_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_nested_block_attr": tftypes.String,
							},
						}}, nil),
						"test_nested_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_computed":   tftypes.String,
										"test_optional":   tftypes.String,
										"test_required":   tftypes.String,
										"test_deprecated": tftypes.List{ElementType: tftypes.String},
									},
								}},
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_computed":   tftypes.String,
											"test_optional":   tftypes.String,
											"test_required":   tftypes.String,
											"test_deprecated": tftypes.List{ElementType: tftypes.String},
										},
									}},
								},
							}, map[string]tftypes.Value{
								"test_nested_nested_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_computed":   tftypes.String,
										"test_optional":   tftypes.String,
										"test_required":   tftypes.String,
										"test_deprecated": tftypes.List{ElementType: tftypes.String},
									},
								}}, []tftypes.Value{
									tftypes.NewValue(tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_computed":   tftypes.String,
											"test_optional":   tftypes.String,
											"test_required":   tftypes.String,
											"test_deprecated": tftypes.List{ElementType: tftypes.String},
										},
									}, map[string]tftypes.Value{
										"test_computed":   tftypes.NewValue(tftypes.String, nil),
										"test_optional":   tftypes.NewValue(tftypes.String, "optional-val-a"),
										"test_required":   tftypes.NewValue(tftypes.String, "required-val-a"),
										"test_deprecated": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
									}),
									tftypes.NewValue(tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_computed":   tftypes.String,
											"test_optional":   tftypes.String,
											"test_required":   tftypes.String,
											"test_deprecated": tftypes.List{ElementType: tftypes.String},
										},
									}, map[string]tftypes.Value{
										"test_computed":   tftypes.NewValue(tftypes.String, nil),
										"test_optional":   tftypes.NewValue(tftypes.String, "optional-val-b"),
										"test_required":   tftypes.NewValue(tftypes.String, "required-val-b"),
										"test_deprecated": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
									}),
								}),
							}),
						}),
						"test_nested_deprecated_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_nested_nested_block_attr": tftypes.String,
									},
								}},
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_nested_nested_block": tftypes.List{ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_nested_nested_block_attr": tftypes.String,
										},
									}},
								},
							}, map[string]tftypes.Value{
								"test_nested_nested_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_nested_nested_block_attr": tftypes.String,
									},
								}}, nil),
							}),
						}),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							"test_computed": schema.StringAttribute{
								Computed: true,
							},
							"test_optional": schema.StringAttribute{
								Optional: true,
							},
							"test_required": schema.StringAttribute{
								Required: true,
							},
							"test_deprecated": schema.ListAttribute{
								ElementType:        types.StringType,
								Optional:           true,
								DeprecationMessage: "deprecated",
							},
							"test_false_bool": schema.BoolAttribute{
								Optional: true,
							},
							"test_empty_string": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"test_deprecated_block": schema.ListNestedBlock{
								DeprecationMessage: "deprecated",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"test_nested_block_attr": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
							"test_nested_block": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"test_nested_nested_block": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"test_computed": schema.StringAttribute{
														Computed: true,
													},
													"test_optional": schema.StringAttribute{
														Optional: true,
													},
													"test_required": schema.StringAttribute{
														Required: true,
													},
													"test_deprecated": schema.ListAttribute{
														ElementType:        types.StringType,
														Optional:           true,
														DeprecationMessage: "deprecated",
													},
												},
											},
										},
									},
								},
							},
							"test_nested_deprecated_block": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"test_nested_nested_block": schema.ListNestedBlock{
											DeprecationMessage: "deprecated",
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"test_nested_nested_block_attr": schema.StringAttribute{
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"response-conflicts-with-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"alpha": schema.StringAttribute{
							Optional: true,
							Validators: []schemavalidator.String{
								testConflictsWithStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
							},
						},
						"beta": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testConflictsWithStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testConflictsWithStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
		"response-exactly-one-of-group-all-null": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"alpha": schema.StringAttribute{
							Optional: true,
							Validators: []schemavalidator.String{
								testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
							},
						},
						"beta": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
		"response-also-requires-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"alpha": schema.StringAttribute{
							Optional: true,
							Validators: []schemavalidator.String{
								testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
							},
						},
						"beta": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
		"response-resource-conflicts-with-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"alpha": schema.StringAttribute{
							Optional: true,
							Validators: []schemavalidator.String{
								testConflictsWithStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
							},
						},
						"beta": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				Resource: &testprovider.ResourceWithConfigValidators{
					Resource: &testprovider.Resource{},
					ConfigValidatorsMethod: func(context.Context) []resource.ConfigValidator {
						return []resource.ConfigValidator{&testResourceConflictsWithValidator{paths: path.Expressions{
							path.MatchRoot("alpha"),
							path.MatchRoot("beta"),
						}}}
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testConflictsWithStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testConflictsWithStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
		"response-resource-exactly-one-of-group-all-null": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"alpha": schema.StringAttribute{
							Optional: true,
							Validators: []schemavalidator.String{
								testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
							},
						},
						"beta": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				Resource: &testprovider.ResourceWithConfigValidators{
					Resource: &testprovider.Resource{},
					ConfigValidatorsMethod: func(context.Context) []resource.ConfigValidator {
						return []resource.ConfigValidator{&testResourceExactlyOneOfValidator{paths: path.Expressions{
							path.MatchRoot("alpha"),
							path.MatchRoot("beta"),
						}}}
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
		"response-exactly-one-of-group-multiple-non-null": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"alpha": schema.StringAttribute{
							Optional: true,
							Validators: []schemavalidator.String{
								testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta"), path.MatchRoot("gamma")}},
							},
						},
						"beta": schema.StringAttribute{
							Optional: true,
						},
						"gamma": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
							"gamma": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
						"gamma": tftypes.NewValue(tftypes.String, "configured-gamma"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta"), path.MatchRoot("gamma")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
							"gamma": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
							"gamma": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, nil),
						"gamma": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta"), path.MatchRoot("gamma")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
							"gamma": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
		"response-resource-also-requires-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"alpha": schema.StringAttribute{
							Optional: true,
							Validators: []schemavalidator.String{
								testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
							},
						},
						"beta": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				Resource: &testprovider.ResourceWithConfigValidators{
					Resource: &testprovider.Resource{},
					ConfigValidatorsMethod: func(context.Context) []resource.ConfigValidator {
						return []resource.ConfigValidator{&testResourceAlsoRequiresValidator{paths: path.Expressions{
							path.MatchRoot("alpha"),
							path.MatchRoot("beta"),
						}}}
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
		"response-also-requires-group-complete": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"alpha": schema.StringAttribute{
							Optional: true,
							Validators: []schemavalidator.String{
								testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
							},
						},
						"beta": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
		"response-resource-also-requires-group-non-member": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"alpha": schema.StringAttribute{
							Optional: true,
							Validators: []schemavalidator.String{
								testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta"), path.MatchRoot("gamma")}},
							},
						},
						"beta": schema.StringAttribute{
							Optional: true,
						},
						"gamma": schema.StringAttribute{
							Optional: true,
						},
					},
				},
				Resource: &testprovider.ResourceWithConfigValidators{
					Resource: &testprovider.Resource{},
					ConfigValidatorsMethod: func(context.Context) []resource.ConfigValidator {
						return []resource.ConfigValidator{&testResourceAlsoRequiresValidator{paths: path.Expressions{
							path.MatchRoot("alpha"),
							path.MatchRoot("beta"),
						}}}
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
							"gamma": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
						"gamma": tftypes.NewValue(tftypes.String, "configured-gamma"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta"), path.MatchRoot("gamma")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
							"gamma": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha": tftypes.String,
							"beta":  tftypes.String,
							"gamma": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
						"gamma": tftypes.NewValue(tftypes.String, "configured-gamma"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"alpha": schema.StringAttribute{
								Optional: true,
								Validators: []schemavalidator.String{
									testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta"), path.MatchRoot("gamma")}},
								},
							},
							"beta": schema.StringAttribute{
								Optional: true,
							},
							"gamma": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
		"response-block-conflicts-with-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"beta": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"alpha_block": schema.ListNestedBlock{
							Validators: []schemavalidator.List{
								testConflictsWithListValidator{paths: path.Expressions{path.MatchRoot("beta")}},
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"value": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"value": tftypes.String,
								},
							}},
							"beta": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"value": tftypes.String,
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"value": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"value": tftypes.NewValue(tftypes.String, "configured-alpha-block"),
							}),
						}),
						"beta": tftypes.NewValue(tftypes.String, "configured-beta"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"alpha_block": schema.ListNestedBlock{
								Validators: []schemavalidator.List{
									testConflictsWithListValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"value": tftypes.String,
								},
							}},
							"beta": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"value": tftypes.String,
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"value": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"value": tftypes.NewValue(tftypes.String, "configured-alpha-block"),
							}),
						}),
						"beta": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"alpha_block": schema.ListNestedBlock{
								Validators: []schemavalidator.List{
									testConflictsWithListValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"response-nested-block-object-conflicts-with-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"beta": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"alpha_block": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Validators: []schemavalidator.Object{
									testConflictsWithObjectValidator{paths: path.Expressions{path.MatchRoot("beta")}},
								},
								Attributes: map[string]schema.Attribute{
									"value": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"value": tftypes.String,
								},
							}},
							"beta": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"value": tftypes.String,
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"value": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"value": tftypes.NewValue(tftypes.String, "configured-alpha-block"),
							}),
						}),
						"beta": tftypes.NewValue(tftypes.String, "configured-beta"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"alpha_block": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Validators: []schemavalidator.Object{
										testConflictsWithObjectValidator{paths: path.Expressions{path.MatchRoot("beta")}},
									},
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"alpha_block": tftypes.List{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"value": tftypes.String,
								},
							}},
							"beta": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"alpha_block": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"value": tftypes.String,
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"value": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"value": tftypes.NewValue(tftypes.String, "configured-alpha-block"),
							}),
						}),
						// beta should be nullified because it conflicts with alpha_block
						"beta": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"beta": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"alpha_block": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Validators: []schemavalidator.Object{
										testConflictsWithObjectValidator{paths: path.Expressions{path.MatchRoot("beta")}},
									},
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"response-all-attribute-types": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"test_int64": schema.Int64Attribute{
							Optional: true,
						},
						"test_int64_computed": schema.Int64Attribute{
							Computed: true,
						},
						"test_int32": schema.Int32Attribute{
							Optional: true,
						},
						"test_float64": schema.Float64Attribute{
							Optional: true,
						},
						"test_float32": schema.Float32Attribute{
							Optional: true,
						},
						"test_dynamic": schema.DynamicAttribute{
							Optional: true,
						},
						"test_object": schema.ObjectAttribute{
							Optional: true,
							AttributeTypes: map[string]attr.Type{
								"obj_str": types.StringType,
							},
						},
						"test_map": schema.MapAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"test_set": schema.SetAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"test_list_nested": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested_str": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"test_map_nested": schema.MapNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested_str": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"test_set_nested": schema.SetNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested_str": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_int64":          tftypes.Number,
							"test_int64_computed": tftypes.Number,
							"test_int32":          tftypes.Number,
							"test_float64":        tftypes.Number,
							"test_float32":        tftypes.Number,
							"test_dynamic":        tftypes.DynamicPseudoType,
							"test_object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"obj_str": tftypes.String,
								},
							},
							"test_map":         tftypes.Map{ElementType: tftypes.String},
							"test_set":         tftypes.Set{ElementType: tftypes.String},
							"test_list_nested": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}},
							"test_map_nested":  tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}},
							"test_set_nested":  tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}},
						},
					}, map[string]tftypes.Value{
						"test_int64":          tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
						"test_int64_computed": tftypes.NewValue(tftypes.Number, big.NewFloat(42)),
						"test_int32":          tftypes.NewValue(tftypes.Number, big.NewFloat(99)),
						"test_float64":        tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
						"test_float32":        tftypes.NewValue(tftypes.Number, big.NewFloat(3.14)),
						"test_dynamic":        tftypes.NewValue(tftypes.String, "dynamic-val"),
						"test_object": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"obj_str": tftypes.String}}, map[string]tftypes.Value{
							"obj_str": tftypes.NewValue(tftypes.String, "obj-val"),
						}),
						"test_map": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
							"key1": tftypes.NewValue(tftypes.String, "val1"),
						}),
						"test_set": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "set-val"),
						}),
						"test_list_nested": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}, map[string]tftypes.Value{
								"nested_str": tftypes.NewValue(tftypes.String, "list-nested-val"),
							}),
						}),
						"test_map_nested": tftypes.NewValue(tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}}, map[string]tftypes.Value{
							"mk1": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}, map[string]tftypes.Value{
								"nested_str": tftypes.NewValue(tftypes.String, "map-nested-val"),
							}),
						}),
						"test_set_nested": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}, map[string]tftypes.Value{
								"nested_str": tftypes.NewValue(tftypes.String, "set-nested-val"),
							}),
						}),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test_int64": schema.Int64Attribute{
								Optional: true,
							},
							"test_int64_computed": schema.Int64Attribute{
								Computed: true,
							},
							"test_int32": schema.Int32Attribute{
								Optional: true,
							},
							"test_float64": schema.Float64Attribute{
								Optional: true,
							},
							"test_float32": schema.Float32Attribute{
								Optional: true,
							},
							"test_dynamic": schema.DynamicAttribute{
								Optional: true,
							},
							"test_object": schema.ObjectAttribute{
								Optional: true,
								AttributeTypes: map[string]attr.Type{
									"obj_str": types.StringType,
								},
							},
							"test_map": schema.MapAttribute{
								Optional:    true,
								ElementType: types.StringType,
							},
							"test_set": schema.SetAttribute{
								Optional:    true,
								ElementType: types.StringType,
							},
							"test_list_nested": schema.ListNestedAttribute{
								Optional: true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"nested_str": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
							"test_map_nested": schema.MapNestedAttribute{
								Optional: true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"nested_str": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
							"test_set_nested": schema.SetNestedAttribute{
								Optional: true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"nested_str": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_int64":          tftypes.Number,
							"test_int64_computed": tftypes.Number,
							"test_int32":          tftypes.Number,
							"test_float64":        tftypes.Number,
							"test_float32":        tftypes.Number,
							"test_dynamic":        tftypes.DynamicPseudoType,
							"test_object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"obj_str": tftypes.String,
								},
							},
							"test_map":         tftypes.Map{ElementType: tftypes.String},
							"test_set":         tftypes.Set{ElementType: tftypes.String},
							"test_list_nested": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}},
							"test_map_nested":  tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}},
							"test_set_nested":  tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}},
						},
					}, map[string]tftypes.Value{
						"test_int64":          tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
						"test_int64_computed": tftypes.NewValue(tftypes.Number, nil),
						"test_int32":          tftypes.NewValue(tftypes.Number, big.NewFloat(99)),
						"test_float64":        tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
						"test_float32":        tftypes.NewValue(tftypes.Number, big.NewFloat(3.14)),
						"test_dynamic":        tftypes.NewValue(tftypes.String, "dynamic-val"),
						"test_object": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"obj_str": tftypes.String}}, map[string]tftypes.Value{
							"obj_str": tftypes.NewValue(tftypes.String, "obj-val"),
						}),
						"test_map": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
							"key1": tftypes.NewValue(tftypes.String, "val1"),
						}),
						"test_set": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "set-val"),
						}),
						"test_list_nested": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}, map[string]tftypes.Value{
								"nested_str": tftypes.NewValue(tftypes.String, "list-nested-val"),
							}),
						}),
						"test_map_nested": tftypes.NewValue(tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}}, map[string]tftypes.Value{
							"mk1": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}, map[string]tftypes.Value{
								"nested_str": tftypes.NewValue(tftypes.String, "map-nested-val"),
							}),
						}),
						"test_set_nested": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"nested_str": tftypes.String}}, map[string]tftypes.Value{
								"nested_str": tftypes.NewValue(tftypes.String, "set-nested-val"),
							}),
						}),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test_int64": schema.Int64Attribute{
								Optional: true,
							},
							"test_int64_computed": schema.Int64Attribute{
								Computed: true,
							},
							"test_int32": schema.Int32Attribute{
								Optional: true,
							},
							"test_float64": schema.Float64Attribute{
								Optional: true,
							},
							"test_float32": schema.Float32Attribute{
								Optional: true,
							},
							"test_dynamic": schema.DynamicAttribute{
								Optional: true,
							},
							"test_object": schema.ObjectAttribute{
								Optional: true,
								AttributeTypes: map[string]attr.Type{
									"obj_str": types.StringType,
								},
							},
							"test_map": schema.MapAttribute{
								Optional:    true,
								ElementType: types.StringType,
							},
							"test_set": schema.SetAttribute{
								Optional:    true,
								ElementType: types.StringType,
							},
							"test_list_nested": schema.ListNestedAttribute{
								Optional: true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"nested_str": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
							"test_map_nested": schema.MapNestedAttribute{
								Optional: true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"nested_str": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
							"test_set_nested": schema.SetNestedAttribute{
								Optional: true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"nested_str": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"response-set-nested-block": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Blocks: map[string]schema.Block{
						"test_set_block": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"set_block_attr": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_set_block": tftypes.Set{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"set_block_attr": tftypes.String,
								},
							}},
						},
					}, map[string]tftypes.Value{
						"test_set_block": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"set_block_attr": tftypes.String,
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"set_block_attr": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"set_block_attr": tftypes.NewValue(tftypes.String, "val-a"),
							}),
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"set_block_attr": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"set_block_attr": tftypes.NewValue(tftypes.String, "val-b"),
							}),
						}),
					}),
					Schema: schema.Schema{
						Blocks: map[string]schema.Block{
							"test_set_block": schema.SetNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"set_block_attr": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_set_block": tftypes.Set{ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"set_block_attr": tftypes.String,
								},
							}},
						},
					}, map[string]tftypes.Value{
						"test_set_block": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"set_block_attr": tftypes.String,
							},
						}}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"set_block_attr": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"set_block_attr": tftypes.NewValue(tftypes.String, "val-a"),
							}),
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"set_block_attr": tftypes.String,
								},
							}, map[string]tftypes.Value{
								"set_block_attr": tftypes.NewValue(tftypes.String, "val-b"),
							}),
						}),
					}),
					Schema: schema.Schema{
						Blocks: map[string]schema.Block{
							"test_set_block": schema.SetNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"set_block_attr": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"response-single-nested-block": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Blocks: map[string]schema.Block{
						"test_single_block": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"single_required": schema.StringAttribute{
									Required: true,
								},
								"single_computed": schema.StringAttribute{
									Computed: true,
								},
								"single_deprecated": schema.StringAttribute{
									Optional:           true,
									DeprecationMessage: "deprecated",
								},
							},
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_single_block": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_required":   tftypes.String,
									"single_computed":   tftypes.String,
									"single_deprecated": tftypes.String,
								},
							},
						},
					}, map[string]tftypes.Value{
						"test_single_block": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"single_required":   tftypes.String,
								"single_computed":   tftypes.String,
								"single_deprecated": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"single_required":   tftypes.NewValue(tftypes.String, "req-val"),
							"single_computed":   tftypes.NewValue(tftypes.String, "comp-val"),
							"single_deprecated": tftypes.NewValue(tftypes.String, "dep-val"),
						}),
					}),
					Schema: schema.Schema{
						Blocks: map[string]schema.Block{
							"test_single_block": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"single_required": schema.StringAttribute{
										Required: true,
									},
									"single_computed": schema.StringAttribute{
										Computed: true,
									},
									"single_deprecated": schema.StringAttribute{
										Optional:           true,
										DeprecationMessage: "deprecated",
									},
								},
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_single_block": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_required":   tftypes.String,
									"single_computed":   tftypes.String,
									"single_deprecated": tftypes.String,
								},
							},
						},
					}, map[string]tftypes.Value{
						"test_single_block": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"single_required":   tftypes.String,
								"single_computed":   tftypes.String,
								"single_deprecated": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"single_required":   tftypes.NewValue(tftypes.String, "req-val"),
							"single_computed":   tftypes.NewValue(tftypes.String, nil),
							"single_deprecated": tftypes.NewValue(tftypes.String, nil),
						}),
					}),
					Schema: schema.Schema{
						Blocks: map[string]schema.Block{
							"test_single_block": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"single_required": schema.StringAttribute{
										Required: true,
									},
									"single_computed": schema.StringAttribute{
										Computed: true,
									},
									"single_deprecated": schema.StringAttribute{
										Optional:           true,
										DeprecationMessage: "deprecated",
									},
								},
							},
						},
					},
				},
			},
		},
		"response-timeouts-and-optional-number": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"test_optional_number": schema.NumberAttribute{
							Optional: true,
						},
						"test_nonzero_number": schema.NumberAttribute{
							Optional: true,
						},
						"timeouts": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"create": schema.StringAttribute{
									Optional: true,
								},
							},
						},
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_optional_number": tftypes.Number,
							"test_nonzero_number":  tftypes.Number,
							"timeouts": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"create": tftypes.String,
								},
							},
						},
					}, map[string]tftypes.Value{
						"test_optional_number": tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
						"test_nonzero_number":  tftypes.NewValue(tftypes.Number, big.NewFloat(7)),
						"timeouts": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"create": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"create": tftypes.NewValue(tftypes.String, "30m"),
						}),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test_optional_number": schema.NumberAttribute{
								Optional: true,
							},
							"test_nonzero_number": schema.NumberAttribute{
								Optional: true,
							},
							"timeouts": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"create": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_optional_number": tftypes.Number,
							"test_nonzero_number":  tftypes.Number,
							"timeouts": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"create": tftypes.String,
								},
							},
						},
					}, map[string]tftypes.Value{
						"test_optional_number": tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
						"test_nonzero_number":  tftypes.NewValue(tftypes.Number, big.NewFloat(7)),
						"timeouts": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"create": tftypes.String,
							},
						}, nil),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test_optional_number": schema.NumberAttribute{
								Optional: true,
							},
							"test_nonzero_number": schema.NumberAttribute{
								Optional: true,
							},
							"timeouts": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"create": schema.StringAttribute{
										Optional: true,
									},
								},
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

			response := &fwserver.GenerateResourceConfigResponse{}
			testCase.server.GenerateResourceConfig(t.Context(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

type testConflictsWithStringValidator struct {
	paths path.Expressions
}

func (v testConflictsWithStringValidator) Description(context.Context) string {
	return ""
}

func (v testConflictsWithStringValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testConflictsWithStringValidator) ValidateString(context.Context, schemavalidator.StringRequest, *schemavalidator.StringResponse) {
}

func (v testConflictsWithStringValidator) ConflictsWithPaths() path.Expressions {
	return v.paths
}

type testConflictsWithListValidator struct {
	paths path.Expressions
}

func (v testConflictsWithListValidator) Description(context.Context) string {
	return ""
}

func (v testConflictsWithListValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testConflictsWithListValidator) ValidateList(context.Context, schemavalidator.ListRequest, *schemavalidator.ListResponse) {
}

func (v testConflictsWithListValidator) ConflictsWithPaths() path.Expressions {
	return v.paths
}

type testConflictsWithObjectValidator struct {
	paths path.Expressions
}

func (v testConflictsWithObjectValidator) Description(context.Context) string {
	return ""
}

func (v testConflictsWithObjectValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testConflictsWithObjectValidator) ValidateObject(context.Context, schemavalidator.ObjectRequest, *schemavalidator.ObjectResponse) {
}

func (v testConflictsWithObjectValidator) ConflictsWithPaths() path.Expressions {
	return v.paths
}

type testExactlyOneOfStringValidator struct {
	paths path.Expressions
}

func (v testExactlyOneOfStringValidator) Description(context.Context) string {
	return ""
}

func (v testExactlyOneOfStringValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testExactlyOneOfStringValidator) ValidateString(context.Context, schemavalidator.StringRequest, *schemavalidator.StringResponse) {
}

func (v testExactlyOneOfStringValidator) ExactlyOneOfPaths() path.Expressions {
	return v.paths
}

type testAlsoRequiresStringValidator struct {
	paths path.Expressions
}

func (v testAlsoRequiresStringValidator) Description(context.Context) string {
	return ""
}

func (v testAlsoRequiresStringValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testAlsoRequiresStringValidator) ValidateString(context.Context, schemavalidator.StringRequest, *schemavalidator.StringResponse) {
}

func (v testAlsoRequiresStringValidator) AlsoRequiresPaths() path.Expressions {
	return v.paths
}

type testResourceConflictsWithValidator struct {
	testprovider.ResourceConfigValidator
	paths path.Expressions
}

func (v *testResourceConflictsWithValidator) ConflictsWithPaths() path.Expressions { return v.paths }

type testResourceExactlyOneOfValidator struct {
	testprovider.ResourceConfigValidator
	paths path.Expressions
}

func (v *testResourceExactlyOneOfValidator) ExactlyOneOfPaths() path.Expressions { return v.paths }

type testResourceAlsoRequiresValidator struct {
	testprovider.ResourceConfigValidator
	paths path.Expressions
}

func (v *testResourceAlsoRequiresValidator) AlsoRequiresPaths() path.Expressions { return v.paths }

var _ schemavalidator.String = testConflictsWithStringValidator{}
var _ schemavalidator.ConflictsWithValidator = testConflictsWithStringValidator{}
var _ schemavalidator.List = testConflictsWithListValidator{}
var _ schemavalidator.ConflictsWithValidator = testConflictsWithListValidator{}
var _ schemavalidator.Object = testConflictsWithObjectValidator{}
var _ schemavalidator.ConflictsWithValidator = testConflictsWithObjectValidator{}
var _ schemavalidator.String = testExactlyOneOfStringValidator{}
var _ schemavalidator.ExactlyOneOfValidator = testExactlyOneOfStringValidator{}
var _ schemavalidator.String = testAlsoRequiresStringValidator{}
var _ schemavalidator.AlsoRequiresValidator = testAlsoRequiresStringValidator{}
var _ resource.ConfigValidator = &testResourceConflictsWithValidator{}
var _ resource.ConfigValidator = &testResourceExactlyOneOfValidator{}
var _ resource.ConfigValidator = &testResourceAlsoRequiresValidator{}
var _ schemavalidator.ConflictsWithValidator = &testResourceConflictsWithValidator{}
var _ schemavalidator.ExactlyOneOfValidator = &testResourceExactlyOneOfValidator{}
var _ schemavalidator.AlsoRequiresValidator = &testResourceAlsoRequiresValidator{}
