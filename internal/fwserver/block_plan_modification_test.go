package fwserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	modifyAttributePlanRequest := func(attrPath path.Path, schema tfsdk.Schema, values modifyAttributePlanValues) tfsdk.ModifyAttributePlanRequest {
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

	modifyAttributePlanWithPrivateRequest := func(attrPath path.Path, schema tfsdk.Schema, values modifyAttributePlanValues, privateProviderData *privatestate.ProviderData) tfsdk.ModifyAttributePlanRequest {
		req := modifyAttributePlanRequest(attrPath, schema, values)
		req.Private = privateProviderData

		return req
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testCases := map[string]struct {
		req          tfsdk.ModifyAttributePlanRequest
		resp         ModifySchemaPlanResponse // Plan automatically copied from req
		expectedResp ModifySchemaPlanResponse
	}{
		"no-plan-modifiers": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
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
				path.Root("test"),
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
				Private: testEmptyProviderData,
			},
		},
		"block-request-private": {
			req: modifyAttributePlanWithPrivateRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					testBlockPlanModifierPrivateGet{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
				testProviderData,
			),
			resp: ModifySchemaPlanResponse{},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: schemaTfValue("TESTATTRONE"),
					Schema: schema([]tfsdk.AttributePlanModifier{
						testBlockPlanModifierPrivateGet{},
					}, nil),
				},
				Private: testProviderData,
			},
		},
		"block-response-private": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					testBlockPlanModifierPrivateSet{},
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
					Raw: schemaTfValue("TESTATTRONE"),
					Schema: schema([]tfsdk.AttributePlanModifier{
						testBlockPlanModifierPrivateSet{},
					}, nil),
				},
				Private: testProviderData,
			},
		},
		"block-modified-previous-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
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
				Private: testEmptyProviderData,
			},
		},
		"block-requires-replacement": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
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
						resource.RequiresReplace(),
					}, nil),
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-requires-replacement-previous-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
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
						resource.RequiresReplace(),
					}, nil),
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-requires-replacement-passthrough": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
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
						resource.RequiresReplace(),
						testBlockPlanModifierNullList{},
					}, nil),
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-requires-replacement-unset": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					planmodifiers.TestRequiresReplaceFalseModifier{},
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
						resource.RequiresReplace(),
						planmodifiers.TestRequiresReplaceFalseModifier{},
					}, nil),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-warnings": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
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
						planmodifiers.TestWarningDiagModifier{},
						planmodifiers.TestWarningDiagModifier{},
					}, nil),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-warnings-previous-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
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
						planmodifiers.TestWarningDiagModifier{},
						planmodifiers.TestWarningDiagModifier{},
					}, nil),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
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
						planmodifiers.TestErrorDiagModifier{},
						planmodifiers.TestErrorDiagModifier{},
					}, nil),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-error-previous-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
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
						planmodifiers.TestErrorDiagModifier{},
						planmodifiers.TestErrorDiagModifier{},
					}, nil),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-modified": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanValueModifierOne{},
					planmodifiers.TestAttrPlanValueModifierTwo{},
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
						planmodifiers.TestAttrPlanValueModifierOne{},
						planmodifiers.TestAttrPlanValueModifierTwo{},
					}),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-modified-previous-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanValueModifierOne{},
					planmodifiers.TestAttrPlanValueModifierTwo{},
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
						planmodifiers.TestAttrPlanValueModifierOne{},
						planmodifiers.TestAttrPlanValueModifierTwo{},
					}),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-requires-replacement": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
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
						resource.RequiresReplace(),
					}),
				},
				RequiresReplace: path.Paths{
					path.Root("test").AtListIndex(0).AtName("nested_attr"),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-requires-replacement-previous-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
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
						resource.RequiresReplace(),
					}),
				},
				RequiresReplace: path.Paths{
					path.Root("test").AtListIndex(0).AtName("nested_attr"),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-requires-replacement-passthrough": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					planmodifiers.TestAttrPlanValueModifierOne{},
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
						resource.RequiresReplace(),
						planmodifiers.TestAttrPlanValueModifierOne{},
					}),
				},
				RequiresReplace: path.Paths{
					path.Root("test").AtListIndex(0).AtName("nested_attr"),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-requires-replacement-unset": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					planmodifiers.TestRequiresReplaceFalseModifier{},
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
						resource.RequiresReplace(),
						planmodifiers.TestRequiresReplaceFalseModifier{},
					}),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-warnings": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
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
						planmodifiers.TestWarningDiagModifier{},
						planmodifiers.TestWarningDiagModifier{},
					}),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-warnings-previous-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
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
						planmodifiers.TestWarningDiagModifier{},
						planmodifiers.TestWarningDiagModifier{},
					}),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
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
						planmodifiers.TestErrorDiagModifier{},
						planmodifiers.TestErrorDiagModifier{},
					}),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-error-previous-error": {
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
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
						planmodifiers.TestErrorDiagModifier{},
						planmodifiers.TestErrorDiagModifier{},
					}),
				},
				Private: testEmptyProviderData,
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			block, ok := tc.req.Config.Schema.Blocks["test"]

			if !ok {
				t.Fatalf("Unexpected error getting schema block")
			}

			tc.resp.Plan = tc.req.Plan

			BlockModifyPlan(context.Background(), block, tc.req, &tc.resp)

			if diff := cmp.Diff(tc.expectedResp, tc.resp, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
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

type testBlockPlanModifierPrivateGet struct{}

func (t testBlockPlanModifierPrivateGet) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	expected := `{"pKeyOne": {"k0": "zero", "k1": 1}}`

	key := "providerKeyOne"
	got, diags := req.Private.GetKey(ctx, key)

	resp.Diagnostics.Append(diags...)

	if string(got) != expected {
		resp.Diagnostics.AddError("unexpected req.Private.Provider value: %s", string(got))
	}
}

func (t testBlockPlanModifierPrivateGet) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testBlockPlanModifierPrivateGet) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type testBlockPlanModifierPrivateSet struct{}

func (t testBlockPlanModifierPrivateSet) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

	resp.Diagnostics.Append(diags...)
}

func (t testBlockPlanModifierPrivateSet) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testBlockPlanModifierPrivateSet) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}
