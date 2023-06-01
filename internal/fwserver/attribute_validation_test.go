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
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testvalidator"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAttributeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req  ValidateAttributeRequest
		resp ValidateAttributeResponse
	}{
		"missing-required-optional-and-computed": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type: types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Invalid Attribute Definition",
						"Attribute missing Required, Optional, or Computed definition. This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"config-error": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"List Type Validation Error",
						"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"expected List value, received tftypes.Value with value: tftypes.String<\"testvalue\">",
					),
				},
			},
		},
		"config-computed-null": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Computed: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"config-computed-unknown": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Computed: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Invalid Configuration for Read-Only Attribute",
						"Cannot set value for this attribute as the provider has marked it as read-only. Remove the configuration line setting the value.\n\n"+
							"Refer to the provider documentation or contact the provider developers for additional information about configurable and read-only attributes that are supported.",
					),
				},
			},
		},
		"config-computed-value": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Computed: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Invalid Configuration for Read-Only Attribute",
						"Cannot set value for this attribute as the provider has marked it as read-only. Remove the configuration line setting the value.\n\n"+
							"Refer to the provider documentation or contact the provider developers for additional information about configurable and read-only attributes that are supported.",
					),
				},
			},
		},
		"config-optional-computed-null": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Computed: true,
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"config-optional-computed-unknown": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Computed: true,
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"config-optional-computed-value": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Computed: true,
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"config-required-null": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Missing Configuration for Required Attribute",
						"Must set a configuration value for the test attribute as the provider has marked it as required.\n\n"+
							"Refer to the provider documentation or contact the provider developers for additional information about configurable attributes that are required.",
					),
				},
			},
		},
		"config-required-unknown": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"config-required-value": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"deprecation-message-known": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:               types.StringType,
								Optional:           true,
								DeprecationMessage: "Use something else instead.",
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						path.Root("test"),
						"Attribute Deprecated",
						"Use something else instead.",
					),
				},
			},
		},
		"deprecation-message-null": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:               types.StringType,
								Optional:           true,
								DeprecationMessage: "Use something else instead.",
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
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:               types.StringType,
								Optional:           true,
								DeprecationMessage: "Use something else instead.",
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
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringValidators{
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
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringValidators{
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
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
					testErrorDiagnostic2,
				},
			},
		},
		"type-with-validate-error": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     testtypes.StringTypeWithValidateError{},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestErrorDiagnostic(path.Root("test")),
				},
			},
		},
		"type-with-validate-warning": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     testtypes.StringTypeWithValidateWarning{},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestWarningDiagnostic(path.Root("test")),
				},
			},
		},
		"nested-attr-list-no-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
								Required:    true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-custom-attr-list-no-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
								Type: testtypes.ListNestedAttributesCustomTypeType{
									ListType: types.ListType{
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_attr": types.StringType,
											},
										},
									},
								},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-list-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
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
								NestingMode: fwschema.NestingModeList,
								Required:    true,
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
		"nested-custom-attr-list-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
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
								NestingMode: fwschema.NestingModeList,
								Type: testtypes.ListNestedAttributesCustomTypeType{
									ListType: types.ListType{
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_attr": types.StringType,
											},
										},
									},
								},
								Required: true,
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
		"nested-attr-map-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{
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
								tftypes.Map{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeMap,
								Required:    true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-custom-attr-map-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{
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
								tftypes.Map{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeMap,
								Type: testtypes.MapNestedAttributesCustomTypeType{
									MapType: types.MapType{
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_attr": types.StringType,
											},
										},
									},
								},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-map-validation": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{
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
								tftypes.Map{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
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
								NestingMode: fwschema.NestingModeMap,
								Required:    true,
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
		"nested-custom-attr-map-validation": {
			req: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{
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
								tftypes.Map{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
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
								NestingMode: fwschema.NestingModeMap,
								Type: testtypes.MapNestedAttributesCustomTypeType{
									MapType: types.MapType{
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_attr": types.StringType,
											},
										},
									},
								},
								Required: true,
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
		"nested-attr-set-no-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
								Required:    true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-custom-attr-set-no-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
								Type: testtypes.SetNestedAttributesCustomTypeType{
									SetType: types.SetType{
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_attr": types.StringType,
											},
										},
									},
								},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-set-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
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
								NestingMode: fwschema.NestingModeSet,
								Required:    true,
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
		"nested-custom-attr-set-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
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
								NestingMode: fwschema.NestingModeSet,
								Type: testtypes.SetNestedAttributesCustomTypeType{
									SetType: types.SetType{
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"nested_attr": types.StringType,
											},
										},
									},
								},
								Required: true,
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
		"nested-attr-single-no-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSingle,
								Required:    true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-custom-attr-single-no-validation": {
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.Attribute{
											Type:     types.StringType,
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSingle,
								Required:    true,
								Type: testtypes.SingleNestedAttributesCustomTypeType{
									ObjectType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_attr": types.StringType,
										},
									},
								},
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-single-validation": {
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
						}, map[string]tftypes.Value{
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
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
								NestingMode: fwschema.NestingModeSingle,
								Required:    true,
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
		"nested-custom-attr-single-validation": {
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
						}, map[string]tftypes.Value{
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
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
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
								NestingMode: fwschema.NestingModeSingle,
								Required:    true,
								Type: testtypes.SingleNestedAttributesCustomTypeType{
									ObjectType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"nested_attr": types.StringType,
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
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var got ValidateAttributeResponse

			attribute, diags := tc.req.Config.Schema.AttributeAtPath(ctx, tc.req.AttributePath)

			if diags.HasError() {
				t.Fatalf("Unexpected diagnostics: %s", diags)
			}

			AttributeValidate(ctx, attribute, tc.req, &got)

			if diff := cmp.Diff(got, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestAttributeValidateBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithBoolValidators
		request   ValidateAttributeRequest
		response  *ValidateAttributeResponse
		expected  *ValidateAttributeResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithBoolValidators{
				Validators: []validator.Bool{
					testvalidator.Bool{
						ValidateBoolMethod: func(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithBoolValidators{
				Validators: []validator.Bool{
					testvalidator.Bool{
						ValidateBoolMethod: func(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig:         types.BoolValue(true),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			attribute: testschema.AttributeWithBoolValidators{
				Validators: []validator.Bool{
					testvalidator.Bool{
						ValidateBoolMethod: func(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithBoolValidators{
				Validators: []validator.Bool{
					testvalidator.Bool{
						ValidateBoolMethod: func(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithBoolValidators{
				Validators: []validator.Bool{
					testvalidator.Bool{
						ValidateBoolMethod: func(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.BoolValue(true),
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

			AttributeValidateBool(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributeValidateFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithFloat64Validators
		request   ValidateAttributeRequest
		response  *ValidateAttributeResponse
		expected  *ValidateAttributeResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithFloat64Validators{
				Validators: []validator.Float64{
					testvalidator.Float64{
						ValidateFloat64Method: func(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithFloat64Validators{
				Validators: []validator.Float64{
					testvalidator.Float64{
						ValidateFloat64Method: func(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
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
			request: ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig:         types.Float64Value(1.2),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			attribute: testschema.AttributeWithFloat64Validators{
				Validators: []validator.Float64{
					testvalidator.Float64{
						ValidateFloat64Method: func(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithFloat64Validators{
				Validators: []validator.Float64{
					testvalidator.Float64{
						ValidateFloat64Method: func(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithFloat64Validators{
				Validators: []validator.Float64{
					testvalidator.Float64{
						ValidateFloat64Method: func(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Float64Value(1.2),
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

			AttributeValidateFloat64(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributeValidateInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithInt64Validators
		request   ValidateAttributeRequest
		response  *ValidateAttributeResponse
		expected  *ValidateAttributeResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithInt64Validators{
				Validators: []validator.Int64{
					testvalidator.Int64{
						ValidateInt64Method: func(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(123),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithInt64Validators{
				Validators: []validator.Int64{
					testvalidator.Int64{
						ValidateInt64Method: func(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
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
			request: ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig:         types.Int64Value(123),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			attribute: testschema.AttributeWithInt64Validators{
				Validators: []validator.Int64{
					testvalidator.Int64{
						ValidateInt64Method: func(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.Number,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.Number, 123),
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(123),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.Number, 123),
						},
					),
				},
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithInt64Validators{
				Validators: []validator.Int64{
					testvalidator.Int64{
						ValidateInt64Method: func(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
							got := req.ConfigValue
							expected := types.Int64Value(123)

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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(123),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithInt64Validators{
				Validators: []validator.Int64{
					testvalidator.Int64{
						ValidateInt64Method: func(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.Int64Value(123),
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

			AttributeValidateInt64(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributeValidateList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithListValidators
		request   ValidateAttributeRequest
		response  *ValidateAttributeResponse
		expected  *ValidateAttributeResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithListValidators{
				ElementType: types.StringType,
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
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithListValidators{
				ElementType: types.StringType,
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
				AttributeConfig:         types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			attribute: testschema.AttributeWithListValidators{
				ElementType: types.StringType,
				Validators: []validator.List{
					testvalidator.List{
						ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
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
											[]tftypes.Value{
												tftypes.NewValue(tftypes.String, "test"),
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
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
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
								[]tftypes.Value{
									tftypes.NewValue(tftypes.String, "test"),
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
			attribute: testschema.AttributeWithListValidators{
				ElementType: types.StringType,
				Validators: []validator.List{
					testvalidator.List{
						ValidateListMethod: func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
							got := req.ConfigValue
							expected := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")})

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
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithListValidators{
				ElementType: types.StringType,
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
				AttributePath:   path.Root("test"),
				AttributeConfig: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
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

			AttributeValidateList(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributeValidateMap(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithMapValidators
		request   ValidateAttributeRequest
		response  *ValidateAttributeResponse
		expected  *ValidateAttributeResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithMapValidators{
				ElementType: types.StringType,
				Validators: []validator.Map{
					testvalidator.Map{
						ValidateMapMethod: func(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{"testkey": types.StringValue("testvalue")},
				),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithMapValidators{
				ElementType: types.StringType,
				Validators: []validator.Map{
					testvalidator.Map{
						ValidateMapMethod: func(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{"testkey": types.StringValue("testvalue")},
				),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			attribute: testschema.AttributeWithMapValidators{
				ElementType: types.StringType,
				Validators: []validator.Map{
					testvalidator.Map{
						ValidateMapMethod: func(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{"testkey": types.StringValue("testvalue")},
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithMapValidators{
				ElementType: types.StringType,
				Validators: []validator.Map{
					testvalidator.Map{
						ValidateMapMethod: func(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
							got := req.ConfigValue
							expected := types.MapValueMust(
								types.StringType,
								map[string]attr.Value{"testkey": types.StringValue("testvalue")},
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
			request: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{"testkey": types.StringValue("testvalue")},
				),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithMapValidators{
				ElementType: types.StringType,
				Validators: []validator.Map{
					testvalidator.Map{
						ValidateMapMethod: func(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ValidateAttributeRequest{
				AttributePath: path.Root("test"),
				AttributeConfig: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{"testkey": types.StringValue("testvalue")},
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

			AttributeValidateMap(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributeValidateNumber(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithNumberValidators
		request   ValidateAttributeRequest
		response  *ValidateAttributeResponse
		expected  *ValidateAttributeResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithNumberValidators{
				Validators: []validator.Number{
					testvalidator.Number{
						ValidateNumberMethod: func(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1.2)),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithNumberValidators{
				Validators: []validator.Number{
					testvalidator.Number{
						ValidateNumberMethod: func(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig:         types.NumberValue(big.NewFloat(1.2)),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			attribute: testschema.AttributeWithNumberValidators{
				Validators: []validator.Number{
					testvalidator.Number{
						ValidateNumberMethod: func(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1.2)),
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
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithNumberValidators{
				Validators: []validator.Number{
					testvalidator.Number{
						ValidateNumberMethod: func(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
							got := req.ConfigValue
							expected := types.NumberValue(big.NewFloat(1.2))

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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1.2)),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithNumberValidators{
				Validators: []validator.Number{
					testvalidator.Number{
						ValidateNumberMethod: func(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.NumberValue(big.NewFloat(1.2)),
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

			AttributeValidateNumber(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributeValidateObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithObjectValidators
		request   ValidateAttributeRequest
		response  *ValidateAttributeResponse
		expected  *ValidateAttributeResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithObjectValidators{
				AttributeTypes: map[string]attr.Type{
					"testattr": types.StringType,
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
					map[string]attr.Value{"testattr": types.StringValue("testvalue")},
				),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithObjectValidators{
				AttributeTypes: map[string]attr.Type{
					"testattr": types.StringType,
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
					map[string]attr.Value{"testattr": types.StringValue("testvalue")},
				),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			attribute: testschema.AttributeWithObjectValidators{
				AttributeTypes: map[string]attr.Type{
					"testattr": types.StringType,
				},
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
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
					map[string]attr.Value{"testattr": types.StringValue("testvalue")},
				),
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
								map[string]tftypes.Value{
									"testattr": tftypes.NewValue(tftypes.String, "testvalue"),
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
			attribute: testschema.AttributeWithObjectValidators{
				AttributeTypes: map[string]attr.Type{
					"testattr": types.StringType,
				},
				Validators: []validator.Object{
					testvalidator.Object{
						ValidateObjectMethod: func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
							got := req.ConfigValue
							expected := types.ObjectValueMust(
								map[string]attr.Type{"testattr": types.StringType},
								map[string]attr.Value{"testattr": types.StringValue("testvalue")},
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
					map[string]attr.Value{"testattr": types.StringValue("testvalue")},
				),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithObjectValidators{
				AttributeTypes: map[string]attr.Type{
					"testattr": types.StringType,
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
					map[string]attr.Value{"testattr": types.StringValue("testvalue")},
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

			AttributeValidateObject(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributeValidateSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithSetValidators
		request   ValidateAttributeRequest
		response  *ValidateAttributeResponse
		expected  *ValidateAttributeResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithSetValidators{
				ElementType: types.StringType,
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
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithSetValidators{
				ElementType: types.StringType,
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
				AttributeConfig:         types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			attribute: testschema.AttributeWithSetValidators{
				ElementType: types.StringType,
				Validators: []validator.Set{
					testvalidator.Set{
						ValidateSetMethod: func(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
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
											[]tftypes.Value{
												tftypes.NewValue(tftypes.String, "test"),
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
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
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
								[]tftypes.Value{
									tftypes.NewValue(tftypes.String, "test"),
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
			attribute: testschema.AttributeWithSetValidators{
				ElementType: types.StringType,
				Validators: []validator.Set{
					testvalidator.Set{
						ValidateSetMethod: func(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
							got := req.ConfigValue
							expected := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")})

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
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithSetValidators{
				ElementType: types.StringType,
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
				AttributePath:   path.Root("test"),
				AttributeConfig: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
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

			AttributeValidateSet(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributeValidateString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute fwxschema.AttributeWithStringValidators
		request   ValidateAttributeRequest
		response  *ValidateAttributeResponse
		expected  *ValidateAttributeResponse
	}{
		"request-path": {
			attribute: testschema.AttributeWithStringValidators{
				Validators: []validator.String{
					testvalidator.String{
						ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("test"),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-pathexpression": {
			attribute: testschema.AttributeWithStringValidators{
				Validators: []validator.String{
					testvalidator.String{
						ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
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
			request: ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig:         types.StringValue("test"),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-config": {
			attribute: testschema.AttributeWithStringValidators{
				Validators: []validator.String{
					testvalidator.String{
						ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
							got := req.Config
							expected := tfsdk.Config{
								Raw: tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "test"),
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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("test"),
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(tftypes.String, "test"),
						},
					),
				},
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"request-configvalue": {
			attribute: testschema.AttributeWithStringValidators{
				Validators: []validator.String{
					testvalidator.String{
						ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
							got := req.ConfigValue
							expected := types.StringValue("test")

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
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("test"),
			},
			response: &ValidateAttributeResponse{},
			expected: &ValidateAttributeResponse{},
		},
		"response-diagnostics": {
			attribute: testschema.AttributeWithStringValidators{
				Validators: []validator.String{
					testvalidator.String{
						ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
							resp.Diagnostics.AddAttributeWarning(req.Path, "New Warning Summary", "New Warning Details")
							resp.Diagnostics.AddAttributeError(req.Path, "New Error Summary", "New Error Details")
						},
					},
				},
			},
			request: ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: types.StringValue("test"),
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

			AttributeValidateString(context.Background(), testCase.attribute, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNestedAttributeObjectValidateObject(t *testing.T) {
	t.Parallel()

	testAttributeConfig := types.ObjectValueMust(
		map[string]attr.Type{"testattr": types.StringType},
		map[string]attr.Value{"testattr": types.StringValue("testvalue")},
	)
	testConfig := tfsdk.Config{
		Raw: tftypes.NewValue(
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
		),
		Schema: testschema.Schema{
			Attributes: map[string]fwschema.Attribute{
				"test": testschema.AttributeWithObjectValidators{
					AttributeTypes: map[string]attr.Type{
						"testattr": types.StringType,
					},
					Required: true,
				},
			},
		},
	}

	testCases := map[string]struct {
		object   fwschema.NestedAttributeObject
		request  ValidateAttributeRequest
		response *ValidateAttributeResponse
		expected *ValidateAttributeResponse
	}{
		"request-path": {
			object: testschema.NestedAttributeObjectWithValidators{
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
			object: testschema.NestedAttributeObjectWithValidators{
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
			object: testschema.NestedAttributeObjectWithValidators{
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
			object: testschema.NestedAttributeObjectWithValidators{
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
			object: testschema.NestedAttributeObjectWithValidators{
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
			object: testschema.NestedAttributeObjectWithValidators{
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
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			NestedAttributeObjectValidate(context.Background(), testCase.object, testCase.request, testCase.response)

			if diff := cmp.Diff(testCase.response, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

var (
	testErrorDiagnostic1 = diag.NewErrorDiagnostic(
		"Error Diagnostic 1",
		"This is an error.",
	)
	testErrorDiagnostic2 = diag.NewErrorDiagnostic(
		"Error Diagnostic 2",
		"This is an error.",
	)
	testWarningDiagnostic1 = diag.NewWarningDiagnostic(
		"Warning Diagnostic 1",
		"This is a warning.",
	)
	testWarningDiagnostic2 = diag.NewWarningDiagnostic(
		"Warning Diagnostic 2",
		"This is a warning.",
	)
)
