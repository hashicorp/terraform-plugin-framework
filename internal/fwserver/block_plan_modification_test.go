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
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ attr.Type = CustomType{}
)

type CustomType struct {
	types.ObjectType
}

func (t CustomType) Equal(candidate attr.Type) bool {
	other, ok := candidate.(CustomType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t CustomType) ValueType(_ context.Context) attr.Value {
	return CustomValue{
		types.Object{
			AttrTypes: t.AttrTypes,
		},
	}
}

var (
	_ attr.ValueWithAttrs = CustomValue{}
)

type CustomValue struct {
	types.Object
}

func (t CustomValue) SetAttrs(attrs map[string]attr.Value) attr.ValueWithAttrs {
	t.Object.Attrs = attrs

	return t
}

func (t CustomValue) Type(_ context.Context) attr.Type {
	return CustomType{
		types.ObjectType{AttrTypes: t.AttrTypes},
	}
}

func (t CustomValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	if t.AttrTypes == nil {
		return tftypes.Value{}, fmt.Errorf("cannot convert CustomValue to tftypes.Value if AttrTypes field is not set")
	}

	return t.Object.ToTerraformValue(ctx)
}

func (t CustomValue) Equal(c attr.Value) bool {
	other, ok := c.(CustomValue)
	if !ok {
		return false
	}

	return t.Object.Equal(other.Object)
}

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

	schemaCustomType := func(blockPlanModifiers tfsdk.AttributePlanModifiers, nestedAttrPlanModifiers tfsdk.AttributePlanModifiers) tfsdk.Schema {
		return tfsdk.Schema{
			Blocks: map[string]tfsdk.Block{
				"test": {
					Typ: CustomType{},
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

	type modifyAttributePlanValues struct {
		config string
		plan   string
		state  string
	}

	modifyAttributePlanRequest := func(attrPath path.Path, schema tfsdk.Schema, values modifyAttributePlanValues) tfsdk.ModifyAttributePlanRequest {
		return tfsdk.ModifyAttributePlanRequest{
			AttributeConfig: types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: values.config},
						},
					},
				},
			},
			AttributePath: attrPath,
			AttributePlan: types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: values.plan},
						},
					},
				},
			},
			AttributeState: types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: values.state},
						},
					},
				},
			},
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

	modifyAttributePlanRequestCustomType := func(attrPath path.Path, schema tfsdk.Schema, values modifyAttributePlanValues) tfsdk.ModifyAttributePlanRequest {
		return tfsdk.ModifyAttributePlanRequest{
			AttributeConfig: types.List{
				ElemType: CustomType{
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				},
				Elems: []attr.Value{
					CustomValue{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: values.config},
							},
						},
					},
				},
			},
			AttributePath: attrPath,
			AttributePlan: types.List{
				ElemType: CustomType{
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				},
				Elems: []attr.Value{
					CustomValue{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: values.plan},
							},
						},
					},
				},
			},
			AttributeState: types.List{
				ElemType: CustomType{
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
				},
				Elems: []attr.Value{
					CustomValue{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: values.state},
							},
						},
					},
				},
			},
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

	modifyAttributePlanWithPrivateRequestCustomType := func(attrPath path.Path, schema tfsdk.Schema, values modifyAttributePlanValues, privateProviderData *privatestate.ProviderData) tfsdk.ModifyAttributePlanRequest {
		req := modifyAttributePlanRequestCustomType(attrPath, schema, values)
		req.Private = privateProviderData

		return req
	}

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testCases := map[string]struct {
		block        fwschema.Block
		req          tfsdk.ModifyAttributePlanRequest
		expectedResp ModifyAttributePlanResponse
	}{
		"no-plan-modifiers": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:          types.StringType,
						Required:      true,
						PlanModifiers: nil,
					},
				},
				NestingMode:   tfsdk.BlockNestingModeList,
				PlanModifiers: nil,
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
								"nested_attr": types.String{Value: "testvalue"},
							},
						),
					},
				),
			},
		},
		"no-plan-modifiers-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:          types.StringType,
						Required:      true,
						PlanModifiers: nil,
					},
				},
				NestingMode:   tfsdk.BlockNestingModeList,
				PlanModifiers: nil,
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType(nil, nil),
				modifyAttributePlanValues{
					config: "testvalue",
					plan:   "testvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
			},
		},
		"block-modified": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:          types.StringType,
						Required:      true,
						PlanModifiers: nil,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					testBlockPlanModifierNullList{},
				},
			},
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
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Null: true,
				},
				Private: testEmptyProviderData,
			},
		},
		"block-modified-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:          types.StringType,
						Required:      true,
						PlanModifiers: nil,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					testBlockPlanModifierNullListCustomType{},
				},
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType([]tfsdk.AttributePlanModifier{
					testBlockPlanModifierNullList{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Null: true,
				},
				Private: testEmptyProviderData,
			},
		},
		"block-request-private": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:          types.StringType,
						Required:      true,
						PlanModifiers: nil,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					testBlockPlanModifierPrivateGet{},
				},
			},
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
								"nested_attr": types.String{Value: "TESTATTRONE"},
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-request-private-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:          types.StringType,
						Required:      true,
						PlanModifiers: nil,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					testBlockPlanModifierPrivateGet{},
				},
			},
			req: modifyAttributePlanWithPrivateRequestCustomType(
				path.Root("test"),
				schemaCustomType([]tfsdk.AttributePlanModifier{
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
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "TESTATTRONE"},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"block-response-private": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:          types.StringType,
						Required:      true,
						PlanModifiers: nil,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					testBlockPlanModifierPrivateSet{},
				},
			},
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
								"nested_attr": types.String{Value: "TESTATTRONE"},
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-response-private-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:          types.StringType,
						Required:      true,
						PlanModifiers: nil,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					testBlockPlanModifierPrivateSet{},
				},
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType([]tfsdk.AttributePlanModifier{
					testBlockPlanModifierPrivateSet{},
				}, nil),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "TESTATTRONE"},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"block-list-null-plan": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Null: true,
				},
				AttributeState: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Null: true,
				},
				Private: testProviderData,
			},
		},
		"block-list-null-plan-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Null: true,
				},
				AttributeState: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Null: true,
				},
				Private: testProviderData,
			},
		},
		"block-list-null-state": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributeState: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Null: true,
				},
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
								"nested_attr": types.String{Value: "testvalue"},
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-list-null-state-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributeState: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Null: true,
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"block-list-nested-private": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributeState: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
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
								"nested_attr": types.String{Value: "testvalue"},
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-list-nested-private-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributeState: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"block-set-null-plan": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Null: true,
				},
				AttributeState: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Null: true,
				},
				Private: testProviderData,
			},
		},
		"block-set-null-plan-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Null: true,
				},
				AttributeState: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Null: true,
				},
				Private: testProviderData,
			},
		},
		"block-set-null-state": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributeState: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Null: true,
				},
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
								"nested_attr": types.String{Value: "testvalue"},
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-set-null-state-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributeState: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Null: true,
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"block-set-nested-private": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributeState: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
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
								"nested_attr": types.String{Value: "testvalue"},
							},
						),
					},
				),
				Private: testProviderData,
			},
		},
		"block-set-nested-private-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				AttributeState: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "testvalue"},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"block-list-nested-block-list": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"id": {
						Type:     types.StringType,
						Computed: true,
						Optional: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.UseStateForUnknown(),
						},
					},
				},
				Blocks: map[string]tfsdk.Block{
					"list": {
						Attributes: map[string]tfsdk.Attribute{
							"nested_computed": {
								Type:     types.StringType,
								Computed: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.UseStateForUnknown(),
								},
							},
							"nested_required": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: tfsdk.BlockNestingModeList,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.List{
					ElemType: types.ObjectType{
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
					Elems: []attr.Value{
						types.Object{
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
							Attrs: map[string]attr.Value{
								"id": types.String{Value: "configvalue"},
								"list": types.List{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
									Elems: []attr.Value{
										types.Object{
											AttrTypes: map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
											Attrs: map[string]attr.Value{
												"nested_computed": types.String{Null: true},
												"nested_required": types.String{Value: "configvalue"},
											},
										},
									},
								},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.List{
					ElemType: types.ObjectType{
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
					Elems: []attr.Value{
						types.Object{
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
							Attrs: map[string]attr.Value{
								"id": types.String{Value: "one"},
								"list": types.List{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
									Elems: []attr.Value{
										types.Object{
											AttrTypes: map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
											Attrs: map[string]attr.Value{
												"nested_computed": types.String{Unknown: true},
												"nested_required": types.String{Value: "configvalue"},
											},
										},
									},
								},
							},
						},
					},
				},
				AttributeState: types.List{
					ElemType: types.ObjectType{
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
					Elems: []attr.Value{
						types.Object{
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
							Attrs: map[string]attr.Value{
								"id": types.String{Value: "one"},
								"list": types.List{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_computed": types.StringType,
											"nested_required": types.StringType,
										},
									},
									Elems: []attr.Value{
										types.Object{
											AttrTypes: map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
											Attrs: map[string]attr.Value{
												"nested_computed": types.String{Value: "statevalue"},
												"nested_required": types.String{Value: "configvalue"},
											},
										},
									},
								},
							},
						},
					},
				},
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
								"id": types.String{Value: "one"},
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
												"nested_computed": types.String{Value: "statevalue"},
												"nested_required": types.String{Value: "configvalue"},
											},
										),
									},
								),
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"block-list-nested-block-list-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"id": {
						Type:     types.StringType,
						Computed: true,
						Optional: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.UseStateForUnknown(),
						},
					},
				},
				Blocks: map[string]tfsdk.Block{
					"list": {
						Typ: CustomType{},
						Attributes: map[string]tfsdk.Attribute{
							"nested_computed": {
								Type:     types.StringType,
								Computed: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.UseStateForUnknown(),
								},
							},
							"nested_required": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: tfsdk.BlockNestingModeList,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"id": types.StringType,
								"list": types.ListType{
									ElemType: CustomType{
										types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
										},
									},
								},
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"id": types.StringType,
									"list": types.ListType{
										ElemType: CustomType{
											types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"nested_computed": types.StringType,
													"nested_required": types.StringType,
												},
											},
										},
									},
								},
								Attrs: map[string]attr.Value{
									"id": types.String{Value: "configvalue"},
									"list": types.List{
										ElemType: CustomType{
											types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"nested_computed": types.StringType,
													"nested_required": types.StringType,
												},
											},
										},
										Elems: []attr.Value{
											CustomValue{
												types.Object{
													AttrTypes: map[string]attr.Type{
														"nested_computed": types.StringType,
														"nested_required": types.StringType,
													},
													Attrs: map[string]attr.Value{
														"nested_computed": types.String{Null: true},
														"nested_required": types.String{Value: "configvalue"},
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
				AttributePath: path.Root("test"),
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"id": types.StringType,
								"list": types.ListType{
									ElemType: CustomType{
										types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
										},
									},
								},
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"id": types.StringType,
									"list": types.ListType{
										ElemType: CustomType{
											types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"nested_computed": types.StringType,
													"nested_required": types.StringType,
												},
											},
										},
									},
								},
								Attrs: map[string]attr.Value{
									"id": types.String{Value: "one"},
									"list": types.List{
										ElemType: CustomType{
											types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"nested_computed": types.StringType,
													"nested_required": types.StringType,
												},
											},
										},
										Elems: []attr.Value{
											CustomValue{
												types.Object{
													AttrTypes: map[string]attr.Type{
														"nested_computed": types.StringType,
														"nested_required": types.StringType,
													},
													Attrs: map[string]attr.Value{
														"nested_computed": types.String{Unknown: true},
														"nested_required": types.String{Value: "configvalue"},
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
				AttributeState: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"id": types.StringType,
								"list": types.ListType{
									ElemType: CustomType{
										types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
										},
									},
								},
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"id": types.StringType,
									"list": types.ListType{
										ElemType: CustomType{
											types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"nested_computed": types.StringType,
													"nested_required": types.StringType,
												},
											},
										},
									},
								},
								Attrs: map[string]attr.Value{
									"id": types.String{Value: "one"},
									"list": types.List{
										ElemType: CustomType{
											types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"nested_computed": types.StringType,
													"nested_required": types.StringType,
												},
											},
										},
										Elems: []attr.Value{
											CustomValue{
												types.Object{
													AttrTypes: map[string]attr.Type{
														"nested_computed": types.StringType,
														"nested_required": types.StringType,
													},
													Attrs: map[string]attr.Value{
														"nested_computed": types.String{Value: "statevalue"},
														"nested_required": types.String{Value: "configvalue"},
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
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"id": types.StringType,
								"list": types.ListType{
									ElemType: CustomType{
										types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_computed": types.StringType,
												"nested_required": types.StringType,
											},
										},
									},
								},
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"id": types.StringType,
									"list": types.ListType{
										ElemType: CustomType{
											types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"nested_computed": types.StringType,
													"nested_required": types.StringType,
												},
											},
										},
									},
								},
								Attrs: map[string]attr.Value{
									"id": types.String{Value: "one"},
									"list": types.List{
										ElemType: CustomType{
											types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"nested_computed": types.StringType,
													"nested_required": types.StringType,
												},
											},
										},
										Elems: []attr.Value{
											CustomValue{
												types.Object{
													AttrTypes: map[string]attr.Type{
														"nested_computed": types.StringType,
														"nested_required": types.StringType,
													},
													Attrs: map[string]attr.Value{
														"nested_computed": types.String{Value: "statevalue"},
														"nested_required": types.String{Value: "configvalue"},
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
				Private: testEmptyProviderData,
			},
		},
		"block-set-nested-usestateforunknown": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_computed": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							resource.UseStateForUnknown(),
						},
					},
					"nested_required": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Null: true},
								"nested_required": types.String{Value: "testvalue1"},
							},
						},
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Null: true},
								"nested_required": types.String{Value: "testvalue2"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Unknown: true},
								"nested_required": types.String{Value: "testvalue1"},
							},
						},
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Unknown: true},
								"nested_required": types.String{Value: "testvalue2"},
							},
						},
					},
				},
				AttributeState: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Value: "statevalue1"},
								"nested_required": types.String{Value: "testvalue1"},
							},
						},
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Value: "statevalue2"},
								"nested_required": types.String{Value: "testvalue2"},
							},
						},
					},
				},
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
								"nested_computed": types.String{Value: "statevalue1"},
								"nested_required": types.String{Value: "testvalue1"},
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							map[string]attr.Value{
								"nested_computed": types.String{Value: "statevalue2"},
								"nested_required": types.String{Value: "testvalue2"},
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"block-set-nested-usestateforunknown-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_computed": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							resource.UseStateForUnknown(),
						},
					},
					"nested_required": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_computed": types.String{Null: true},
									"nested_required": types.String{Value: "testvalue1"},
								},
							},
						},
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_computed": types.String{Null: true},
									"nested_required": types.String{Value: "testvalue2"},
								},
							},
						}},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_computed": types.String{Unknown: true},
									"nested_required": types.String{Value: "testvalue1"},
								},
							},
						},
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_computed": types.String{Unknown: true},
									"nested_required": types.String{Value: "testvalue2"},
								},
							},
						}},
				},
				AttributeState: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_computed": types.String{Value: "statevalue1"},
									"nested_required": types.String{Value: "testvalue1"},
								},
							},
						},
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_computed": types.String{Value: "statevalue2"},
									"nested_required": types.String{Value: "testvalue2"},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Set{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_computed": types.String{Value: "statevalue1"},
									"nested_required": types.String{Value: "testvalue1"},
								},
							},
						},
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_computed": types.String{Value: "statevalue2"},
									"nested_required": types.String{Value: "testvalue2"},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"block-single-null-plan": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSingle,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_attr": types.String{Value: "testvalue"},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Null: true,
				},
				AttributeState: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_attr": types.String{Value: "testvalue"},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Null: true,
				},
				Private: testProviderData,
			},
		},
		"block-single-null-plan-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSingle,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: "testvalue"},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Null: true,
					},
				},
				AttributeState: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: "testvalue"},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Null: true},
						},
						Null: true,
					},
				},
				Private: testProviderData,
			},
		},
		"block-single-null-state": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSingle,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_attr": types.String{Value: "testvalue"},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_attr": types.String{Value: "testvalue"},
					},
				},
				AttributeState: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Null: true,
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.String{Value: "testvalue"},
					},
				),
				Private: testProviderData,
			},
		},
		"block-single-null-state-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSingle,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: "testvalue"},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: "testvalue"},
						},
					},
				},
				AttributeState: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Null: true,
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: "testvalue"},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"block-single-nested-private": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSingle,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_attr": types.String{Value: "testvalue"},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_attr": types.String{Value: "testvalue"},
					},
				},
				AttributeState: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_attr": types.String{Value: "testvalue"},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_attr": types.StringType,
					},
					map[string]attr.Value{
						"nested_attr": types.String{Value: "testvalue"},
					},
				),
				Private: testProviderData,
			},
		},
		"block-single-nested-private-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							planmodifiers.TestAttrPlanPrivateModifierGet{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSingle,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: "testvalue"},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: "testvalue"},
						},
					},
				},
				AttributeState: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: "testvalue"},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_attr": types.String{Value: "testvalue"},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"block-single-nested-usestateforunknown": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_computed": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							resource.UseStateForUnknown(),
						},
					},
					"nested_required": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeSingle,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_computed": types.StringType,
						"nested_required": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_computed": types.String{Null: true},
						"nested_required": types.String{Value: "testvalue"},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_computed": types.StringType,
						"nested_required": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_computed": types.String{Unknown: true},
						"nested_required": types.String{Value: "testvalue"},
					},
				},
				AttributeState: types.Object{
					AttrTypes: map[string]attr.Type{
						"nested_computed": types.StringType,
						"nested_required": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_computed": types.String{Value: "statevalue"},
						"nested_required": types.String{Value: "testvalue"},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_computed": types.StringType,
						"nested_required": types.StringType,
					},
					map[string]attr.Value{
						"nested_computed": types.String{Value: "statevalue"},
						"nested_required": types.String{Value: "testvalue"},
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"block-single-nested-usestateforunknown-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_computed": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							resource.UseStateForUnknown(),
						},
					},
					"nested_required": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeSingle,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_computed": types.String{Null: true},
							"nested_required": types.String{Value: "testvalue"},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_computed": types.String{Unknown: true},
							"nested_required": types.String{Value: "testvalue"},
						},
					},
				},
				AttributeState: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_computed": types.String{Value: "statevalue"},
							"nested_required": types.String{Value: "testvalue"},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: CustomValue{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_computed": types.String{Value: "statevalue"},
							"nested_required": types.String{Value: "testvalue"},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"block-requires-replacement": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
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
								"nested_attr": types.String{Value: "newtestvalue"},
							},
						),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-requires-replacement-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType([]tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				}, nil),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "newtestvalue"},
								},
							},
						},
					},
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-requires-replacement-passthrough": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					testBlockPlanModifierNullList{},
				},
			},
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
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Null: true,
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-requires-replacement-passthrough-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					testBlockPlanModifierNullListCustomType{},
				},
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType([]tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					testBlockPlanModifierNullListCustomType{},
				}, nil),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Null: true,
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"block-requires-replacement-unset": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					planmodifiers.TestRequiresReplaceFalseModifier{},
				},
			},
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
								"nested_attr": types.String{Value: "newtestvalue"},
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"block-warnings": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
				},
			},
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
								"nested_attr": types.String{Value: "TESTDIAG"},
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"block-warnings-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
				},
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType([]tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
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
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "TESTDIAG"},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"block-error": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
				},
			},
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
			expectedResp: ModifyAttributePlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "TESTDIAG"},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"block-error-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
				},
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType([]tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
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
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "TESTDIAG"},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-modified": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifiers.TestAttrPlanValueModifierOne{},
							planmodifiers.TestAttrPlanValueModifierTwo{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
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
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-modified-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifiers.TestAttrPlanValueModifierOne{},
							planmodifiers.TestAttrPlanValueModifierTwo{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType(nil, []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanValueModifierOne{},
					planmodifiers.TestAttrPlanValueModifierTwo{},
				}),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "TESTATTRONE",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "MODIFIED_TWO"},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-requires-replacement": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
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
								"nested_attr": types.String{Value: "newtestvalue"},
							},
						),
					},
				),
				RequiresReplace: path.Paths{
					path.Root("test").AtListIndex(0).AtName("nested_attr"),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-requires-replacement-passthrough": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
							planmodifiers.TestAttrPlanValueModifierOne{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
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
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-requires-replacement-passthrough-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
							planmodifiers.TestAttrPlanValueModifierOne{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType(nil, []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					planmodifiers.TestAttrPlanValueModifierOne{},
				}),
				modifyAttributePlanValues{
					config: "TESTATTRONE",
					plan:   "TESTATTRONE",
					state:  "previousvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "TESTATTRTWO"},
								},
							},
						},
					},
				},
				RequiresReplace: path.Paths{
					path.Root("test").AtListIndex(0).AtName("nested_attr"),
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-requires-replacement-unset": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
							planmodifiers.TestRequiresReplaceFalseModifier{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
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
								"nested_attr": types.String{Value: "newtestvalue"},
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-requires-replacement-unset-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
							planmodifiers.TestRequiresReplaceFalseModifier{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType(nil, []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					planmodifiers.TestRequiresReplaceFalseModifier{},
				}),
				modifyAttributePlanValues{
					config: "newtestvalue",
					plan:   "newtestvalue",
					state:  "testvalue",
				},
			),
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "newtestvalue"},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-warnings": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifiers.TestWarningDiagModifier{},
							planmodifiers.TestWarningDiagModifier{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
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
								"nested_attr": types.String{Value: "TESTDIAG"},
							},
						),
					},
				),
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-warnings-custom-type": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifiers.TestWarningDiagModifier{},
							planmodifiers.TestWarningDiagModifier{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType(nil, []tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
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
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "TESTDIAG"},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-error": {
			block: tfsdk.Block{
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifiers.TestErrorDiagModifier{},
							planmodifiers.TestErrorDiagModifier{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
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
			expectedResp: ModifyAttributePlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							map[string]attr.Value{
								"nested_attr": types.String{Value: "TESTDIAG"},
							},
						),
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"nested-attribute-error-custom-attribute": {
			block: tfsdk.Block{
				Typ: CustomType{},
				Attributes: map[string]tfsdk.Attribute{
					"nested_attr": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifiers.TestErrorDiagModifier{},
							planmodifiers.TestErrorDiagModifier{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
			req: modifyAttributePlanRequestCustomType(
				path.Root("test"),
				schemaCustomType(nil, []tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
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
				AttributePlan: types.List{
					ElemType: CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
					Elems: []attr.Value{
						CustomValue{
							types.Object{
								AttrTypes: map[string]attr.Type{
									"nested_attr": types.StringType,
								},
								Attrs: map[string]attr.Value{
									"nested_attr": types.String{Value: "TESTDIAG"},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
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

type testBlockPlanModifierNullListCustomType struct{}

func (t testBlockPlanModifierNullListCustomType) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	_, ok := req.AttributePlan.(types.List)
	if !ok {
		return
	}

	resp.AttributePlan = types.List{
		ElemType: CustomType{
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"nested_attr": types.StringType,
				},
			},
		},
		Null: true,
	}
}

func (t testBlockPlanModifierNullListCustomType) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testBlockPlanModifierNullListCustomType) MarkdownDescription(ctx context.Context) string {
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
