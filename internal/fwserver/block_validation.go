package fwserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// BlockValidate performs all Block validation.
//
// TODO: Clean up this abstraction back into an internal Block type method.
// The extra Block parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/365
func BlockValidate(ctx context.Context, b fwschema.Block, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	configData := &fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionConfiguration,
		Schema:         req.Config.Schema,
		TerraformValue: req.Config.Raw,
	}

	attributeConfig, diags := configData.ValueAtPath(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeConfig = attributeConfig

	switch blockWithValidators := b.(type) {
	case fwxschema.BlockWithListValidators:
		BlockValidateList(ctx, blockWithValidators, req, resp)
	case fwxschema.BlockWithObjectValidators:
		BlockValidateObject(ctx, blockWithValidators, req, resp)
	case fwxschema.BlockWithSetValidators:
		BlockValidateSet(ctx, blockWithValidators, req, resp)
	}

	nestedBlockObject := b.GetNestedObject()

	nm := b.GetNestingMode()
	switch nm {
	case fwschema.BlockNestingModeList:
		listVal, ok := req.AttributeConfig.(basetypes.ListValuable)

		if !ok {
			err := fmt.Errorf("unknown block value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Error Invalid Value Type",
				"A type that implements basetypes.ListValuable is expected here. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		l, diags := listVal.ToListValue(ctx)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for idx, value := range l.Elements() {
			nestedBlockObjectReq := ValidateAttributeRequest{
				AttributeConfig:         value,
				AttributePath:           req.AttributePath.AtListIndex(idx),
				AttributePathExpression: req.AttributePathExpression.AtListIndex(idx),
				Config:                  req.Config,
			}
			nestedBlockObjectResp := &ValidateAttributeResponse{}

			NestedBlockObjectValidate(ctx, nestedBlockObject, nestedBlockObjectReq, nestedBlockObjectResp)

			resp.Diagnostics.Append(nestedBlockObjectResp.Diagnostics...)
		}

		// Terraform 0.12 through 0.15.1 do not implement block MaxItems
		// validation.
		//
		// Terraform 0.15.2 and later implements MaxItems validation during
		// configuration decoding, so if this framework drops Terraform support
		// for earlier versions, this validation can be removed.
		if b.GetMaxItems() > 0 && int64(len(l.Elements())) > b.GetMaxItems() {
			resp.Diagnostics.Append(blockMaxItemsDiagnostic(req.AttributePath, b.GetMaxItems(), len(l.Elements())))
		}

		// Terraform 0.12 through 0.15.1 implement conservative block MinItems
		// validation, where the MinItems can be reset to 1 in certain
		// situations. This validation must ensure the list itself is not
		// unknown, which could erroneously trigger the error since the list
		// would have 0 elements.
		//
		// Terraform 0.15.2 and later implements proper MinItems validation
		// during configuration decoding, so if this framework drops Terraform
		// support for earlier versions, this validation can be removed.
		if b.GetMinItems() > 0 && int64(len(l.Elements())) < b.GetMinItems() && !l.IsUnknown() {
			resp.Diagnostics.Append(blockMinItemsDiagnostic(req.AttributePath, b.GetMinItems(), len(l.Elements())))
		}
	case fwschema.BlockNestingModeSet:
		setVal, ok := req.AttributeConfig.(basetypes.SetValuable)

		if !ok {
			err := fmt.Errorf("unknown block value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Error Invalid Value Type",
				"A type that implements basetypes.SetValuable is expected here. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		s, diags := setVal.ToSetValue(ctx)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, value := range s.Elements() {
			nestedBlockObjectReq := ValidateAttributeRequest{
				AttributeConfig:         value,
				AttributePath:           req.AttributePath.AtSetValue(value),
				AttributePathExpression: req.AttributePathExpression.AtSetValue(value),
				Config:                  req.Config,
			}
			nestedBlockObjectResp := &ValidateAttributeResponse{}

			NestedBlockObjectValidate(ctx, nestedBlockObject, nestedBlockObjectReq, nestedBlockObjectResp)

			resp.Diagnostics.Append(nestedBlockObjectResp.Diagnostics...)
		}

		// Terraform 0.12 through 0.15.1 do not implement block MaxItems
		// validation.
		//
		// Terraform 0.15.2 and later implements MaxItems validation during
		// configuration decoding, so if this framework drops Terraform support
		// for earlier versions, this validation can be removed.
		if b.GetMaxItems() > 0 && int64(len(s.Elements())) > b.GetMaxItems() {
			resp.Diagnostics.Append(blockMaxItemsDiagnostic(req.AttributePath, b.GetMaxItems(), len(s.Elements())))
		}

		// Terraform 0.12 through 0.15.1 implement conservative block MinItems
		// validation, where the MinItems can be reset to 1 in certain
		// situations. This validation must ensure the set itself is not
		// unknown, which could erroneously trigger the error since the set
		// would have 0 elements.
		//
		// Terraform 0.15.2 and later implements proper MinItems validation
		// during configuration decoding, so if this framework drops Terraform
		// support for earlier versions, this validation can be removed.
		if b.GetMinItems() > 0 && int64(len(s.Elements())) < b.GetMinItems() && !s.IsUnknown() {
			resp.Diagnostics.Append(blockMinItemsDiagnostic(req.AttributePath, b.GetMinItems(), len(s.Elements())))
		}
	case fwschema.BlockNestingModeSingle:
		objectVal, ok := req.AttributeConfig.(basetypes.ObjectValuable)

		if !ok {
			err := fmt.Errorf("unknown block value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Error Invalid Value Type",
				"A type that implements basetypes.ObjectValuable is expected here. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		o, diags := objectVal.ToObjectValue(ctx)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		nestedBlockObjectReq := ValidateAttributeRequest{
			AttributeConfig:         o,
			AttributePath:           req.AttributePath,
			AttributePathExpression: req.AttributePathExpression,
			Config:                  req.Config,
		}
		nestedBlockObjectResp := &ValidateAttributeResponse{}

		NestedBlockObjectValidate(ctx, nestedBlockObject, nestedBlockObjectReq, nestedBlockObjectResp)

		resp.Diagnostics.Append(nestedBlockObjectResp.Diagnostics...)

		if b.GetMinItems() == 1 && o.IsNull() {
			resp.Diagnostics.Append(blockMinItemsDiagnostic(req.AttributePath, b.GetMinItems(), 0))
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

	// Show deprecation warning only on known values.
	if b.GetDeprecationMessage() != "" && !attributeConfig.IsNull() && !attributeConfig.IsUnknown() {
		resp.Diagnostics.AddAttributeWarning(
			req.AttributePath,
			"Block Deprecated",
			b.GetDeprecationMessage(),
		)
	}
}

// BlockValidateList performs all types.List validation.
func BlockValidateList(ctx context.Context, block fwxschema.BlockWithListValidators, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	// Use basetypes.ListValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(basetypes.ListValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid List Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform List attribute validation. "+
				"The value type must implement the basetypes.ListValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToListValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	validateReq := validator.ListRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
	}

	for _, blockValidator := range block.ListValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.ListResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.List",
			map[string]interface{}{
				logging.KeyDescription: blockValidator.Description(ctx),
			},
		)

		blockValidator.ValidateList(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.List",
			map[string]interface{}{
				logging.KeyDescription: blockValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// BlockValidateObject performs all types.Object validation.
func BlockValidateObject(ctx context.Context, block fwxschema.BlockWithObjectValidators, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	// Use basetypes.ObjectValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(basetypes.ObjectValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Object Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform Object attribute validation. "+
				"The value type must implement the basetypes.ObjectValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToObjectValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	validateReq := validator.ObjectRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
	}

	for _, blockValidator := range block.ObjectValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.ObjectResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.Object",
			map[string]interface{}{
				logging.KeyDescription: blockValidator.Description(ctx),
			},
		)

		blockValidator.ValidateObject(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.Object",
			map[string]interface{}{
				logging.KeyDescription: blockValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// BlockValidateSet performs all types.Set validation.
func BlockValidateSet(ctx context.Context, block fwxschema.BlockWithSetValidators, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	// Use basetypes.SetValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(basetypes.SetValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Set Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform Set attribute validation. "+
				"The value type must implement the basetypes.SetValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToSetValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	validateReq := validator.SetRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
	}

	for _, blockValidator := range block.SetValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.SetResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.Set",
			map[string]interface{}{
				logging.KeyDescription: blockValidator.Description(ctx),
			},
		)

		blockValidator.ValidateSet(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.Set",
			map[string]interface{}{
				logging.KeyDescription: blockValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

func NestedBlockObjectValidate(ctx context.Context, o fwschema.NestedBlockObject, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	objectWithValidators, ok := o.(fwxschema.NestedBlockObjectWithValidators)

	if ok {
		objectVal, ok := req.AttributeConfig.(basetypes.ObjectValuable)

		if !ok {
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Walk Error",
				"An unexpected error occurred while walking the schema for block validation. "+
					"This is an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Unknown block value type (%T) at path: %s", req.AttributeConfig, req.AttributePath),
			)

			return
		}

		object, diags := objectVal.ToObjectValue(ctx)

		resp.Diagnostics.Append(diags...)

		// Only return early on new errors as the resp.Diagnostics may have
		// errors from other attributes.
		if diags.HasError() {
			return
		}

		validateReq := validator.ObjectRequest{
			Config:         req.Config,
			ConfigValue:    object,
			Path:           req.AttributePath,
			PathExpression: req.AttributePathExpression,
		}

		for _, objectValidator := range objectWithValidators.ObjectValidators() {
			// Instantiate a new response for each request to prevent validators
			// from modifying or removing diagnostics.
			validateResp := &validator.ObjectResponse{}

			logging.FrameworkDebug(
				ctx,
				"Calling provider defined validator.Object",
				map[string]interface{}{
					logging.KeyDescription: objectValidator.Description(ctx),
				},
			)

			objectValidator.ValidateObject(ctx, validateReq, validateResp)

			logging.FrameworkDebug(
				ctx,
				"Called provider defined validator.Object",
				map[string]interface{}{
					logging.KeyDescription: objectValidator.Description(ctx),
				},
			)

			resp.Diagnostics.Append(validateResp.Diagnostics...)
		}
	}

	for nestedName, nestedAttr := range o.GetAttributes() {
		nestedAttrReq := ValidateAttributeRequest{
			AttributePath:           req.AttributePath.AtName(nestedName),
			AttributePathExpression: req.AttributePathExpression.AtName(nestedName),
			Config:                  req.Config,
		}
		nestedAttrResp := &ValidateAttributeResponse{}

		AttributeValidate(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

		resp.Diagnostics.Append(nestedAttrResp.Diagnostics...)
	}

	for nestedName, nestedBlock := range o.GetBlocks() {
		nestedBlockReq := ValidateAttributeRequest{
			AttributePath:           req.AttributePath.AtName(nestedName),
			AttributePathExpression: req.AttributePathExpression.AtName(nestedName),
			Config:                  req.Config,
		}
		nestedBlockResp := &ValidateAttributeResponse{}

		BlockValidate(ctx, nestedBlock, nestedBlockReq, nestedBlockResp)

		resp.Diagnostics.Append(nestedBlockResp.Diagnostics...)
	}
}

func blockMaxItemsDiagnostic(attrPath path.Path, maxItems int64, elements int) diag.Diagnostic {
	var details strings.Builder

	details.WriteString("The configuration should declare a maximum of ")

	if maxItems == 1 {
		details.WriteString("1 block")
	} else {
		details.WriteString(fmt.Sprintf("%d blocks", maxItems))
	}

	// Elements will always be greater than 1, so do not need to handle the
	// singular case.
	details.WriteString(fmt.Sprintf(", however %d blocks were configured.", elements))

	return diag.NewAttributeErrorDiagnostic(
		attrPath,
		"Extra Block Configuration",
		details.String(),
	)
}

func blockMinItemsDiagnostic(attrPath path.Path, minItems int64, elements int) diag.Diagnostic {
	var details strings.Builder

	details.WriteString("The configuration should declare a minimum of ")

	if minItems == 1 {
		details.WriteString("1 block")
	} else {
		details.WriteString(fmt.Sprintf("%d blocks", minItems))
	}

	details.WriteString(", however ")

	if elements == 1 {
		details.WriteString("1 block was")
	} else {
		details.WriteString(fmt.Sprintf("%d blocks were", elements))
	}

	details.WriteString(" configured.")

	return diag.NewAttributeErrorDiagnostic(
		attrPath,
		"Missing Block Configuration",
		details.String(),
	)
}
