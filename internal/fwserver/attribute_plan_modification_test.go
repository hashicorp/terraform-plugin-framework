// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"fmt"
	"math/big"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAttributeModifyPlan(t *testing.T) {
	t.Parallel()

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testCases := map[string]struct {
		attribute    fwschema.Attribute
		req          ModifyAttributePlanRequest
		expectedResp ModifyAttributePlanResponse
	}{
		"no-plan-modifiers": {
			attribute: testschema.Attribute{
				Type:     types.StringType,
				Required: true,
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("testvalue"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("testvalue"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"attribute-plan": {
			attribute: testschema.AttributeWithStringPlanModifiers{
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
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTATTRONE"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTATTRONE"),
				AttributeState:  types.StringValue("TESTATTRONE"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("MODIFIED_TWO"),
			},
		},
		"attribute-request-private": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				Required: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.TestAttrPlanPrivateModifierGet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTATTRONE"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTATTRONE"),
				AttributeState:  types.StringValue("TESTATTRONE"),
				Private:         testProviderData,
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTATTRONE"),
				Private:       testProviderData,
			},
		},
		"attribute-response-private": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				Required: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTATTRONE"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTATTRONE"),
				AttributeState:  types.StringValue("TESTATTRONE"),
				Private:         privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTATTRONE"),
				Private:       testProviderData,
			},
		},
		"attribute-list-nested-private": {
			attribute: testschema.NestedAttributeWithListPlanModifiers{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								planmodifiers.TestAttrPlanPrivateModifierGet{},
							},
						},
					},
				},
				Required: true,
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
		"attribute-list-nested-usestateforunknown": {
			attribute: testschema.NestedAttributeWithListPlanModifiers{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.Attribute{
							Type:     types.StringType,
							Computed: true,
						},
					},
				},
				Computed: true,
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
		"attribute-list-nested-nested-usestateforunknown-elements-rearranged": {
			attribute: testschema.NestedAttribute{
				NestedObject: testschema.NestedAttributeObject{
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
				NestingMode: fwschema.NestingModeList,
				Required:    true,
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
								"nested_required": types.StringValue("testvalue1"), // prior state on index 1 is testvalue2
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
		"attribute-list-nested-nested-usestateforunknown-elements-removed": {
			attribute: testschema.NestedAttribute{
				NestedObject: testschema.NestedAttributeObject{
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
				NestingMode: fwschema.NestingModeList,
				Required:    true,
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
		"attribute-set-nested-private": {
			attribute: testschema.NestedAttributeWithSetPlanModifiers{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								planmodifiers.TestAttrPlanPrivateModifierGet{},
							},
						},
					},
				},
				Required: true,
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
		"attribute-set-nested-usestateforunknown": {
			attribute: testschema.NestedAttributeWithSetPlanModifiers{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.Attribute{
							Type:     types.StringType,
							Computed: true,
						},
					},
				},
				Computed: true,
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
		"attribute-set-nested-nested-usestateforunknown": {
			attribute: testschema.NestedAttribute{
				NestedObject: testschema.NestedAttributeObject{
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
				NestingMode: fwschema.NestingModeSet,
				Required:    true,
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
		"attribute-set-nested-nested-usestateforunknown-elements-rearranged": {
			attribute: testschema.NestedAttribute{
				NestedObject: testschema.NestedAttributeObject{
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
				NestingMode: fwschema.NestingModeSet,
				Required:    true,
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
								"nested_required": types.StringValue("testvalue1"), // prior state on index 1 is testvalue2
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
		"attribute-set-nested-nested-usestateforunknown-elements-removed": {
			attribute: testschema.NestedAttribute{
				NestedObject: testschema.NestedAttributeObject{
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
				NestingMode: fwschema.NestingModeSet,
				Required:    true,
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
		"attribute-map-nested-private": {
			attribute: testschema.NestedAttributeWithMapPlanModifiers{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_attr": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								planmodifiers.TestAttrPlanPrivateModifierGet{},
							},
						},
					},
				},
				Required: true,
				PlanModifiers: []planmodifier.Map{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					map[string]attr.Value{
						"testkey": types.ObjectValueMust(
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
				AttributePlan: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					map[string]attr.Value{
						"testkey": types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.StringValue("testvalue"),
							},
						),
					},
				),
				AttributeState: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					map[string]attr.Value{
						"testkey": types.ObjectValueMust(
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
				AttributePlan: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					map[string]attr.Value{
						"testkey": types.ObjectValueMust(
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
		"attribute-map-nested-usestateforunknown": {
			attribute: testschema.NestedAttributeWithMapPlanModifiers{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.Attribute{
							Type:     types.StringType,
							Computed: true,
						},
					},
				},
				Computed: true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.MapNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.MapUnknown(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
				),
				AttributeState: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
					map[string]attr.Value{
						"key1": types.ObjectValueMust(
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
				AttributePlan: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
						},
					},
					map[string]attr.Value{
						"key1": types.ObjectValueMust(
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
		"attribute-single-nested-private": {
			attribute: testschema.NestedAttributeWithObjectPlanModifiers{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"testing": testschema.AttributeWithStringPlanModifiers{
							Required: true,
							PlanModifiers: []planmodifier.String{
								planmodifiers.TestAttrPlanPrivateModifierGet{},
							},
						},
					},
				},
				Required: true,
				PlanModifiers: []planmodifier.Object{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{
						"testing": types.StringType,
					},
					map[string]attr.Value{
						"testing": types.StringValue("testvalue"),
					},
				),
				AttributePath: path.Root("test"),
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testing": types.StringType,
					},
					map[string]attr.Value{
						"testing": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.ObjectValueMust(
					map[string]attr.Type{
						"testing": types.StringType,
					},
					map[string]attr.Value{
						"testing": types.StringValue("testvalue"),
					},
				),
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testing": types.StringType,
					},
					map[string]attr.Value{
						"testing": types.StringValue("testvalue"),
					},
				),
				Private: testProviderData,
			},
		},
		"attribute-single-nested-usestateforunknown": {
			attribute: testschema.NestedAttributeWithObjectPlanModifiers{
				NestedObject: testschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"nested_computed": testschema.Attribute{
							Type:     types.StringType,
							Computed: true,
						},
					},
				},
				Computed: true,
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
		"requires-replacement": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("newtestvalue"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("newtestvalue"),
				AttributeState:  types.StringValue("testvalue"),
				// resource.RequiresReplace() requires non-null plan
				// and state.
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "newtestvalue"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test": schema.StringAttribute{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test": schema.StringAttribute{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("newtestvalue"),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"requires-replacement-passthrough": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				Required: true,
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							if req.PlanValue.ValueString() == "TESTATTRONE" {
								resp.PlanValue = types.StringValue("TESTATTRTWO")
							}
						},
					},
					stringplanmodifier.RequiresReplace(),
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTATTRONE"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTATTRONE"),
				AttributeState:  types.StringValue("TESTATTRONE"),
				// resource.RequiresReplace() requires non-null plan
				// and state.
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test": schema.StringAttribute{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, sr1 planmodifier.StringRequest, sr2 *planmodifier.StringResponse) {
											// Do nothing; RequiresReplace should still be in effect
										},
									},
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"test": schema.StringAttribute{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, sr1 planmodifier.StringRequest, sr2 *planmodifier.StringResponse) {
											// Do nothing; RequiresReplace should still be in effect
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTATTRTWO"),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"requires-replacement-unset": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.RequiresReplace = false
						},
					},
				},
			},
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("testvalue"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("testvalue"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"warnings": {
			attribute: testschema.AttributeWithStringPlanModifiers{
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
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTDIAG"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTDIAG"),
				AttributeState:  types.StringValue("TESTDIAG"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTDIAG"),
				Diagnostics: diag.Diagnostics{
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
			},
		},
		"error": {
			attribute: testschema.AttributeWithStringPlanModifiers{
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
			req: ModifyAttributePlanRequest{
				AttributeConfig: types.StringValue("TESTDIAG"),
				AttributePath:   path.Root("test"),
				AttributePlan:   types.StringValue("TESTDIAG"),
				AttributeState:  types.StringValue("TESTDIAG"),
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("TESTDIAG"),
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			got := ModifyAttributePlanResponse{
				AttributePlan: tc.req.AttributePlan,
				Private:       tc.req.Private,
			}

			AttributeModifyPlan(ctx, tc.attribute, tc.req, &got)

			if diff := cmp.Diff(tc.expectedResp, got, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("Unexpected response (-wanted, +got): %s", diff)
			}
		})
	}
}

func TestAttributePlanModifyBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithBoolPlanModifiers
		request   ModifyAttributePlanRequest
		response  *ModifyAttributePlanResponse
		expected  *ModifyAttributePlanResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected BoolRequest.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolValue(true),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected BoolRequest.PathExpression",
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
				AttributeConfig:         types.BoolValue(true),
				AttributePlan:           types.BoolValue(true),
				AttributeState:          types.BoolValue(true),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
		},
		"request-config": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Bool,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Bool, true),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected BoolRequest.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolValue(true),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Bool,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Bool, true),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							got := req.ConfigValue
							expected := types.BoolValue(true)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected BoolRequest.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
				AttributePlan:   types.BoolNull(),
				AttributeState:  types.BoolNull(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolNull(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolNull(),
			},
		},
		"request-plan": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							got := req.Plan
							expected := tfsdk.Plan{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Bool,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Bool, true),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected BoolRequest.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolValue(true),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Bool,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Bool, true),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
		},
		"request-planvalue": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							got := req.PlanValue
							expected := types.BoolValue(true)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected BoolRequest.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolNull(),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolNull(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
		},
		"request-private": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected BoolRequest.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolNull(),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolNull(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
		},
		"request-state": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							got := req.State
							expected := tfsdk.State{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Bool,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Bool, true),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected BoolRequest.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolValue(true),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Bool,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Bool, true),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
		},
		"request-statevalue": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							got := req.StateValue
							expected := types.BoolValue(true)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected BoolRequest.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolNull(),
				AttributePlan:   types.BoolNull(),
				AttributeState:  types.BoolValue(true),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolNull(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolNull(),
			},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolValue(true),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
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
				AttributePlan: types.BoolValue(true),
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
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							resp.PlanValue = types.BoolValue(true)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolNull(),
				AttributePlan:   types.BoolUnknown(),
				AttributeState:  types.BoolNull(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolUnknown(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
		},
		"response-private": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							resp.Diagnostics.Append(
								resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolNull(),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolNull(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-requiresreplace-add": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolValue(false),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolValue(false),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains as it should not be removed
				},
			},
		},
		"response-requiresreplace-update": {
			attribute: testschema.AttributeWithBoolPlanModifiers{
				PlanModifiers: []planmodifier.Bool{
					testplanmodifier.Bool{
						PlanModifyBoolMethod: func(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
				AttributePlan:   types.BoolValue(true),
				AttributeState:  types.BoolValue(false),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.BoolValue(true),
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

			AttributePlanModifyBool(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributePlanModifyFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithFloat64PlanModifiers
		request   ModifyAttributePlanRequest
		response  *ModifyAttributePlanResponse
		expected  *ModifyAttributePlanResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Float64Request.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Value(1.2),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Float64Request.PathExpression",
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
				AttributeConfig:         types.Float64Value(1.2),
				AttributePlan:           types.Float64Value(1.2),
				AttributeState:          types.Float64Value(1.2),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
		},
		"request-config": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 1.2),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected Float64Request.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Value(1.2),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							got := req.ConfigValue
							expected := types.Float64Value(1.2)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Float64Request.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
				AttributePlan:   types.Float64Null(),
				AttributeState:  types.Float64Null(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Null(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Null(),
			},
		},
		"request-plan": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							got := req.Plan
							expected := tfsdk.Plan{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 1.2),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected Float64Request.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Value(1.2),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
		},
		"request-planvalue": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							got := req.PlanValue
							expected := types.Float64Value(1.2)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Float64Request.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Null(),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Null(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
		},
		"request-private": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected Float64Request.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Null(),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Null(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
		},
		"request-state": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							got := req.State
							expected := tfsdk.State{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 1.2),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected Float64Request.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Value(1.2),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
		},
		"request-statevalue": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							got := req.StateValue
							expected := types.Float64Value(1.2)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Float64Request.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Null(),
				AttributePlan:   types.Float64Null(),
				AttributeState:  types.Float64Value(1.2),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Null(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Null(),
			},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Value(1.2),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
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
				AttributePlan: types.Float64Value(1.2),
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
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							resp.PlanValue = types.Float64Value(1.2)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Null(),
				AttributePlan:   types.Float64Unknown(),
				AttributeState:  types.Float64Null(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Unknown(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
		},
		"response-private": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							resp.Diagnostics.Append(
								resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Null(),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Null(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-requiresreplace-add": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Value(2.4),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Value(2.4),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains as it should not be removed
				},
			},
		},
		"response-requiresreplace-update": {
			attribute: testschema.AttributeWithFloat64PlanModifiers{
				PlanModifiers: []planmodifier.Float64{
					testplanmodifier.Float64{
						PlanModifyFloat64Method: func(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
				AttributePlan:   types.Float64Value(1.2),
				AttributeState:  types.Float64Value(2.4),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Float64Value(1.2),
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

			AttributePlanModifyFloat64(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributePlanModifyInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithInt64PlanModifiers
		request   ModifyAttributePlanRequest
		response  *ModifyAttributePlanResponse
		expected  *ModifyAttributePlanResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Int64Request.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(1),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Value(1),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Int64Request.PathExpression",
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
				AttributeConfig:         types.Int64Value(1),
				AttributePlan:           types.Int64Value(1),
				AttributeState:          types.Int64Value(1),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
		},
		"request-config": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 1.2),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected Int64Request.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(1),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Value(1),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							got := req.ConfigValue
							expected := types.Int64Value(1)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Int64Request.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(1),
				AttributePlan:   types.Int64Null(),
				AttributeState:  types.Int64Null(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Null(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Null(),
			},
		},
		"request-plan": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							got := req.Plan
							expected := tfsdk.Plan{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 1.2),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected Int64Request.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(1),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Value(1),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
		},
		"request-planvalue": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							got := req.PlanValue
							expected := types.Int64Value(1)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Int64Request.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Null(),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Null(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
		},
		"request-private": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected Int64Request.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Null(),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Null(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
		},
		"request-state": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							got := req.State
							expected := tfsdk.State{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 1.2),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected Int64Request.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(1),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Value(1),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
		},
		"request-statevalue": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							got := req.StateValue
							expected := types.Int64Value(1)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected Int64Request.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Null(),
				AttributePlan:   types.Int64Null(),
				AttributeState:  types.Int64Value(1),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Null(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Null(),
			},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(1),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Value(1),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
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
				AttributePlan: types.Int64Value(1),
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
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							resp.PlanValue = types.Int64Value(1)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Null(),
				AttributePlan:   types.Int64Unknown(),
				AttributeState:  types.Int64Null(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Unknown(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
		},
		"response-private": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							resp.Diagnostics.Append(
								resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Null(),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Null(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-requiresreplace-add": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(1),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Value(2),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(1),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Value(2),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains as it should not be removed
				},
			},
		},
		"response-requiresreplace-update": {
			attribute: testschema.AttributeWithInt64PlanModifiers{
				PlanModifiers: []planmodifier.Int64{
					testplanmodifier.Int64{
						PlanModifyInt64Method: func(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(1),
				AttributePlan:   types.Int64Value(1),
				AttributeState:  types.Int64Value(2),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.Int64Value(1),
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

			AttributePlanModifyInt64(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributePlanModifyList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithListPlanModifiers
		request   ModifyAttributePlanRequest
		response  *ModifyAttributePlanResponse
		expected  *ModifyAttributePlanResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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
			attribute: testschema.AttributeWithListPlanModifiers{
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

			AttributePlanModifyList(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributePlanModifyMap(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithMapPlanModifiers
		request   ModifyAttributePlanRequest
		response  *ModifyAttributePlanResponse
		expected  *ModifyAttributePlanResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected MapRequest.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected MapRequest.PathExpression",
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
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-config": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Map{ElementType: tftypes.String},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Map{ElementType: tftypes.String},
											map[string]tftypes.Value{
												"testkey": tftypes.NewValue(tftypes.String, "testvalue"),
											},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected MapRequest.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Map{ElementType: tftypes.String},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							got := req.ConfigValue
							expected := types.MapValueMust(
								types.StringType,
								map[string]attr.Value{
									"testkey": types.StringValue("testvalue"),
								},
							)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected MapRequest.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan:  types.MapNull(types.StringType),
				AttributeState: types.MapNull(types.StringType),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapNull(types.StringType),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapNull(types.StringType),
			},
		},
		"request-plan": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							got := req.Plan
							expected := tfsdk.Plan{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Map{ElementType: tftypes.String},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Map{ElementType: tftypes.String},
											map[string]tftypes.Value{
												"testkey": tftypes.NewValue(tftypes.String, "testvalue"),
											},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected MapRequest.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Map{ElementType: tftypes.String},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-planvalue": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							got := req.PlanValue
							expected := types.MapValueMust(
								types.StringType,
								map[string]attr.Value{
									"testkey": types.StringValue("testvalue"),
								},
							)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected MapRequest.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.MapNull(types.StringType),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapNull(types.StringType),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-private": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected MapRequest.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.MapNull(types.StringType),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapNull(types.StringType),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
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
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
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
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							got := req.State
							expected := tfsdk.State{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Map{ElementType: tftypes.String},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Map{ElementType: tftypes.String},
											map[string]tftypes.Value{
												"testkey": tftypes.NewValue(tftypes.String, "testvalue"),
											},
										),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected MapRequest.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Map{ElementType: tftypes.String},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
		},
		"request-statevalue": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							got := req.StateValue
							expected := types.MapValueMust(
								types.StringType,
								map[string]attr.Value{
									"testkey": types.StringValue("testvalue"),
								},
							)

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected MapRequest.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.MapNull(types.StringType),
				AttributePlan:   types.MapNull(types.StringType),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapNull(types.StringType),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapNull(types.StringType),
			},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
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
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
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
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							resp.PlanValue = types.MapValueMust(
								types.StringType,
								map[string]attr.Value{
									"testkey": types.StringValue("testvalue"),
								},
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.MapNull(types.StringType),
				AttributePlan:   types.MapUnknown(types.StringType),
				AttributeState:  types.MapNull(types.StringType),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapUnknown(types.StringType),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
		},
		"response-private": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							resp.Diagnostics.Append(
								resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.MapNull(types.StringType),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapNull(types.StringType),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
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
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
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
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("oldtestvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("oldtestvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains as it should not be removed
				},
			},
		},
		"response-requiresreplace-update": {
			attribute: testschema.AttributeWithMapPlanModifiers{
				PlanModifiers: []planmodifier.Map{
					testplanmodifier.Map{
						PlanModifyMapMethod: func(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				AttributeState: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("oldtestvalue"),
					},
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("testvalue"),
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

			AttributePlanModifyMap(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributePlanModifyNumber(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithNumberPlanModifiers
		request   ModifyAttributePlanRequest
		response  *ModifyAttributePlanResponse
		expected  *ModifyAttributePlanResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected NumberRequest.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1)),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberValue(big.NewFloat(1)),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected NumberRequest.PathExpression",
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
				AttributeConfig:         types.NumberValue(big.NewFloat(1)),
				AttributePlan:           types.NumberValue(big.NewFloat(1)),
				AttributeState:          types.NumberValue(big.NewFloat(1)),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
		},
		"request-config": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 1.2),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected NumberRequest.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1)),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberValue(big.NewFloat(1)),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							got := req.ConfigValue
							expected := types.NumberValue(big.NewFloat(1))

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected NumberRequest.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1)),
				AttributePlan:   types.NumberNull(),
				AttributeState:  types.NumberNull(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberNull(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberNull(),
			},
		},
		"request-plan": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							got := req.Plan
							expected := tfsdk.Plan{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 1.2),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected NumberRequest.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1)),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberValue(big.NewFloat(1)),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
		},
		"request-planvalue": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							got := req.PlanValue
							expected := types.NumberValue(big.NewFloat(1))

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected NumberRequest.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberNull(),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberNull(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
		},
		"request-private": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected NumberRequest.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberNull(),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberNull(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
		},
		"request-state": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							got := req.State
							expected := tfsdk.State{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 1.2),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected NumberRequest.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1)),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberValue(big.NewFloat(1)),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
		},
		"request-statevalue": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							got := req.StateValue
							expected := types.NumberValue(big.NewFloat(1))

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected NumberRequest.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberNull(),
				AttributePlan:   types.NumberNull(),
				AttributeState:  types.NumberValue(big.NewFloat(1)),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberNull(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberNull(),
			},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1)),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberValue(big.NewFloat(1)),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
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
				AttributePlan: types.NumberValue(big.NewFloat(1)),
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
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							resp.PlanValue = types.NumberValue(big.NewFloat(1))
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberNull(),
				AttributePlan:   types.NumberUnknown(),
				AttributeState:  types.NumberNull(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberUnknown(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
		},
		"response-private": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							resp.Diagnostics.Append(
								resp.Private.SetKey(ctx, "testkey", []byte(`{"newtestproperty":true}`))...,
							)
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberNull(),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberNull(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-requiresreplace-add": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1)),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberValue(big.NewFloat(2)),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1)),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberValue(big.NewFloat(2)),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains as it should not be removed
				},
			},
		},
		"response-requiresreplace-update": {
			attribute: testschema.AttributeWithNumberPlanModifiers{
				PlanModifiers: []planmodifier.Number{
					testplanmodifier.Number{
						PlanModifyNumberMethod: func(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1)),
				AttributePlan:   types.NumberValue(big.NewFloat(1)),
				AttributeState:  types.NumberValue(big.NewFloat(2)),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.NumberValue(big.NewFloat(1)),
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

			AttributePlanModifyNumber(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributePlanModifyObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithObjectPlanModifiers
		request   ModifyAttributePlanRequest
		response  *ModifyAttributePlanResponse
		expected  *ModifyAttributePlanResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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
			attribute: testschema.AttributeWithObjectPlanModifiers{
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

			AttributePlanModifyObject(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributePlanModifySet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithSetPlanModifiers
		request   ModifyAttributePlanRequest
		response  *ModifyAttributePlanResponse
		expected  *ModifyAttributePlanResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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
			attribute: testschema.AttributeWithSetPlanModifiers{
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

			AttributePlanModifySet(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributePlanModifyString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithStringPlanModifiers
		request   ModifyAttributePlanRequest
		response  *ModifyAttributePlanResponse
		expected  *ModifyAttributePlanResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							got := req.Path
							expected := path.Root("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected StringRequest.Path",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("testvalue"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("testvalue"),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							got := req.PathExpression
							expected := path.MatchRoot("test")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected StringRequest.PathExpression",
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
				AttributeConfig:         types.StringValue("testvalue"),
				AttributePlan:           types.StringValue("testvalue"),
				AttributeState:          types.StringValue("testvalue"),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"request-config": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "testvalue"),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected StringRequest.Config",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("testvalue"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("testvalue"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.String, "testvalue"),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							got := req.ConfigValue
							expected := types.StringValue("testvalue")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected StringRequest.ConfigValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("testvalue"),
				AttributePlan:   types.StringNull(),
				AttributeState:  types.StringNull(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringNull(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringNull(),
			},
		},
		"request-plan": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							got := req.Plan
							expected := tfsdk.Plan{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "testvalue"),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected StringRequest.Plan",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("testvalue"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("testvalue"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.String, "testvalue"),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"request-planvalue": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							got := req.PlanValue
							expected := types.StringValue("testvalue")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected StringRequest.PlanValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringNull(),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringNull(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"request-private": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							got, diags := req.Private.GetKey(ctx, "testkey")
							expected := []byte(`{"testproperty":true}`)

							resp.Diagnostics.Append(diags...)

							if diff := cmp.Diff(got, expected); diff != "" {
								resp.Diagnostics.AddError(
									"Unexpected StringRequest.Private",
									diff,
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringNull(),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringNull(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
		},
		"request-state": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							got := req.State
							expected := tfsdk.State{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "testvalue"),
									},
								),
							}

							if !got.Raw.Equal(expected.Raw) {
								resp.Diagnostics.AddError(
									"Unexpected StringRequest.State",
									fmt.Sprintf("expected %s, got: %s", expected.Raw, got.Raw),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("testvalue"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("testvalue"),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.String, "testvalue"),
						},
					),
				},
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"request-statevalue": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							got := req.StateValue
							expected := types.StringValue("testvalue")

							if !got.Equal(expected) {
								resp.Diagnostics.AddError(
									"Unexpected StringRequest.StateValue",
									fmt.Sprintf("expected %s, got: %s", expected, got),
								)
							}
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringNull(),
				AttributePlan:   types.StringNull(),
				AttributeState:  types.StringValue("testvalue"),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringNull(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringNull(),
			},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("testvalue"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("testvalue"),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
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
				AttributePlan: types.StringValue("testvalue"),
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
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.PlanValue = types.StringValue("testvalue")
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringNull(),
				AttributePlan:   types.StringUnknown(),
				AttributeState:  types.StringNull(),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringUnknown(),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
		},
		"response-private": {
			attribute: testschema.AttributeWithStringPlanModifiers{
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
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringNull(),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringNull(),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`),
					}),
				),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"testproperty":true}`), // copied from request
					}),
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
				Private: privatestate.MustProviderData(
					context.Background(),
					privatestate.MustMarshalToJson(map[string][]byte{
						"testkey": []byte(`{"newtestproperty":true}`),
					}),
				),
			},
		},
		"response-requiresreplace-add": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("testvalue"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("oldtestvalue"),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"response-requiresreplace-false": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.RequiresReplace = false // same as not being set
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("testvalue"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("oldtestvalue"),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
				RequiresReplace: path.Paths{
					path.Root("test"), // Remains as it should not be removed
				},
			},
		},
		"response-requiresreplace-update": {
			attribute: testschema.AttributeWithStringPlanModifiers{
				PlanModifiers: []planmodifier.String{
					testplanmodifier.String{
						PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
							resp.RequiresReplace = true
						},
					},
				},
			},
			request: ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("testvalue"),
				AttributePlan:   types.StringValue("testvalue"),
				AttributeState:  types.StringValue("oldtestvalue"),
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
				RequiresReplace: path.Paths{
					path.Root("test"), // Set by prior plan modifier
				},
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.StringValue("testvalue"),
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

			AttributePlanModifyString(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNestedAttributeObjectPlanModify(t *testing.T) {
	t.Parallel()

	fwSchema := testschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"test": testschema.AttributeWithObjectPlanModifiers{
				AttributeTypes: map[string]attr.Type{
					"testattr": types.StringType,
				},
				Required: true,
			},
		},
	}
	fwValue := types.ObjectValueMust(
		map[string]attr.Type{"testattr": types.StringType},
		map[string]attr.Value{"testattr": types.StringValue("testvalue")},
	)
	tfValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"test": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
			},
		},
		map[string]tftypes.Value{
			"test": tftypes.NewValue(
				tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
				map[string]tftypes.Value{
					"testattr": tftypes.NewValue(tftypes.String, "testvalue"),
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
		object   fwschema.NestedAttributeObject
		request  planmodifier.ObjectRequest
		response *ModifyAttributePlanResponse
		expected *ModifyAttributePlanResponse
	}{
		"request-path": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
		"response-diagnostics-nested": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
		"response-planvalue": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.PlanValue = types.ObjectValueMust(
								map[string]attr.Type{
									"testattr": types.StringType,
								},
								map[string]attr.Value{
									"testattr": types.StringValue("newtestvalue"),
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
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("newtestvalue"),
					},
				),
			},
		},
		"response-planvalue-nested": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
				PlanValue:      fwValue,
				State:          testState,
				StateValue:     fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: fwValue,
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("newtestvalue"),
					},
				),
			},
		},
		"response-planvalue-unknown-to-known": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
				PlanModifiers: []planmodifier.Object{
					testplanmodifier.Object{
						PlanModifyObjectMethod: func(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
							resp.PlanValue = types.ObjectValueMust(
								map[string]attr.Type{
									"testattr": types.StringType,
								},
								map[string]attr.Value{
									"testattr": types.StringValue("newtestvalue"),
								},
							)
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
								nil,
							),
						},
					),
					Schema: fwSchema,
				},
				ConfigValue: types.ObjectNull(
					map[string]attr.Type{"testattr": types.StringType},
				),
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
								tftypes.UnknownValue,
							),
						},
					),
					Schema: fwSchema,
				},
				PlanValue: types.ObjectUnknown(
					map[string]attr.Type{"testattr": types.StringType},
				),
				State:      testState,
				StateValue: fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectUnknown(
					map[string]attr.Type{"testattr": types.StringType},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("newtestvalue"),
					},
				),
			},
		},
		"response-planvalue-unknown-to-known-nested": {
			object: testschema.NestedAttributeObject{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringPlanModifiers{
						Required: true,
						PlanModifiers: []planmodifier.String{
							testplanmodifier.String{
								PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
									resp.PlanValue = types.StringValue("newtestvalue") // should win over object
								},
							},
						},
					},
				},
			},
			request: planmodifier.ObjectRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
								nil,
							),
						},
					),
					Schema: fwSchema,
				},
				ConfigValue: types.ObjectNull(
					map[string]attr.Type{"testattr": types.StringType},
				),
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
								tftypes.UnknownValue,
							),
						},
					),
					Schema: fwSchema,
				},
				PlanValue: types.ObjectUnknown(
					map[string]attr.Type{"testattr": types.StringType},
				),
				State:      testState,
				StateValue: fwValue,
			},
			response: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectUnknown(
					map[string]attr.Type{"testattr": types.StringType},
				),
			},
			expected: &ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"testattr": types.StringType,
					},
					map[string]attr.Value{
						"testattr": types.StringValue("newtestvalue"),
					},
				),
			},
		},
		"response-private": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
		"response-private-nested": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
		"response-requiresreplace-add": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
		"response-requiresreplace-nested": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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
		"response-requiresreplace-update": {
			object: testschema.NestedAttributeObjectWithPlanModifiers{
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

			NestedAttributeObjectPlanModify(context.Background(), testCase.object, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
