// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/planmodifierdiag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBlockModifyPlan(t *testing.T) {
	t.Parallel()

	schema := func(blockPlanModifiers []planmodifier.List, nestedAttrPlanModifiers []planmodifier.String) testschema.Schema {
		return testschema.Schema{
			Blocks: map[string]fwschema.Block{
				"test": testschema.BlockWithListPlanModifiers{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required:      true,
							PlanModifiers: nestedAttrPlanModifiers,
						},
					},
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

	type modifyAttributePlanValues struct {
		config string
		plan   string
		state  string
	}

	modifyAttributePlanRequest := func(attrPath path.Path, schema fwschema.Schema, values modifyAttributePlanValues) ModifyAttributePlanRequest {
		return ModifyAttributePlanRequest{
			AttributeConfig: types.ListValueMust(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
				},
				[]attr.Value{
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						map[string]attr.Value{
							"nested_attr": types.StringValue(values.config),
						},
					),
				},
			),
			AttributePath: attrPath,
			AttributePlan: types.ListValueMust(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
				},
				[]attr.Value{
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						map[string]attr.Value{
							"nested_attr": types.StringValue(values.plan),
						},
					),
				},
			),
			AttributeState: types.ListValueMust(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
				},
				[]attr.Value{
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						map[string]attr.Value{
							"nested_attr": types.StringValue(values.state),
						},
					),
				},
			),
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

	modifyAttributePlanWithPrivateRequest := func(attrPath path.Path, schema fwschema.Schema, values modifyAttributePlanValues, privateProviderData *privatestate.ProviderData) ModifyAttributePlanRequest {
		req := modifyAttributePlanRequest(attrPath, schema, values)
		req.Private = privateProviderData

		return req
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testCases := map[string]struct {
		block        fwschema.Block
		req          ModifyAttributePlanRequest
		expectedResp ModifyAttributePlanResponse
	}{
		"no-plan-modifiers": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, nil),
				modifyAttributePlanValues{
					config: "testvalue",
					plan:   "testvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
			},
		},
		"block-modified": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.List{
					testBlockPlanModifierNullList{},
				},
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]planmodifier.List{
					testBlockPlanModifierNullList{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				),
			},
		},
		"block-request-private": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.List{
					testBlockPlanModifierPrivateGet{},
				},
			},
			req: modifyAttributePlanWithPrivateRequest(
				path.Root("test"),
				schema([]planmodifier.List{
					testBlockPlanModifierPrivateGet{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
				testProviderData,
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("TESTATTRONE"),
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-response-private": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.List{
					testBlockPlanModifierPrivateSet{},
				},
			},
			req: modifyAttributePlanWithPrivateRequest(
				path.Root("test"),
				schema([]planmodifier.List{
					testBlockPlanModifierPrivateSet{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
				privatestate.EmptyProviderData(context.Background()),
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("TESTATTRONE"),
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-list-null-plan": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				),
				AttributeState: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				),
				Private: testProviderData,
			},
		},
		"block-list-null-state": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributeState: types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-list-nested-private": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributeState: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-list-usestateforunknown": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_computed": testschema.Attribute{
						Type:     types.StringType,
						Computed: true,
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ListUnknown(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
				),
				AttributeState: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
							},
						),
					},
				),
			},
		},
		"block-set-null-plan": {
			block: testschema.BlockWithSetPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				PlanModifiers: []planmodifier.Set{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.SetNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				),
				AttributeState: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				),
				Private: testProviderData,
			},
		},
		"block-set-null-state": {
			block: testschema.BlockWithSetPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				PlanModifiers: []planmodifier.Set{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributeState: types.SetNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-set-nested-private": {
			block: testschema.BlockWithSetPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				PlanModifiers: []planmodifier.Set{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributeState: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-list-nested-block-list": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"id": testschema.AttributeWithStringPlanModifiers{
							Computed: true,
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_computed": testschema.AttributeWithStringPlanModifiers{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"nested_required": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id": types.StringType,
							"list": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_computed": types.StringType,
										"nested_required": types.StringType,
									},
								},
							},
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"id": types.StringType,
								"list": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
								},
							},
							map[string]attr.Value{
								"id": types.StringValue("configvalue"),
								"list": types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
											map[string]attr.Value{
												"nested_computed": types.StringNull(),
												"nested_required": types.StringValue("configvalue"),
											},
										),
									},
								),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id": types.StringType,
							"list": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_computed": types.StringType,
										"nested_required": types.StringType,
									},
								},
							},
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"id": types.StringType,
								"list": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
								},
							},
							map[string]attr.Value{
								"id": types.StringValue("one"),
								"list": types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
											map[string]attr.Value{
												"nested_computed": types.StringUnknown(),
												"nested_required": types.StringValue("configvalue"),
											},
										),
									},
								),
							},
						),
					},
				),
				AttributeState: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id": types.StringType,
							"list": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_computed": types.StringType,
										"nested_required": types.StringType,
									},
								},
							},
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"id": types.StringType,
								"list": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
								},
							},
							map[string]attr.Value{
								"id": types.StringValue("one"),
								"list": types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
											map[string]attr.Value{
												"nested_computed": types.StringValue("statevalue"),
												"nested_required": types.StringValue("configvalue"),
											},
										),
									},
								),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"id": types.StringType,
							"list": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_computed": types.StringType,
										"nested_required": types.StringType,
									},
								},
							},
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"id": types.StringType,
								"list": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
								},
							},
							map[string]attr.Value{
								"id": types.StringValue("one"),
								"list": types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
											map[string]attr.Value{
												"nested_computed": types.StringUnknown(),
												"nested_required": types.StringValue("configvalue"),
											},
										),
									},
								),
							},
						),
					},
				),
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtListIndex(0).AtName("id"),
					),
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtListIndex(0).AtName("list").AtListIndex(0).AtName("nested_computed"),
					),
				},
			},
		},
		"block-list-nested-usestateforunknown-elements-rearranged": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"nested_required": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},

				NestingMode: fwschema.BlockNestingModeList,
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue2"), // prior state on index 0 is testvalue1
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
					},
				),
				AttributeState: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue2"),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
					},
				),
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtListIndex(0).AtName("nested_computed"),
					),
				},
			},
		},
		"block-list-nested-usestateforunknown-elements-removed": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"nested_required": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},

				NestingMode: fwschema.BlockNestingModeList,
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue2"), // prior state on index 0 is testvalue1
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				AttributeState: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue2"),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtListIndex(0).AtName("nested_computed"),
					),
				},
			},
		},
		"block-set-nested-usestateforunknown": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"nested_required": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},

				NestingMode: fwschema.BlockNestingModeSet,
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				AttributeState: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue2"),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtSetValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								map[string]attr.Value{
									"nested_computed": types.StringUnknown(),
									"nested_required": types.StringValue("testvalue1"),
								},
							),
						).AtName("nested_computed"),
					),
				},
			},
		},
		"block-set-nested-usestateforunknown-elements-rearranged": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"nested_required": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},

				NestingMode: fwschema.BlockNestingModeSet,
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue2"), // prior state on index 0 is testvalue1
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
					},
				),
				AttributeState: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue2"),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
					},
				),
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtSetValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								map[string]attr.Value{
									"nested_computed": types.StringUnknown(),
									"nested_required": types.StringValue("testvalue2"),
								},
							),
						).AtName("nested_computed"),
					),
				},
			},
		},
		"block-set-nested-usestateforunknown-elements-removed": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"nested_required": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},

				NestingMode: fwschema.BlockNestingModeSet,
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringNull(),
								"nested_required": types.StringValue("testvalue2"), // prior state on index 0 is testvalue1
							},
						),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				AttributeState: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
								"nested_required": types.StringValue("testvalue1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue2"),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringUnknown(),
								"nested_required": types.StringValue("testvalue2"),
							},
						),
					},
				),
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtSetValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								map[string]attr.Value{
									"nested_computed": types.StringUnknown(),
									"nested_required": types.StringValue("testvalue2"),
								},
							),
						).AtName("nested_computed"),
					),
				},
			},
		},
		"block-set-usestateforunknown": {
			block: testschema.BlockWithSetPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_computed": testschema.Attribute{
						Type:     types.StringType,
						Computed: true,
					},
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.SetNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.SetUnknown(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
				),
				AttributeState: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
							},
						),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.StringValue("statevalue1"),
							},
						),
					},
				),
			},
		},
		"block-single-null-plan": {
			block: testschema.BlockWithObjectPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.StringValue("testvalue"),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ObjectNull(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.StringValue("testvalue"),
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectNull(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
				),
				Private: testProviderData,
			},
		},
		"block-single-null-state": {
			block: testschema.BlockWithObjectPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.StringValue("testvalue"),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectNull(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.StringValue("testvalue"),
					},
				),
				Private: testProviderData,
			},
		},
		"block-single-nested-private": {
			block: testschema.BlockWithObjectPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.StringValue("testvalue"),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.StringValue("testvalue"),
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.StringValue("testvalue"),
					},
				),
				Private: testProviderData,
			},
		},
		"block-single-nested-usestateforunknown": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"nested_required": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeSingle,
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_computed": types.StringType,
						"nested_required": types.StringType,
					},
					map[string]attr.Value{
						"nested_computed": types.StringNull(),
						"nested_required": types.StringValue("testvalue"),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_computed": types.StringType,
						"nested_required": types.StringType,
					},
					map[string]attr.Value{
						"nested_computed": types.StringUnknown(),
						"nested_required": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_computed": types.StringType,
						"nested_required": types.StringType,
					},
					map[string]attr.Value{
						"nested_computed": types.StringValue("statevalue"),
						"nested_required": types.StringValue("testvalue"),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_computed": types.StringType,
						"nested_required": types.StringType,
					},
					map[string]attr.Value{
						"nested_computed": types.StringValue("statevalue"),
						"nested_required": types.StringValue("testvalue"),
					},
				),
			},
		},
		"block-single-usestateforunknown": {
			block: testschema.BlockWithObjectPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_computed": testschema.Attribute{
						Type:     types.StringType,
						Computed: true,
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ObjectNull(
					map[string]attr.Type{
						"nested_computed": types.StringType,
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ObjectUnknown(
					map[string]attr.Type{
						"nested_computed": types.StringType,
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_computed": types.StringType,
					},
					map[string]attr.Value{
						"nested_computed": types.StringValue("statevalue1"),
					},
				),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_computed": types.StringType,
					},
					map[string]attr.Value{
						"nested_computed": types.StringValue("statevalue1"),
					},
				),
			},
		},
		"block-requires-replacement": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = true
						},
					},
				}, nil),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("newtestvalue"),
							},
						),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"block-requires-replacement-passthrough": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = true
						},
					},
					testBlockPlanModifierNullList{},
				},
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = true
						},
					},
					testBlockPlanModifierNullList{},
				}, nil),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"block-requires-replacement-unset": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = true
						},
					},
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = false
						},
					},
				},
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = true
						},
					},
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = false
						},
					},
				}, nil),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("newtestvalue"),
							},
						),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"block-warnings": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
						},
					},
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
						},
					},
				},
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
						},
					},
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
						},
					},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				Diagnostics: diag.Diagnostics{
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("TESTDIAG"),
							},
						),
					},
				),
			},
		},
		"block-error": {
			block: testschema.BlockWithListPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"nested_attr": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.AddError("Error diag", "This is an error")
						},
					},
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.AddError("Error diag", "This is an error")
						},
					},
				},
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema([]planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.AddError("Error diag", "This is an error")
						},
					},
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.AddError("Error diag", "This is an error")
						},
					},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("TESTDIAG"),
							},
						),
					},
				),
			},
		},
		"nested-attribute-modified": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										if req.PlanValue.ValueString() == "TESTATTRONE" {
											resp.PlanValue = types.StringValue("TESTATTRTWO")
										}
									},
								},
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										if req.PlanValue.ValueString() == "TESTATTRTWO" {
											resp.PlanValue = types.StringValue("MODIFIED_TWO")
										}
									},
								},
							},
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							if req.PlanValue.ValueString() == "TESTATTRONE" {
								resp.PlanValue = types.StringValue("TESTATTRTWO")
							}
						},
					},
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							if req.PlanValue.ValueString() == "TESTATTRTWO" {
								resp.PlanValue = types.StringValue("MODIFIED_TWO")
							}
						},
					},
				}),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("MODIFIED_TWO"),
							},
						),
					},
				),
			},
		},
		"nested-attribute-requires-replacement": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										resp.RequiresReplace = true
									},
								},
							},
						},
					},
				},

				NestingMode: fwschema.BlockNestingModeList,
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.RequiresReplace = true
						},
					},
				}),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("newtestvalue"),
							},
						),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test").AtListIndex(0).AtName("nested_attr"),
				},
			},
		},
		"nested-attribute-requires-replacement-passthrough": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										resp.RequiresReplace = true
									},
								},
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										if req.PlanValue.ValueString() == "TESTATTRONE" {
											resp.PlanValue = types.StringValue("TESTATTRTWO")
										}
									},
								},
							},
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.RequiresReplace = true
						},
					},
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							if req.PlanValue.ValueString() == "TESTATTRONE" {
								resp.PlanValue = types.StringValue("TESTATTRTWO")
							}
						},
					},
				}),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "previousvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("TESTATTRTWO"),
							},
						),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test").AtListIndex(0).AtName("nested_attr"),
				},
			},
		},
		"nested-attribute-requires-replacement-unset": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										resp.RequiresReplace = true
									},
								},
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										resp.RequiresReplace = false
									},
								},
							},
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.RequiresReplace = true
						},
					},
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.RequiresReplace = false
						},
					},
				}),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("newtestvalue"),
							},
						),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test").AtListIndex(0).AtName("nested_attr"),
				},
			},
		},
		"nested-attribute-warnings": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
									},
								},
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
									},
								},
							},
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
						},
					},
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
						},
					},
				}),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				Diagnostics: diag.Diagnostics{
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("TESTDIAG"),
							},
						),
					},
				),
			},
		},
		"nested-attribute-error": {
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										resp.Diagnostics.AddError("Error diag", "This is an error")
									},
								},
								testplanmodifier.String{
									PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
										resp.Diagnostics.AddError("Error diag", "This is an error")
									},
								},
							},
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
			},
			req: modifyAttributePlanRequest(
				path.Root("test"),
				schema(nil, []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.Diagnostics.AddError("Error diag", "This is an error")
						},
					},
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.Diagnostics.AddError("Error diag", "This is an error")
						},
					},
				}),
				modifyAttributePlanValues{
					config: "TESTDIAG",
					plan:   "TESTDIAG",
					state:  "TESTDIAG",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				AttributePlan: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("TESTDIAG"),
							},
						),
					},
				),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := ModifyAttributePlanResponse{
				AttributePlan: tc.req.AttributePlan,
				Private:       tc.req.Private,
			}

			BlockModifyPlan(context.Background(), tc.block, tc.req, &got)

			if diff := cmp.Diff(tc.expectedResp, got, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				for _, d := range got.Diagnostics {
					t.Logf("%s: %s\n%s\n", d.Severity(), d.Summary(), d.Detail())
				}
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestBlockPlanModifyList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    fwxschema.BlockWithListPlanModifiers
		request  ModifyAttributePlanRequest
		response *ModifyAttributePlanResponse
		expected *ModifyAttributePlanResponse
	}{
		"request-path": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ListRequest.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-pathexpression": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ListRequest.PathExpression",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig:         types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:          types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-config": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.List{ElementType: tftypes.String},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.List{ElementType: tftypes.String},
											[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected ListRequest.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.List{ElementType: tftypes.String},
								[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-configvalue": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							got := req.ConfigValue
							expected := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")})

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ListRequest.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.ListNull(types.StringType),
				AttributeState:  types.ListNull(types.StringType),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListNull(types.StringType),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListNull(types.StringType),
			},
		},
		"request-plan": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							got := req.Plan
							expected := tfsdk.Plan{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.List{ElementType: tftypes.String},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.List{ElementType: tftypes.String},
											[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected ListRequest.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.List{ElementType: tftypes.String},
								[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-planvalue": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							got := req.PlanValue
							expected := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")})

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ListRequest.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListNull(types.StringType),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListNull(types.StringType),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-private": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected ListRequest.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListNull(types.StringType),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListNull(types.StringType),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
		},
		"request-state": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							got := req.State
							expected := tfsdk.State{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.List{ElementType: tftypes.String},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.List{ElementType: tftypes.String},
											[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected ListRequest.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.List{ElementType: tftypes.String},
								[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-statevalue": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							got := req.StateValue
							expected := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")})

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ListRequest.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListNull(types.StringType),
				AttributePlan:   types.ListNull(types.StringType),
				AttributeState:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListNull(types.StringType),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListNull(types.StringType),
			},
		},
		"response-diagnostics": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("other"),
						"Existing Warning Summary",
						"Existing Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("other"),
						"Existing Error Summary",
						"Existing Error Details",
					),
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("other"),
						"Existing Warning Summary",
						"Existing Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("other"),
						"Existing Error Summary",
						"Existing Error Details",
					),
					diag.NewAttributeWarningDiagnostic(
						path.Root("test"),
						"New Warning Summary",
						"New Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"New Error Summary",
						"New Error Details",
					),
				},
			},
		},
		"response-planvalue": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.PlanValue = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")})
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListNull(types.StringType),
				AttributePlan:   types.ListUnknown(types.StringType),
				AttributeState:  types.ListNull(types.StringType),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListUnknown(types.StringType),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"response-private": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.Diagnostics.Append(
								resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListNull(types.StringType),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListNull(types.StringType),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-requiresreplace-add": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("oldtestvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("oldtestvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains as it should not be removed
				},
			},
		},
		"response-requiresreplace-update": {
			block: testschema.BlockWithListPlanModifiers{
				PlanModifiers: []planmodifier.List{
					testplanmodifier.List{
						PlanModifyListMethod: func(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("oldtestvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains deduplicated
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			BlockPlanModifyList(context.Background(), testCase.block, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestBlockPlanModifyObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    fwxschema.BlockWithObjectPlanModifiers
		request  ModifyAttributePlanRequest
		response *ModifyAttributePlanResponse
		expected *ModifyAttributePlanResponse
	}{
		"request-path": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-pathexpression": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.PathExpression",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-config": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"testattr": tftypes.String,
												},
											},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"testattr": tftypes.String,
												},
											},
											map[string]tftypes.Value{
												"testattr": tftypes.NewValue(tftypes.String, "testvalue"),
											},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testattr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testattr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"testattr": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-configvalue": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.ConfigValue
							expected := types.ObjectValueMust(
								map[string]attr.Type{
									"testattr": types.StringType,
								},
								map[string]attr.Value{
									"testattr": types.StringValue("testvalue"),
								},
							)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				AttributeState: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
			},
		},
		"request-plan": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.Plan
							expected := tfsdk.Plan{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"testattr": tftypes.String,
												},
											},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"testattr": tftypes.String,
												},
											},
											map[string]tftypes.Value{
												"testattr": tftypes.NewValue(tftypes.String, "testvalue"),
											},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testattr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testattr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"testattr": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-planvalue": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.PlanValue
							expected := types.ObjectValueMust(
								map[string]attr.Type{
									"testattr": types.StringType,
								},
								map[string]attr.Value{
									"testattr": types.StringValue("testvalue"),
								},
							)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-private": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
		},
		"request-state": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.State
							expected := tfsdk.State{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"testattr": tftypes.String,
												},
											},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"testattr": tftypes.String,
												},
											},
											map[string]tftypes.Value{
												"testattr": tftypes.NewValue(tftypes.String, "testvalue"),
											},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testattr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testattr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"testattr": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-statevalue": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.StateValue
							expected := types.ObjectValueMust(
								map[string]attr.Type{
									"testattr": types.StringType,
								},
								map[string]attr.Value{
									"testattr": types.StringValue("testvalue"),
								},
							)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				AttributePlan: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
			},
		},
		"response-diagnostics": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("other"),
						"Existing Warning Summary",
						"Existing Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("other"),
						"Existing Error Summary",
						"Existing Error Details",
					),
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("other"),
						"Existing Warning Summary",
						"Existing Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("other"),
						"Existing Error Summary",
						"Existing Error Details",
					),
					diag.NewAttributeWarningDiagnostic(
						path.Root("test"),
						"New Warning Summary",
						"New Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"New Error Summary",
						"New Error Details",
					),
				},
			},
		},
		"response-planvalue": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.PlanValue = types.ObjectValueMust(
								map[string]attr.Type{
									"testattr": types.StringType,
								},
								map[string]attr.Value{
									"testattr": types.StringValue("testvalue"),
								},
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				AttributePlan: types.ObjectUnknown(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				AttributeState: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectUnknown(map[string]attr.Type{
					"testattr": types.StringType,
				}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
		},
		"response-private": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.Diagnostics.Append(
								resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectNull(map[string]attr.Type{
					"testattr": types.StringType,
				}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-requiresreplace-add": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("oldtestvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("oldtestvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains as it should not be removed
				},
			},
		},
		"response-requiresreplace-update": {
			block: testschema.BlockWithObjectPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("oldtestvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains deduplicated
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			BlockPlanModifyObject(context.Background(), testCase.block, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestBlockPlanModifySet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    fwxschema.BlockWithSetPlanModifiers
		request  ModifyAttributePlanRequest
		response *ModifyAttributePlanResponse
		expected *ModifyAttributePlanResponse
	}{
		"request-path": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected SetRequest.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-pathexpression": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected SetRequest.PathExpression",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig:         types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:           types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:          types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-config": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Set{ElementType: tftypes.String},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Set{ElementType: tftypes.String},
											[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected SetRequest.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Set{ElementType: tftypes.String},
								[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-configvalue": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							got := req.ConfigValue
							expected := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")})

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected SetRequest.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.SetNull(types.StringType),
				AttributeState:  types.SetNull(types.StringType),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetNull(types.StringType),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetNull(types.StringType),
			},
		},
		"request-plan": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							got := req.Plan
							expected := tfsdk.Plan{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Set{ElementType: tftypes.String},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Set{ElementType: tftypes.String},
											[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected SetRequest.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Set{ElementType: tftypes.String},
								[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-planvalue": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							got := req.PlanValue
							expected := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")})

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected SetRequest.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetNull(types.StringType),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetNull(types.StringType),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-private": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected SetRequest.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetNull(types.StringType),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetNull(types.StringType),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
		},
		"request-state": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							got := req.State
							expected := tfsdk.State{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Set{ElementType: tftypes.String},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Set{ElementType: tftypes.String},
											[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected SetRequest.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Set{ElementType: tftypes.String},
								[]tftypes.Value{tftypes.NewValue(tftypes.String, "testvalue")},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"request-statevalue": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							got := req.StateValue
							expected := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")})

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected SetRequest.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetNull(types.StringType),
				AttributePlan:   types.SetNull(types.StringType),
				AttributeState:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetNull(types.StringType),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetNull(types.StringType),
			},
		},
		"response-diagnostics": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("other"),
						"Existing Warning Summary",
						"Existing Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("other"),
						"Existing Error Summary",
						"Existing Error Details",
					),
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("other"),
						"Existing Warning Summary",
						"Existing Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("other"),
						"Existing Error Summary",
						"Existing Error Details",
					),
					diag.NewAttributeWarningDiagnostic(
						path.Root("test"),
						"New Warning Summary",
						"New Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"New Error Summary",
						"New Error Details",
					),
				},
			},
		},
		"response-planvalue": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							resp.PlanValue = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")})
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetNull(types.StringType),
				AttributePlan:   types.SetUnknown(types.StringType),
				AttributeState:  types.SetNull(types.StringType),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetUnknown(types.StringType),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
		},
		"response-private": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							resp.Diagnostics.Append(
								resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetNull(types.StringType),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetNull(types.StringType),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-requiresreplace-add": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("oldtestvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("oldtestvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains as it should not be removed
				},
			},
		},
		"response-requiresreplace-update": {
			block: testschema.BlockWithSetPlanModifiers{
				PlanModifiers: []planmodifier.Set{
					testplanmodifier.Set{
						PlanModifySetMethod: func(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributePlan:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				AttributeState:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("oldtestvalue")}),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("testvalue")}),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains deduplicated
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			BlockPlanModifySet(context.Background(), testCase.block, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNestedBlockObjectPlanModify(t *testing.T) {
	t.Parallel()

	fwSchema := testschema.Schema{
		Blocks: map[string]fwschema.Block{
			"test": testschema.BlockWithObjectPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringPlanModifiers{},
				},
				Blocks: map[string]fwschema.Block{
					"testblock": testschema.BlockWithObjectPlanModifiers{
						Attributes: map[string]fwschema.Attribute{
							"testblockattr": testschema.AttributeWithStringPlanModifiers{},
						},
					},
				},
			},
		},
	}
	fwValue := types.ObjectValueMust(
		map[string]attr.Type{
			"testattr": types.StringType,
			"testblock": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"testblockattr": types.StringType,
				},
			},
		},
		map[string]attr.Value{
			"testattr": types.StringValue("testvalue"),
			"testblock": types.ObjectValueMust(
				map[string]attr.Type{
					"testblockattr": types.StringType,
				},
				map[string]attr.Value{
					"testblockattr": types.StringValue("testvalue"),
				},
			),
		},
	)
	tfValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"test": tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"testattr": tftypes.String,
						"testblock": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"testblockattr": tftypes.String,
							},
						},
					},
				},
			},
		},
		map[string]tftypes.Value{
			"test": tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"testattr": tftypes.String,
						"testblock": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"testblockattr": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"testattr": tftypes.NewValue(tftypes.String, "testvalue"),
					"testblock": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"testblockattr": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"testblockattr": tftypes.NewValue(tftypes.String, "testvalue"),
						},
					),
				},
			),
		},
	)
	testConfig := tfsdk.Config{
		Raw:    tfValue,
		Schema: fwSchema,
	}
	testPlan := tfsdk.Plan{
		Raw:    tfValue,
		Schema: fwSchema,
	}
	testState := tfsdk.State{
		Raw:    tfValue,
		Schema: fwSchema,
	}

	testCases := map[string]struct {
		object   fwschema.NestedBlockObject
		request  planmodifier.ObjectRequest
		response *ModifyAttributePlanResponse
		expected *ModifyAttributePlanResponse
	}{
		"request-path": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
		},
		"request-pathexpression": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.PathExpression",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
		},
		"request-config": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.Config
							expected := testConfig

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
		},
		"request-configvalue": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.ConfigValue
							expected := fwValue

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
		},
		"request-plan": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.Plan
							expected := testPlan

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
		},
		"request-planvalue": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.PlanValue
							expected := fwValue

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
		},
		"request-private": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
				State:      testState,
				StateValue: fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
		},
		"request-state": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.State
							expected := testState

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
		},
		"request-statevalue": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							got := req.StateValue
							expected := fwValue

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected ObjectRequest.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
		},
		"response-diagnostics": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("other"),
						"Existing Warning Summary",
						"Existing Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("other"),
						"Existing Error Summary",
						"Existing Error Details",
					),
				},
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("other"),
						"Existing Warning Summary",
						"Existing Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("other"),
						"Existing Error Summary",
						"Existing Error Details",
					),
					diag.NewAttributeWarningDiagnostic(
						path.Root("test"),
						"New Warning Summary",
						"New Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"New Error Summary",
						"New Error Details",
					),
				},
				AttributePlan: fwValue,
			},
		},
		"response-diagnostics-nested-attributes": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							testplanmodifier.String{
								PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
									resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
									resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
								},
							},
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("test").AtName("testattr"),
						"New Warning Summary",
						"New Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("test").AtName("testattr"),
						"New Error Summary",
						"New Error Details",
					),
				},
			},
		},
		"response-diagnostics-nested-blocks": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				Blocks: map[string]fwschema.Block{
					"testblock": testschema.BlockWithObjectPlanModifiers{
						PlanModifiers: []planmodifier.Object{
							testplanmodifier.Object{
								PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
									resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
									resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
								},
							},
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("test").AtName("testblock"),
						"New Warning Summary",
						"New Warning Details",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("test").AtName("testblock"),
						"New Error Summary",
						"New Error Details",
					),
				},
			},
		},
		"response-planvalue": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.PlanValue = types.ObjectValueMust(
								map[string]attr.Type{
									"testattr": types.StringType,
									"testblock": types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"testblockattr": types.StringType,
										},
									},
								},
								map[string]attr.Value{
									"testattr": types.StringValue("newtestvalue"),
									"testblock": types.ObjectValueMust(
										map[string]attr.Type{
											"testblockattr": types.StringType,
										},
										map[string]attr.Value{
											"testblockattr": types.StringValue("newtestvalue"),
										},
									),
								},
							)
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
						"testblock": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"testblockattr": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
						"testblock": types.ObjectValueMust(
							map[string]attr.Type{
								"testblockattr": types.StringType,
							},
							map[string]attr.Value{
								"testblockattr": types.StringValue("testvalue"),
							},
						),
					},
				),
				State:      testState,
				StateValue: fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
						"testblock": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"testblockattr": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
						"testblock": types.ObjectValueMust(
							map[string]attr.Type{
								"testblockattr": types.StringType,
							},
							map[string]attr.Value{
								"testblockattr": types.StringValue("testvalue"),
							},
						),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
						"testblock": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"testblockattr": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"testattr": types.StringValue("newtestvalue"),
						"testblock": types.ObjectValueMust(
							map[string]attr.Type{
								"testblockattr": types.StringType,
							},
							map[string]attr.Value{
								"testblockattr": types.StringValue("newtestvalue"),
							},
						),
					},
				),
			},
		},
		"response-planvalue-nested-attributes": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							testplanmodifier.String{
								PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
									resp.PlanValue = types.StringValue("newtestvalue")
								},
							},
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
						"testblock": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"testblockattr": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
						"testblock": types.ObjectValueMust(
							map[string]attr.Type{
								"testblockattr": types.StringType,
							},
							map[string]attr.Value{
								"testblockattr": types.StringValue("testvalue"),
							},
						),
					},
				),
				State:      testState,
				StateValue: fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
						"testblock": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"testblockattr": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
						"testblock": types.ObjectValueMust(
							map[string]attr.Type{
								"testblockattr": types.StringType,
							},
							map[string]attr.Value{
								"testblockattr": types.StringValue("testvalue"),
							},
						),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
						"testblock": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"testblockattr": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"testattr": types.StringValue("newtestvalue"),
						"testblock": types.ObjectValueMust(
							map[string]attr.Type{
								"testblockattr": types.StringType,
							},
							map[string]attr.Value{
								"testblockattr": types.StringValue("testvalue"),
							},
						),
					},
				),
			},
		},
		"response-planvalue-nested-blocks": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				Blocks: map[string]fwschema.Block{
					"testblock": testschema.BlockWithObjectPlanModifiers{
						PlanModifiers: []planmodifier.Object{
							testplanmodifier.Object{
								PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
									resp.PlanValue = types.ObjectValueMust(
										map[string]attr.Type{
											"testblockattr": types.StringType,
										},
										map[string]attr.Value{
											"testblockattr": types.StringValue("newtestvalue"),
										},
									)
								},
							},
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
						"testblock": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"testblockattr": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
						"testblock": types.ObjectValueMust(
							map[string]attr.Type{
								"testblockattr": types.StringType,
							},
							map[string]attr.Value{
								"testblockattr": types.StringValue("testvalue"),
							},
						),
					},
				),
				State:      testState,
				StateValue: fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
						"testblock": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"testblockattr": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
						"testblock": types.ObjectValueMust(
							map[string]attr.Type{
								"testblockattr": types.StringType,
							},
							map[string]attr.Value{
								"testblockattr": types.StringValue("testvalue"),
							},
						),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
						"testblock": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"testblockattr": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"testattr": types.StringValue("testvalue"),
						"testblock": types.ObjectValueMust(
							map[string]attr.Type{
								"testblockattr": types.StringType,
							},
							map[string]attr.Value{
								"testblockattr": types.StringValue("newtestvalue"),
							},
						),
					},
				),
			},
		},
		"response-private": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.Diagnostics.Append(
								resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
							)
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
				State:      testState,
				StateValue: fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-private-nested-attributes": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							testplanmodifier.String{
								PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
									resp.Diagnostics.Append(
										resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
									)
								},
							},
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
				State:      testState,
				StateValue: fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-private-nested-blocks": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				Blocks: map[string]fwschema.Block{
					"testblock": testschema.BlockWithObjectPlanModifiers{
						PlanModifiers: []planmodifier.Object{
							testplanmodifier.Object{
								PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
									resp.Diagnostics.Append(
										resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
									)
								},
							},
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
				State:      testState,
				StateValue: fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-requiresreplace-add": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				RequiresReplace: path.Paths{
					path.Root("test"), // set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				RequiresReplace: path.Paths{
					path.Root("test"), // should not be removed
				},
			},
		},
		"response-requiresreplace-nested-attributes": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							testplanmodifier.String{
								PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
									resp.RequiresReplace = true
								},
							},
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				RequiresReplace: path.Paths{
					path.Root("test").AtName("testattr"),
				},
			},
		},
		"response-requiresreplace-nested-blocks": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				Blocks: map[string]fwschema.Block{
					"testblock": testschema.BlockWithObjectPlanModifiers{
						PlanModifiers: []planmodifier.Object{
							testplanmodifier.Object{
								PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
									resp.RequiresReplace = true
								},
							},
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				RequiresReplace: path.Paths{
					path.Root("test").AtName("testblock"),
				},
			},
		},
		"response-requiresreplace-update": {
			object: testschema.NestedBlockObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config:         testConfig,
				ConfigValue:    fwValue,
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan:           testPlan,
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				RequiresReplace: path.Paths{
					path.Root("test"), // set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
				RequiresReplace: path.Paths{
					path.Root("test"), // remains deduplicated
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			NestedBlockObjectPlanModify(context.Background(), testCase.object, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

type testBlockPlanModifierNullList struct{}

func (t testBlockPlanModifierNullList) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	resp.PlanValue = types.ListNull(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"nested_attr": types.StringType,
			},
		},
	)
}

func (t testBlockPlanModifierNullList) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testBlockPlanModifierNullList) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type testBlockPlanModifierPrivateGet struct{}

func (t testBlockPlanModifierPrivateGet) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
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

func (t testBlockPlanModifierPrivateSet) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	diags := resp.Private.SetKey(ctx, "providerKeyOne", []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`))

	resp.Diagnostics.Append(diags...)
}

func (t testBlockPlanModifierPrivateSet) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testBlockPlanModifierPrivateSet) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}
