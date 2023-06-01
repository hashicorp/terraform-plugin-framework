// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSchemaValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req  ValidateSchemaRequest
		resp ValidateSchemaResponse
	}{
		"no-validation": {
			req: ValidateSchemaRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"attr1": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
							"attr2": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateSchemaResponse{},
		},
		"deprecation-message": {
			req: ValidateSchemaRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"attr1": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
							"attr2": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
						DeprecationMessage: "Use something else instead.",
					},
				},
			},
			resp: ValidateSchemaResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic(
						"Deprecated",
						"Use something else instead.",
					),
				},
			},
		},
		"warnings": {
			req: ValidateSchemaRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"attr1": testschema.AttributeWithStringValidators{
								Required: true,
								Validators: []validator.String{
									testvalidator.String{
										ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
											resp.Diagnostics.Append(testWarningDiagnostic1)
											resp.Diagnostics.Append(testWarningDiagnostic2)
										},
									},
								},
							},
							"attr2": testschema.Attribute{
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateSchemaResponse{
				Diagnostics: diag.Diagnostics{
					testWarningDiagnostic1,
					testWarningDiagnostic2,
				},
			},
		},
		"errors": {
			req: ValidateSchemaRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"attr1": testschema.AttributeWithStringValidators{
								Required: true,
								Validators: []validator.String{
									testvalidator.String{
										ValidateStringMethod: func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
											resp.Diagnostics.Append(testErrorDiagnostic1)
											resp.Diagnostics.Append(testErrorDiagnostic2)
										},
									},
								},
							},
							"attr2": testschema.Attribute{
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateSchemaResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
					testErrorDiagnostic2,
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got ValidateSchemaResponse
			SchemaValidate(context.Background(), tc.req.Config.Schema, tc.req, &got)

			if diff := cmp.Diff(got, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}
