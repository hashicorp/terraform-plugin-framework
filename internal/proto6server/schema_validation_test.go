package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
							},
							"attr2": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
							},
							"attr2": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
								Validators: []tfsdk.AttributeValidator{
									testWarningAttributeValidator{},
								},
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
								Validators: []tfsdk.AttributeValidator{
									testWarningAttributeValidator{},
								},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
								Validators: []tfsdk.AttributeValidator{
									testErrorAttributeValidator{},
								},
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
								Validators: []tfsdk.AttributeValidator{
									testErrorAttributeValidator{},
								},
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
