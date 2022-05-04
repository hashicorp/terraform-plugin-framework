package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AttributeValidate performs all Attribute validation.
//
// TODO: Clean up this abstraction back into an internal Attribute type method.
// The extra Attribute parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func AttributeValidate(ctx context.Context, a tfsdk.Attribute, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	ctx = logging.FrameworkWithAttributePath(ctx, req.AttributePath.String())

	if (a.Attributes == nil || len(a.Attributes.GetAttributes()) == 0) && a.Type == nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute must define either Attributes or Type. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if a.Attributes != nil && len(a.Attributes.GetAttributes()) > 0 && a.Type != nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute cannot define both Attributes and Type. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if !a.Required && !a.Optional && !a.Computed {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute missing Required, Optional, or Computed definition. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	attributeConfig, diags := ConfigGetAttributeValue(ctx, req.Config, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeConfig = attributeConfig

	for _, validator := range a.Validators {
		logging.FrameworkDebug(
			ctx,
			"Calling provider defined AttributeValidator",
			map[string]interface{}{
				logging.KeyDescription: validator.Description(ctx),
			},
		)
		validator.Validate(ctx, req, resp)
		logging.FrameworkDebug(
			ctx,
			"Called provider defined AttributeValidator",
			map[string]interface{}{
				logging.KeyDescription: validator.Description(ctx),
			},
		)
	}

	AttributeValidateNestedAttributes(ctx, a, req, resp)

	if a.DeprecationMessage != "" && attributeConfig != nil {
		tfValue, err := attributeConfig.ToTerraformValue(ctx)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot convert value. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		if !tfValue.IsNull() {
			resp.Diagnostics.AddAttributeWarning(
				req.AttributePath,
				"Attribute Deprecated",
				a.DeprecationMessage,
			)
		}
	}
}

// AttributeValidateNestedAttributes performs all nested Attributes validation.
//
// TODO: Clean up this abstraction back into an internal Attribute type method.
// The extra Attribute parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func AttributeValidateNestedAttributes(ctx context.Context, a tfsdk.Attribute, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	if a.Attributes == nil || len(a.Attributes.GetAttributes()) == 0 {
		return
	}

	nm := a.Attributes.GetNestingMode()
	switch nm {
	case tfsdk.NestingModeList:
		l, ok := req.AttributeConfig.(types.List)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for idx := range l.Elems {
			for nestedName, nestedAttr := range a.Attributes.GetAttributes() {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				AttributeValidate(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case tfsdk.NestingModeSet:
		s, ok := req.AttributeConfig.(types.Set)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for _, value := range s.Elems {
			tfValue, err := value.ToTerraformValue(ctx)
			if err != nil {
				err := fmt.Errorf("error running ToTerraformValue on element value: %v", value)
				resp.Diagnostics.AddAttributeError(
					req.AttributePath,
					"Attribute Validation Error",
					"Attribute validation cannot convert element into a Terraform value. Report this to the provider developer:\n\n"+err.Error(),
				)

				return
			}

			for nestedName, nestedAttr := range a.Attributes.GetAttributes() {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				AttributeValidate(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case tfsdk.NestingModeMap:
		m, ok := req.AttributeConfig.(types.Map)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for key := range m.Elems {
			for nestedName, nestedAttr := range a.Attributes.GetAttributes() {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyString(key).WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				AttributeValidate(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case tfsdk.NestingModeSingle:
		o, ok := req.AttributeConfig.(types.Object)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		if !o.Null && !o.Unknown {
			for nestedName, nestedAttr := range a.Attributes.GetAttributes() {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				AttributeValidate(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	default:
		err := fmt.Errorf("unknown attribute validation nesting mode (%T: %v) at path: %s", nm, nm, req.AttributePath)
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Attribute Validation Error",
			"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}
}
