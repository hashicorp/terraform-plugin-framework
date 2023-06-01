// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
		req  ValidateAttributeRequest
		resp ValidateAttributeResponse
	}{
		"deprecation-message-known": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								DeprecationMessage: "Use something else instead.",
								NestingMode:        fwschema.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
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
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								DeprecationMessage: "Use something else instead.",
								NestingMode:        fwschema.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"deprecation-message-unknown": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								DeprecationMessage: "Use something else instead.",
								NestingMode:        fwschema.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"warnings": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.BlockWithListValidators{
								Attributes: map[string]fwschema.Attribute{
									"nested_attr": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
								Validators: []validator.List{
									testvalidator.List{
										ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
											resp.Diagnostics.Append(testWarningDiagnostic1)
										},
									},
									testvalidator.List{
										ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
											resp.Diagnostics.Append(testWarningDiagnostic2)
										},
									},
								},
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testWarningDiagnostic1,
					testWarningDiagnostic2,
				},
			},
		},
		"errors": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.BlockWithListValidators{
								Attributes: map[string]fwschema.Attribute{
									"nested_attr": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
								Validators: []validator.List{
									testvalidator.List{
										ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
											resp.Diagnostics.Append(testErrorDiagnostic1)
										},
									},
									testvalidator.List{
										ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
											resp.Diagnostics.Append(testErrorDiagnostic2)
										},
									},
								},
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
					testErrorDiagnostic2,
				},
			},
		},
		"nested-attr-warnings": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringValidators{
											Required: true,
											Validators: []validator.String{
												testvalidator.String{
													ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
														resp.Diagnostics.Append(testWarningDiagnostic1)
													},
												},
												testvalidator.String{
													ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
														resp.Diagnostics.Append(testWarningDiagnostic2)
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testWarningDiagnostic1,
					testWarningDiagnostic2,
				},
			},
		},
		"nested-attr-errors": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringValidators{
											Required: true,
											Validators: []validator.String{
												testvalidator.String{
													ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
														resp.Diagnostics.Append(testErrorDiagnostic1)
													},
												},
												testvalidator.String{
													ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
														resp.Diagnostics.Append(testErrorDiagnostic2)
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
					testErrorDiagnostic2,
				},
			},
		},
		"nested-attr-type-with-validate-error": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     testtypes.StringTypeWithValidateError{},
											Required: true,
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestErrorDiagnostic(path.Root("test").AtListIndex(0).AtName("nested_attr")),
				},
			},
		},
		"nested-attr-type-with-validate-warning": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     testtypes.StringTypeWithValidateWarning{},
											Required: true,
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestWarningDiagnostic(path.Root("test").AtListIndex(0).AtName("nested_attr")),
				},
			},
		},
		"list-no-validation": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
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
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"list-validation": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringValidators{
											Required: true,
											Validators: []validator.String{
												testvalidator.String{
													ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
														resp.Diagnostics.Append(testErrorDiagnostic1)
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"set-no-validation": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"set-validation": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringValidators{
											Required: true,
											Validators: []validator.String{
												testvalidator.String{
													ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
														resp.Diagnostics.Append(testErrorDiagnostic1)
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"single-no-validation": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"single-validation": {
			req: ValidateAttributeRequest{
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
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringValidators{
											Required: true,
											Validators: []validator.String{
												testvalidator.String{
													ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
														resp.Diagnostics.Append(testErrorDiagnostic1)
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got ValidateAttributeResponse
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
		request  ValidateAttributeRequest
		response *ValidateAttributeResponse
		expected *ValidateAttributeResponse
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{
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
			expected: &ValidateAttributeResponse{
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
		request  ValidateAttributeRequest
		response *ValidateAttributeResponse
		expected *ValidateAttributeResponse
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
			request: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{"testattr": types.StringType},
					map[string]attr.Value{"testattr": types.StringValue("test")},
				),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{"testattr": types.StringType},
					map[string]attr.Value{"testattr": types.StringValue("test")},
				),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{"testattr": types.StringType},
					map[string]attr.Value{"testattr": types.StringValue("test")},
				),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.ObjectValueMust(
					map[string]attr.Type{"testattr": types.StringType},
					map[string]attr.Value{"testattr": types.StringValue("test")},
				),
			},
			response: &ValidateAttributeResponse{
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
			expected: &ValidateAttributeResponse{
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
		request  ValidateAttributeRequest
		response *ValidateAttributeResponse
		expected *ValidateAttributeResponse
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
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
			request: ValidateAttributeRequest{
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
			response: &ValidateAttributeResponse{
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
			expected: &ValidateAttributeResponse{
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

func TestNestedBlockObjectValidateObject(t *testing.T) {
	t.Parallel()

	testAttributeConfig := types.ObjectValueMust(
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
				map[string]attr.Type{"testblockattr": types.StringType},
				map[string]attr.Value{"testblockattr": types.StringValue("testvalue")},
			),
		},
	)
	testConfig := tfsdk.Config{
		Raw: tftypes.NewValue(
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
		),
		Schema: testschema.Schema{
			Blocks: map[string]fwschema.Block{
				"test": testschema.BlockWithObjectValidators{
					Attributes: map[string]fwschema.Attribute{
						"testattr": testschema.AttributeWithStringValidators{
							Required: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"testblock": testschema.BlockWithObjectValidators{
							Attributes: map[string]fwschema.Attribute{
								"testblockattr": testschema.AttributeWithStringValidators{
									Required: true,
								},
							},
						},
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		object   fwschema.NestedBlockObject
		request  ValidateAttributeRequest
		response *ValidateAttributeResponse
		expected *ValidateAttributeResponse
	}{
		"request-path": {
			object: testschema.NestedBlockObjectWithValidators{
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: testAttributeConfig,
				Config:          testConfig,
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			object: testschema.NestedBlockObjectWithValidators{
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
			request: ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig:         testAttributeConfig,
				Config:                  testConfig,
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			object: testschema.NestedBlockObjectWithValidators{
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: testAttributeConfig,
				Config:          testConfig,
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-configvalue": {
			object: testschema.NestedBlockObjectWithValidators{
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
							got := req.ConfigValue
							expected := testAttributeConfig

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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: testAttributeConfig,
				Config:          testConfig,
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			object: testschema.NestedBlockObjectWithValidators{
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: testAttributeConfig,
				Config:          testConfig,
			},
			response: &ValidateAttributeResponse{
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
			expected: &ValidateAttributeResponse{
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
		"nested-attributes-validation": {
			object: testschema.NestedBlockObjectWithValidators{
				Attributes: map[string]fwschema.Attribute{
					"testattr": testschema.AttributeWithStringValidators{
						Required: true,
						Validators: []validator.String{
							testvalidator.String{
								ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
									resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
									resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
								},
							},
						},
					},
				},
			},
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: testAttributeConfig,
				Config:          testConfig,
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{
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
		"nested-blocks-validation": {
			object: testschema.NestedBlockObjectWithValidators{
				Blocks: map[string]fwschema.Block{
					"testblock": testschema.BlockWithObjectValidators{
						Validators: []validator.Object{
							testvalidator.Object{
								ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
									resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
									resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
								},
							},
						},
					},
				},
			},
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: testAttributeConfig,
				Config:          testConfig,
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{
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
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			NestedBlockObjectValidate(context.Background(), testCase.object, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
