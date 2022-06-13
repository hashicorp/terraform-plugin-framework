package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBlockAttributeType(t *testing.T) {
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
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.block.attributeType()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestBlockTerraformType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    Block
		expected tftypes.Type
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
			expected: tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
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
			expected: tftypes.Set{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
								},
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

			got := testCase.block.terraformType(context.Background())

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
