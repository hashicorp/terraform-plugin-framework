package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBlockModifyPlan(t *testing.T) {
	t.Parallel()

	schema := func(blockPlanModifiers tfsdk.AttributePlanModifiers, nestedAttrPlanModifiers tfsdk.AttributePlanModifiers) tfsdk.Schema {
		return tfsdk.Schema{
			Blocks: map[string]tfsdk.Block{
				"test": {
					Attributes: map[string]tfsdk.Attribute{
						"nested_attr": {
							Type:          types.StringType,
							Required:      true,
							PlanModifiers: nestedAttrPlanModifiers,
						},
					},
					NestingMode:   tfsdk.BlockNestingModeList,
					PlanModifiers: blockPlanModifiers,
				},
			},
		}
	}

	schemaTfValue := func(nestedAttrValue string) tftypes.Value {
		return tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested_attr": tftypes.String,
							},
						},
					},
				},
			},
			map[string]tftypes.Value{
				"test": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested_attr": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_attr": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_attr": tftypes.NewValue(tftypes.String, nestedAttrValue),
							},
						),
					},
				),
			},
		)
	}

	var schemaNullTfValue tftypes.Value = tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"test": tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested_attr": tftypes.String,
						},
					},
				},
			},
		},
		map[string]tftypes.Value{
			"test": tftypes.NewValue(
				tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested_attr": tftypes.String,
						},
					},
				},
				nil,
			),
		},
	)

	type modifyAttributePlanValues struct {
		config string
		plan   string
		state  string
	}

	modifyAttributePlanRequest := func(attrPath *tftypes.AttributePath, schema tfsdk.Schema, values modifyAttributePlanValues) tfsdk.ModifyAttributePlanRequest {
		return tfsdk.ModifyAttributePlanRequest{
			AttributePath: attrPath,
			Config: tfsdk.Config{
				Raw:    schemaTfValue(values.config),
				Schema: schema,
			},
			Plan: tfsdk.Plan{
				Raw:    schemaTfValue(values.plan),
				Schema: schema,
			},
			State: tfsdk.State{
				Raw:    schemaTfValue(values.state),
				Schema: schema,
			},
		}
	}

	testCases := map[string]struct {
		req          tfsdk.ModifyAttributePlanRequest
		resp         ModifySchemaPlanResponse // Plan automatically copied from req
		expectedResp ModifySchemaPlanResponse
	}{
		"no-plan-modifiers": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, nil),
				modifyAttributePlanValues{
					config: "testvalue",
					plan:   "testvalue",
					state:  "testvalue",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw:    schemaTfValue("testvalue"),
					Schema: schema(nil, nil),
				},
			},
		},
		"block-modified": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					testBlockPlanModifierNullList{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: schemaNullTfValue,
					Schema: schema([]tfsdk.AttributePlanModifier{
						testBlockPlanModifierNullList{},
					}, nil),
				},
			},
		},
		"block-modified-previous-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					testBlockPlanModifierNullList{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
			),
			resp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaNullTfValue,
					Schema: schema([]tfsdk.AttributePlanModifier{
						testBlockPlanModifierNullList{},
					}, nil),
				},
			},
		},
		"block-requires-replacement": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				}, nil),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("newtestvalue"),
					Schema: schema([]tfsdk.AttributePlanModifier{
						tfsdk.RequiresReplace(),
					}, nil),
				},
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("test"),
				},
			},
		},
		"block-requires-replacement-previous-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				}, nil),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			resp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("newtestvalue"),
					Schema: schema([]tfsdk.AttributePlanModifier{
						tfsdk.RequiresReplace(),
					}, nil),
				},
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("test"),
				},
			},
		},
		"block-requires-replacement-passthrough": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					testBlockPlanModifierNullList{},
				}, nil),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: schemaNullTfValue,
					Schema: schema([]tfsdk.AttributePlanModifier{
						tfsdk.RequiresReplace(),
						testBlockPlanModifierNullList{},
					}, nil),
				},
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("test"),
				},
			},
		},
		"block-requires-replacement-unset": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					testRequiresReplaceFalseModifier{},
				}, nil),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("newtestvalue"),
					Schema: schema([]tfsdk.AttributePlanModifier{
						tfsdk.RequiresReplace(),
						testRequiresReplaceFalseModifier{},
					}, nil),
				},
			},
		},
		"block-warnings": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					testWarningDiagModifier{},
					testWarningDiagModifier{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTDIAG"),
					Schema: schema([]tfsdk.AttributePlanModifier{
						testWarningDiagModifier{},
						testWarningDiagModifier{},
					}, nil),
				},
			},
		},
		"block-warnings-previous-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					testWarningDiagModifier{},
					testWarningDiagModifier{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			resp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTDIAG"),
					Schema: schema([]tfsdk.AttributePlanModifier{
						testWarningDiagModifier{},
						testWarningDiagModifier{},
					}, nil),
				},
			},
		},
		"block-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					testErrorDiagModifier{},
					testErrorDiagModifier{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTDIAG"),
					Schema: schema([]tfsdk.AttributePlanModifier{
						testErrorDiagModifier{},
						testErrorDiagModifier{},
					}, nil),
				},
			},
		},
		"block-error-previous-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema([]tfsdk.AttributePlanModifier{
					testErrorDiagModifier{},
					testErrorDiagModifier{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			resp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTDIAG"),
					Schema: schema([]tfsdk.AttributePlanModifier{
						testErrorDiagModifier{},
						testErrorDiagModifier{},
					}, nil),
				},
			},
		},
		"nested-attribute-modified": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					testAttrPlanValueModifierOne{},
					testAttrPlanValueModifierTwo{},
				}),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("MODIFIED_TWO"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						testAttrPlanValueModifierOne{},
						testAttrPlanValueModifierTwo{},
					}),
				},
			},
		},
		"nested-attribute-modified-previous-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					testAttrPlanValueModifierOne{},
					testAttrPlanValueModifierTwo{},
				}),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
			),
			resp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("MODIFIED_TWO"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						testAttrPlanValueModifierOne{},
						testAttrPlanValueModifierTwo{},
					}),
				},
			},
		},
		"nested-attribute-requires-replacement": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				}),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("newtestvalue"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						tfsdk.RequiresReplace(),
					}),
				},
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("nested_attr"),
				},
			},
		},
		"nested-attribute-requires-replacement-previous-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				}),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			resp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("newtestvalue"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						tfsdk.RequiresReplace(),
					}),
				},
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("nested_attr"),
				},
			},
		},
		"nested-attribute-requires-replacement-passthrough": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					testAttrPlanValueModifierOne{},
				}),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "previousvalue",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTATTRTWO"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						tfsdk.RequiresReplace(),
						testAttrPlanValueModifierOne{},
					}),
				},
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("nested_attr"),
				},
			},
		},
		"nested-attribute-requires-replacement-unset": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					testRequiresReplaceFalseModifier{},
				}),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("newtestvalue"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						tfsdk.RequiresReplace(),
						testRequiresReplaceFalseModifier{},
					}),
				},
			},
		},
		"nested-attribute-warnings": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					testWarningDiagModifier{},
					testWarningDiagModifier{},
				}),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTDIAG"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						testWarningDiagModifier{},
						testWarningDiagModifier{},
					}),
				},
			},
		},
		"nested-attribute-warnings-previous-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					testWarningDiagModifier{},
					testWarningDiagModifier{},
				}),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			resp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTDIAG"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						testWarningDiagModifier{},
						testWarningDiagModifier{},
					}),
				},
			},
		},
		"nested-attribute-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					testErrorDiagModifier{},
					testErrorDiagModifier{},
				}),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTDIAG"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						testErrorDiagModifier{},
						testErrorDiagModifier{},
					}),
				},
			},
		},
		"nested-attribute-error-previous-error": {
			req: modifyAttributePlanRequest(
				tftypes.NewAttributePath().WithAttributeName("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					testErrorDiagModifier{},
					testErrorDiagModifier{},
				}),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			resp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTDIAG"),
					Schema: schema(nil, []tfsdk.AttributePlanModifier{
						testErrorDiagModifier{},
						testErrorDiagModifier{},
					}),
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			block, err := fwserver.SchemaBlockAtPath(tc.req.Config.Schema, tc.req.AttributePath)

			if err != nil {
				t.Fatalf("Unexpected error getting %s", err)
			}

			tc.resp.Plan = tc.req.Plan

			BlockModifyPlan(context.Background(), block, tc.req, &tc.resp)

			if diff := cmp.Diff(tc.expectedResp, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

type testBlockPlanModifierNullList struct{}

func (t testBlockPlanModifierNullList) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	_, ok := req.AttributePlan.(types.List)
	if !ok {
		return
	}

	resp.AttributePlan = types.List{
		ElemType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"nested_attr": types.StringType,
			},
		},
		Null: true,
	}
}

func (t testBlockPlanModifierNullList) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testBlockPlanModifierNullList) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}
