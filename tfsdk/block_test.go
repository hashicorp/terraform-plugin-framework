package tfsdk

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ attr.Type = CustomType{}
)

type CustomType struct {
	types.ObjectType
}

func (t CustomType) WithAttributeTypes(typs map[string]attr.Type) attr.TypeWithAttributeTypes {
	return CustomType{
		types.ObjectType{
			AttrTypes: typs,
		},
	}
}

func (t CustomType) Equal(candidate attr.Type) bool {
	other, ok := candidate.(CustomType)
	if !ok {
		return false
	}
	if len(other.AttrTypes) != len(t.AttrTypes) {
		return false
	}
	for k, v := range t.AttrTypes {
		attrType, ok := other.AttrTypes[k]
		if !ok {
			return false
		}
		if !v.Equal(attrType) {
			return false
		}
	}
	return true
}

func TestBlockType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    Block
		expected attr.Type
	}{
		"NestingMode-List": {
			block: Block{
				Attributes: map[string]Attribute{
					"test_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				},
				Blocks: map[string]Block{
					"test_block": {
						Attributes: map[string]Attribute{
							"test_block_attribute": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeList,
					},
				},
				NestingMode: BlockNestingModeList,
			},
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_attribute": types.StringType,
						"test_block": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"test_block_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
		},
		"NestingMode-List-CustomType": {
			block: Block{
				Attributes: map[string]Attribute{
					"test_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				},
				Blocks: map[string]Block{
					"test_block": {
						Typ: CustomType{},
						Attributes: map[string]Attribute{
							"test_block_attribute": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeList,
					},
				},
				NestingMode: BlockNestingModeList,
			},
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_attribute": types.StringType,
						"test_block": types.ListType{
							ElemType: CustomType{
								types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"test_block_attribute": types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
		"NestingMode-Set": {
			block: Block{
				Attributes: map[string]Attribute{
					"test_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				},
				Blocks: map[string]Block{
					"test_block": {
						Attributes: map[string]Attribute{
							"test_block_attribute": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
				},
				NestingMode: BlockNestingModeSet,
			},
			expected: types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_attribute": types.StringType,
						"test_block": types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"test_block_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
		},
		"NestingMode-Set-CustomType": {
			block: Block{
				Attributes: map[string]Attribute{
					"test_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				},
				Blocks: map[string]Block{
					"test_block": {
						Typ: CustomType{},
						Attributes: map[string]Attribute{
							"test_block_attribute": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
				},
				NestingMode: BlockNestingModeSet,
			},
			expected: types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_attribute": types.StringType,
						"test_block": types.SetType{
							ElemType: CustomType{
								types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"test_block_attribute": types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
		"NestingMode-Single": {
			block: Block{
				Attributes: map[string]Attribute{
					"test_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				},
				Blocks: map[string]Block{
					"test_block": {
						Attributes: map[string]Attribute{
							"test_block_attribute": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
				NestingMode: BlockNestingModeSingle,
			},
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attribute": types.StringType,
					"test_block": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_block_attribute": types.StringType,
						},
					},
				},
			},
		},
		"NestingMode-Single-CustomType": {
			block: Block{
				Attributes: map[string]Attribute{
					"test_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				},
				Blocks: map[string]Block{
					"test_block": {
						Typ: CustomType{},
						Attributes: map[string]Attribute{
							"test_block_attribute": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
				NestingMode: BlockNestingModeSingle,
			},
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attribute": types.StringType,
					"test_block": CustomType{
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_block_attribute": types.StringType,
							},
						},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.block.Type()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
