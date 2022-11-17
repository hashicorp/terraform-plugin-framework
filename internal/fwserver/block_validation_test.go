package fwserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testvalidator"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBlockValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req  tfsdk.ValidateAttributeRequest
		resp tfsdk.ValidateAttributeResponse
	}{
		"deprecation-message-known": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								DeprecationMessage: "Use something else instead.",
								NestingMode:        tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("test"),
						"Block Deprecated",
						"Use something else instead.",
					),
				},
			},
		},
		"deprecation-message-null": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								DeprecationMessage: "Use something else instead.",
								NestingMode:        tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"deprecation-message-unknown": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
								tftypes.UnknownValue,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								DeprecationMessage: "Use something else instead.",
								NestingMode:        tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"warnings": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
								Validators: []tfsdk.AttributeValidator{
									testWarningAttributeValidator{},
									testWarningAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testWarningDiagnostic1,
					testWarningDiagnostic2,
				},
			},
		},
		"errors": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
								Validators: []tfsdk.AttributeValidator{
									testErrorAttributeValidator{},
									testErrorAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
					testErrorDiagnostic2,
				},
			},
		},
		"nested-attr-warnings": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []tfsdk.AttributeValidator{
											testWarningAttributeValidator{},
											testWarningAttributeValidator{},
										},
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testWarningDiagnostic1,
					testWarningDiagnostic2,
				},
			},
		},
		"nested-attr-errors": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []tfsdk.AttributeValidator{
											testErrorAttributeValidator{},
											testErrorAttributeValidator{},
										},
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
					testErrorDiagnostic2,
				},
			},
		},
		"nested-attr-type-with-validate-error": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     testtypes.StringTypeWithValidateError{},
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestErrorDiagnostic(path.Root("test").AtListIndex(0).AtName("nested_attr")),
				},
			},
		},
		"nested-attr-type-with-validate-warning": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     testtypes.StringTypeWithValidateWarning{},
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestWarningDiagnostic(path.Root("test").AtListIndex(0).AtName("nested_attr")),
				},
			},
		},
		"list-no-validation": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"list-validation": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []tfsdk.AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"list-maxitems-validation-known-invalid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue2"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Extra Block Configuration",
						"The configuration should declare a maximum of 1 block, however 2 blocks were configured.",
					),
				},
			},
		},
		"list-maxitems-validation-known-valid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"list-maxitems-validation-null": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"list-maxitems-validation-null-values": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
										nil,
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										nil,
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Extra Block Configuration",
						"The configuration should declare a maximum of 1 block, however 2 blocks were configured.",
					),
				},
			},
		},
		"list-maxitems-validation-unknown": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
								tftypes.UnknownValue,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"list-maxitems-validation-unknown-values": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
										tftypes.UnknownValue,
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										tftypes.UnknownValue,
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Extra Block Configuration",
						"The configuration should declare a maximum of 1 block, however 2 blocks were configured.",
					),
				},
			},
		},
		"list-minitems-validation-known-invalid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    2,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Block Configuration",
						"The configuration should declare a minimum of 2 blocks, however 1 block was configured.",
					),
				},
			},
		},
		"list-minitems-validation-known-valid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    1,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"list-minitems-validation-null": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    1,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Block Configuration",
						"The configuration should declare a minimum of 1 block, however 0 blocks were configured.",
					),
				},
			},
		},
		"list-minitems-validation-null-values": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
										nil,
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    2,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Block Configuration",
						"The configuration should declare a minimum of 2 blocks, however 1 block was configured.",
					),
				},
			},
		},
		"list-minitems-validation-unknown": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
								tftypes.UnknownValue,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    1,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"list-minitems-validation-unknown-values": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
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
										tftypes.UnknownValue,
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    2,
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Block Configuration",
						"The configuration should declare a minimum of 2 blocks, however 1 block was configured.",
					),
				},
			},
		},
		"set-no-validation": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"set-validation": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []tfsdk.AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								},
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"set-maxitems-validation-known-invalid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue2"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Extra Block Configuration",
						"The configuration should declare a maximum of 1 block, however 2 blocks were configured.",
					),
				},
			},
		},
		"set-maxitems-validation-known-valid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"set-maxitems-validation-null": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								nil,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"set-maxitems-validation-null-values": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
										nil,
									),
									// Must not be a duplicate value.
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Extra Block Configuration",
						"The configuration should declare a maximum of 1 block, however 2 blocks were configured.",
					),
				},
			},
		},
		"set-maxitems-validation-unknown": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"set-maxitems-validation-unknown-values": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
										tftypes.UnknownValue,
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										tftypes.UnknownValue,
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Extra Block Configuration",
						"The configuration should declare a maximum of 1 block, however 2 blocks were configured.",
					),
				},
			},
		},
		"set-minitems-validation-known-invalid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    2,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Block Configuration",
						"The configuration should declare a minimum of 2 blocks, however 1 block was configured.",
					),
				},
			},
		},
		"set-minitems-validation-known-valid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    1,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"set-minitems-validation-null": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								nil,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    1,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Block Configuration",
						"The configuration should declare a minimum of 1 block, however 0 blocks were configured.",
					),
				},
			},
		},
		"set-minitems-validation-null-values": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
										nil,
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    2,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Block Configuration",
						"The configuration should declare a minimum of 2 blocks, however 1 block was configured.",
					),
				},
			},
		},
		"set-minitems-validation-unknown": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    1,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"set-minitems-validation-unknown-values": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
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
								tftypes.Set{
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
										tftypes.UnknownValue,
									),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    2,
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Block Configuration",
						"The configuration should declare a minimum of 2 blocks, however 1 block was configured.",
					),
				},
			},
		},
		"single-no-validation": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"single-validation": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []tfsdk.AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								},
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"single-maxitems-validation-known-valid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"single-maxitems-validation-null": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								nil,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"single-maxitems-validation-unknown": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MaxItems:    1,
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"single-minitems-validation-known-valid": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"nested_attr": tftypes.NewValue(tftypes.String, "testvalue1"),
								},
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    1,
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
		"single-minitems-validation-null": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								nil,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    1,
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Block Configuration",
						"The configuration should declare a minimum of 1 block, however 0 blocks were configured.",
					),
				},
			},
		},
		"single-minitems-validation-unknown": {
			req: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								tftypes.UnknownValue,
							),
						},
					),
					Schema: tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test": {
								Attributes: map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Optional: true,
									},
								},
								MinItems:    1,
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: tfsdk.ValidateAttributeResponse{},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got tfsdk.ValidateAttributeResponse
			block, ok := tc.req.Config.Schema.GetBlocks()["test"]

			if !ok {
				t.Fatalf("Unexpected error getting schema block")
			}

			BlockValidate(context.Background(), block, tc.req, &got)

			if diff := cmp.Diff(got, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestBlockValidateList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    fwxschema.BlockWithListValidators
		request  tfsdk.ValidateAttributeRequest
		response *tfsdk.ValidateAttributeResponse
		expected *tfsdk.ValidateAttributeResponse
	}{
		"request-path": {
			block: testschema.BlockWithListValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.List{
					testvalidator.List{
						ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			block: testschema.BlockWithListValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.List{
					testvalidator.List{
						ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"request-config": {
			block: testschema.BlockWithListValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.List{
					testvalidator.List{
						ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.List{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"testattr": tftypes.String,
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
														"testattr": tftypes.String,
													},
												},
											},
											[]tftypes.Value{
												tftypes.NewValue(
													tftypes.Object{
														AttributeTypes: map[string]tftypes.Type{
															"testattr": tftypes.String,
														},
													},
													map[string]tftypes.Value{
														"testattr": tftypes.NewValue(tftypes.String, "test"),
													},
												),
											},
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"testattr": tftypes.String,
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
											"testattr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"testattr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"testattr": tftypes.NewValue(tftypes.String, "test"),
										},
									),
								},
							),
						},
					),
				},
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"request-configvalue": {
			block: testschema.BlockWithListValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.List{
					testvalidator.List{
						ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
							got := req.ConfigValue
							expected := types.ListValueMust(
								types.ObjectType{
									AttrTypes: map[string]attr.Type{"testattr": types.StringType},
								},
								[]attr.Value{
									types.ObjectValueMust(
										map[string]attr.Type{"testattr": types.StringType},
										map[string]attr.Value{"testattr": types.StringValue("test")},
									),
								},
							)

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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			block: testschema.BlockWithListValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.List{
					testvalidator.List{
						ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{
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
			expected: &tfsdk.ValidateAttributeResponse{
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
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			BlockValidateList(context.Background(), testCase.block, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestBlockValidateObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    fwxschema.BlockWithObjectValidators
		request  tfsdk.ValidateAttributeRequest
		response *tfsdk.ValidateAttributeResponse
		expected *tfsdk.ValidateAttributeResponse
	}{
		"request-path": {
			block: testschema.BlockWithObjectValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{"testattr": types.StringType},
					map[string]attr.Value{"testattr": types.StringValue("test")},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			block: testschema.BlockWithObjectValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{"testattr": types.StringType},
					map[string]attr.Value{"testattr": types.StringValue("test")},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"request-config": {
			block: testschema.BlockWithObjectValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
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
												"testattr": tftypes.NewValue(tftypes.String, "test"),
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{"testattr": types.StringType},
					map[string]attr.Value{"testattr": types.StringValue("test")},
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
									"testattr": tftypes.NewValue(tftypes.String, "test"),
								},
							),
						},
					),
				},
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"request-configvalue": {
			block: testschema.BlockWithObjectValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
							got := req.ConfigValue
							expected := types.ObjectValueMust(
								map[string]attr.Type{"testattr": types.StringType},
								map[string]attr.Value{"testattr": types.StringValue("test")},
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{"testattr": types.StringType},
					map[string]attr.Value{"testattr": types.StringValue("test")},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			block: testschema.BlockWithObjectValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{"testattr": types.StringType},
					map[string]attr.Value{"testattr": types.StringValue("test")},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{
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
			expected: &tfsdk.ValidateAttributeResponse{
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
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			BlockValidateObject(context.Background(), testCase.block, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestBlockValidateSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    fwxschema.BlockWithSetValidators
		request  tfsdk.ValidateAttributeRequest
		response *tfsdk.ValidateAttributeResponse
		expected *tfsdk.ValidateAttributeResponse
	}{
		"request-path": {
			block: testschema.BlockWithSetValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Set{
					testvalidator.Set{
						ValidateSetMethod: func(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			block: testschema.BlockWithSetValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Set{
					testvalidator.Set{
						ValidateSetMethod: func(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"request-config": {
			block: testschema.BlockWithSetValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Set{
					testvalidator.Set{
						ValidateSetMethod: func(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Set{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"testattr": tftypes.String,
													},
												},
											},
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(
											tftypes.Set{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"testattr": tftypes.String,
													},
												},
											},
											[]tftypes.Value{
												tftypes.NewValue(
													tftypes.Object{
														AttributeTypes: map[string]tftypes.Type{
															"testattr": tftypes.String,
														},
													},
													map[string]tftypes.Value{
														"testattr": tftypes.NewValue(tftypes.String, "test"),
													},
												),
											},
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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"testattr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"testattr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"testattr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"testattr": tftypes.NewValue(tftypes.String, "test"),
										},
									),
								},
							),
						},
					),
				},
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"request-configvalue": {
			block: testschema.BlockWithSetValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Set{
					testvalidator.Set{
						ValidateSetMethod: func(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
							got := req.ConfigValue
							expected := types.SetValueMust(
								types.ObjectType{
									AttrTypes: map[string]attr.Type{"testattr": types.StringType},
								},
								[]attr.Value{
									types.ObjectValueMust(
										map[string]attr.Type{"testattr": types.StringType},
										map[string]attr.Value{"testattr": types.StringValue("test")},
									),
								},
							)

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
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{},
			expected: &tfsdk.ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			block: testschema.BlockWithSetValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{},
				},
				Validators: []validator.Set{
					testvalidator.Set{
						ValidateSetMethod: func(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: tfsdk.ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"testattr": types.StringType},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{"testattr": types.StringType},
							map[string]attr.Value{"testattr": types.StringValue("test")},
						),
					},
				),
			},
			response: &tfsdk.ValidateAttributeResponse{
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
			expected: &tfsdk.ValidateAttributeResponse{
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
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			BlockValidateSet(context.Background(), testCase.block, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestBlockMaxItemsDiagnostic(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		maxItems int64
		elements int
		expected diag.Diagnostic
	}{
		"1-maxitems-2-elements": {
			path:     path.Root("test"),
			maxItems: 1,
			elements: 2,
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				"Extra Block Configuration",
				"The configuration should declare a maximum of 1 block, however 2 blocks were configured.",
			),
		},
		"2-maxitems-3-elements": {
			path:     path.Root("test"),
			maxItems: 2,
			elements: 3,
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				"Extra Block Configuration",
				"The configuration should declare a maximum of 2 blocks, however 3 blocks were configured.",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := blockMaxItemsDiagnostic(testCase.path, testCase.maxItems, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestBlockMinItemsDiagnostic(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		minItems int64
		elements int
		expected diag.Diagnostic
	}{
		"1-minitems-0-elements": {
			path:     path.Root("test"),
			minItems: 1,
			elements: 0,
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				"Missing Block Configuration",
				"The configuration should declare a minimum of 1 block, however 0 blocks were configured.",
			),
		},
		"2-minitems-1-element": {
			path:     path.Root("test"),
			minItems: 2,
			elements: 1,
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				"Missing Block Configuration",
				"The configuration should declare a minimum of 2 blocks, however 1 block was configured.",
			),
		},
		"3-minitems-2-elements": {
			path:     path.Root("test"),
			minItems: 3,
			elements: 2,
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				"Missing Block Configuration",
				"The configuration should declare a minimum of 3 blocks, however 2 blocks were configured.",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := blockMinItemsDiagnostic(testCase.path, testCase.minItems, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
