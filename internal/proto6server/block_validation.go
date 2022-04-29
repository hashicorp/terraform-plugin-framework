package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BlockValidate performs all Block validation.
//
// TODO: Clean up this abstraction back into an internal Block type method.
// The extra Block parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func BlockValidate(ctx context.Context, b tfsdk.Block, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	attributeConfig, diags := ConfigGetAttributeValue(ctx, req.Config, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeConfig = attributeConfig

	for _, validator := range b.Validators {
		validator.Validate(ctx, req, resp)
	}

	nm := b.NestingMode
	switch nm {
	case tfsdk.BlockNestingModeList:
		l, ok := req.AttributeConfig.(types.List)

		if !ok {
			err := fmt.Errorf("unknown block value type (%s) for nesting mode (%T) at path: %s", req.AttributeConfig.Type(ctx), nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Error",
				"Block validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for idx := range l.Elems {
			for name, attr := range b.Attributes {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(name),
					Config:        req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				AttributeValidate(ctx, attr, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}

			for name, block := range b.Blocks {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(name),
					Config:        req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				BlockValidate(ctx, block, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case tfsdk.BlockNestingModeSet:
		s, ok := req.AttributeConfig.(types.Set)

		if !ok {
			err := fmt.Errorf("unknown block value type (%s) for nesting mode (%T) at path: %s", req.AttributeConfig.Type(ctx), nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Error",
				"Block validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for _, value := range s.Elems {
			tfValue, err := value.ToTerraformValue(ctx)
			if err != nil {
				err := fmt.Errorf("error running ToTerraformValue on element value: %v", value)
				resp.Diagnostics.AddAttributeError(
					req.AttributePath,
					"Block Validation Error",
					"Block validation cannot convert element into a Terraform value. Report this to the provider developer:\n\n"+err.Error(),
				)

				return
			}

			for name, attr := range b.Attributes {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(name),
					Config:        req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				AttributeValidate(ctx, attr, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}

			for name, block := range b.Blocks {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(name),
					Config:        req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				BlockValidate(ctx, block, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	default:
		err := fmt.Errorf("unknown block validation nesting mode (%T: %v) at path: %s", nm, nm, req.AttributePath)
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Block Validation Error",
			"Block validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}

	if b.DeprecationMessage != "" && attributeConfig != nil {
		tfValue, err := attributeConfig.ToTerraformValue(ctx)

		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Error",
				"Block validation cannot convert value. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		if !tfValue.IsNull() {
			resp.Diagnostics.AddAttributeWarning(
				req.AttributePath,
				"Block Deprecated",
				b.DeprecationMessage,
			)
		}
	}
}
