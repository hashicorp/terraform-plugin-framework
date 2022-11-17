package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AttributeValidate performs all Attribute validation.
//
// TODO: Clean up this abstraction back into an internal Attribute type method.
// The extra Attribute parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/365
func AttributeValidate(ctx context.Context, a fwschema.Attribute, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	ctx = logging.FrameworkWithAttributePath(ctx, req.AttributePath.String())

	tfsdkAttribute, ok := a.(tfsdk.Attribute)

	if ok && tfsdkAttribute.GetType() == nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute must define either Attributes or Type. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if ok && len(tfsdkAttribute.GetAttributes()) > 0 && tfsdkAttribute.GetNestingMode() == fwschema.NestingModeUnknown {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute cannot define both Attributes and Type. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if !a.IsRequired() && !a.IsOptional() && !a.IsComputed() {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute missing Required, Optional, or Computed definition. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

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

	// Terraform CLI does not automatically perform certain configuration
	// checks yet. If it eventually does, this logic should remain at least
	// until Terraform CLI versions 0.12 through the release containing the
	// checks are considered end-of-life.
	// Reference: https://github.com/hashicorp/terraform/issues/30669
	if a.IsComputed() && !a.IsOptional() && !attributeConfig.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Configuration for Read-Only Attribute",
			"Cannot set value for this attribute as the provider has marked it as read-only. Remove the configuration line setting the value.\n\n"+
				"Refer to the provider documentation or contact the provider developers for additional information about configurable and read-only attributes that are supported.",
		)
	}

	// Terraform CLI does not automatically perform certain configuration
	// checks yet. If it eventually does, this logic should remain at least
	// until Terraform CLI versions 0.12 through the release containing the
	// checks are considered end-of-life.
	// Reference: https://github.com/hashicorp/terraform/issues/30669
	if a.IsRequired() && attributeConfig.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Missing Configuration for Required Attribute",
			fmt.Sprintf("Must set a configuration value for the %s attribute as the provider has marked it as required.\n\n", req.AttributePath.String())+
				"Refer to the provider documentation or contact the provider developers for additional information about configurable attributes that are required.",
		)
	}

	req.AttributeConfig = attributeConfig

	switch attributeWithValidators := a.(type) {
	// Legacy tfsdk.AttributeValidator handling
	case fwxschema.AttributeWithValidators:
		for _, validator := range attributeWithValidators.GetValidators() {
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
	case fwxschema.AttributeWithBoolValidators:
		AttributeValidateBool(ctx, attributeWithValidators, req, resp)
	case fwxschema.AttributeWithFloat64Validators:
		AttributeValidateFloat64(ctx, attributeWithValidators, req, resp)
	case fwxschema.AttributeWithInt64Validators:
		AttributeValidateInt64(ctx, attributeWithValidators, req, resp)
	case fwxschema.AttributeWithListValidators:
		AttributeValidateList(ctx, attributeWithValidators, req, resp)
	case fwxschema.AttributeWithMapValidators:
		AttributeValidateMap(ctx, attributeWithValidators, req, resp)
	case fwxschema.AttributeWithNumberValidators:
		AttributeValidateNumber(ctx, attributeWithValidators, req, resp)
	case fwxschema.AttributeWithObjectValidators:
		AttributeValidateObject(ctx, attributeWithValidators, req, resp)
	case fwxschema.AttributeWithSetValidators:
		AttributeValidateSet(ctx, attributeWithValidators, req, resp)
	case fwxschema.AttributeWithStringValidators:
		AttributeValidateString(ctx, attributeWithValidators, req, resp)
	}

	AttributeValidateNestedAttributes(ctx, a, req, resp)

	// Show deprecation warnings only for known values.
	if a.GetDeprecationMessage() != "" && !attributeConfig.IsNull() && !attributeConfig.IsUnknown() {
		resp.Diagnostics.AddAttributeWarning(
			req.AttributePath,
			"Attribute Deprecated",
			a.GetDeprecationMessage(),
		)
	}
}

// AttributeValidateBool performs all types.Bool validation.
func AttributeValidateBool(ctx context.Context, attribute fwxschema.AttributeWithBoolValidators, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// Use types.BoolValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.BoolValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Bool Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform Bool attribute validation. "+
				"The value type must implement the types.BoolValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToBoolValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	validateReq := validator.BoolRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
	}

	for _, attributeValidator := range attribute.BoolValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.BoolResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.Bool",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		attributeValidator.ValidateBool(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.Bool",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// AttributeValidateFloat64 performs all types.Float64 validation.
func AttributeValidateFloat64(ctx context.Context, attribute fwxschema.AttributeWithFloat64Validators, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// Use types.Float64Valuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.Float64Valuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Float64 Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform Float64 attribute validation. "+
				"The value type must implement the types.Float64Valuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToFloat64Value(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	validateReq := validator.Float64Request{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
	}

	for _, attributeValidator := range attribute.Float64Validators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.Float64Response{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.Float64",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		attributeValidator.ValidateFloat64(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.Float64",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// AttributeValidateInt64 performs all types.Int64 validation.
func AttributeValidateInt64(ctx context.Context, attribute fwxschema.AttributeWithInt64Validators, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// Use types.Int64Valuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.Int64Valuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Int64 Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform Int64 attribute validation. "+
				"The value type must implement the types.Int64Valuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToInt64Value(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	validateReq := validator.Int64Request{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
	}

	for _, attributeValidator := range attribute.Int64Validators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.Int64Response{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.Int64",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		attributeValidator.ValidateInt64(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.Int64",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// AttributeValidateList performs all types.List validation.
func AttributeValidateList(ctx context.Context, attribute fwxschema.AttributeWithListValidators, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// Use types.ListValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.ListValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid List Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform List attribute validation. "+
				"The value type must implement the types.ListValuable interface. "+
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

	for _, attributeValidator := range attribute.ListValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.ListResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.List",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		attributeValidator.ValidateList(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.List",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// AttributeValidateMap performs all types.Map validation.
func AttributeValidateMap(ctx context.Context, attribute fwxschema.AttributeWithMapValidators, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// Use types.MapValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.MapValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Map Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform Map attribute validation. "+
				"The value type must implement the types.MapValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToMapValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	validateReq := validator.MapRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
	}

	for _, attributeValidator := range attribute.MapValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.MapResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.Map",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		attributeValidator.ValidateMap(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.Map",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// AttributeValidateNumber performs all types.Number validation.
func AttributeValidateNumber(ctx context.Context, attribute fwxschema.AttributeWithNumberValidators, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// Use types.NumberValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.NumberValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Number Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform Number attribute validation. "+
				"The value type must implement the types.NumberValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToNumberValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	validateReq := validator.NumberRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
	}

	for _, attributeValidator := range attribute.NumberValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.NumberResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.Number",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		attributeValidator.ValidateNumber(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.Number",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// AttributeValidateObject performs all types.Object validation.
func AttributeValidateObject(ctx context.Context, attribute fwxschema.AttributeWithObjectValidators, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// Use types.ObjectValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.ObjectValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Object Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform Object attribute validation. "+
				"The value type must implement the types.ObjectValuable interface. "+
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

	for _, attributeValidator := range attribute.ObjectValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.ObjectResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.Object",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		attributeValidator.ValidateObject(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.Object",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// AttributeValidateSet performs all types.Set validation.
func AttributeValidateSet(ctx context.Context, attribute fwxschema.AttributeWithSetValidators, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// Use types.SetValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.SetValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Set Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform Set attribute validation. "+
				"The value type must implement the types.SetValuable interface. "+
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

	for _, attributeValidator := range attribute.SetValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.SetResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.Set",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		attributeValidator.ValidateSet(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.Set",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// AttributeValidateString performs all types.String validation.
func AttributeValidateString(ctx context.Context, attribute fwxschema.AttributeWithStringValidators, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// Use types.StringValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.StringValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid String Attribute Validator Value Type",
			"An unexpected value type was encountered while attempting to perform String attribute validation. "+
				"The value type must implement the types.StringValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToStringValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	validateReq := validator.StringRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
	}

	for _, attributeValidator := range attribute.StringValidators() {
		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		validateResp := &validator.StringResponse{}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined validator.String",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		attributeValidator.ValidateString(ctx, validateReq, validateResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined validator.String",
			map[string]interface{}{
				logging.KeyDescription: attributeValidator.Description(ctx),
			},
		)

		resp.Diagnostics.Append(validateResp.Diagnostics...)
	}
}

// AttributeValidateNestedAttributes performs all nested Attributes validation.
//
// TODO: Clean up this abstraction back into an internal Attribute type method.
// The extra Attribute parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/365
func AttributeValidateNestedAttributes(ctx context.Context, a fwschema.Attribute, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	nestedAttribute, ok := a.(fwschema.NestedAttribute)

	if !ok {
		return
	}

	tfsdkAttribute, ok := a.(tfsdk.Attribute) //nolint:staticcheck // Handle tfsdk.Attribute until its removed.

	if ok && tfsdkAttribute.GetNestingMode() == fwschema.NestingModeUnknown {
		return
	}

	nm := nestedAttribute.GetNestingMode()
	switch nm {
	case fwschema.NestingModeList:
		listVal, ok := req.AttributeConfig.(types.ListValuable)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error Invalid Value Type",
				"A type that implements types.ListValuable is expected here. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		l, diags := listVal.ToListValue(ctx)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for idx := range l.Elements() {
			for nestedName, nestedAttr := range nestedAttribute.GetAttributes() {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath:           req.AttributePath.AtListIndex(idx).AtName(nestedName),
					AttributePathExpression: req.AttributePathExpression.AtListIndex(idx).AtName(nestedName),
					Config:                  req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				AttributeValidate(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case fwschema.NestingModeSet:
		setVal, ok := req.AttributeConfig.(types.SetValuable)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error Invalid Value Type",
				"A type that implements types.SetValuable is expected here. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		s, diags := setVal.ToSetValue(ctx)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, value := range s.Elements() {
			for nestedName, nestedAttr := range nestedAttribute.GetAttributes() {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath:           req.AttributePath.AtSetValue(value).AtName(nestedName),
					AttributePathExpression: req.AttributePathExpression.AtSetValue(value).AtName(nestedName),
					Config:                  req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				AttributeValidate(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case fwschema.NestingModeMap:
		mapVal, ok := req.AttributeConfig.(types.MapValuable)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error Invalid Value Type",
				"A type that implements types.MapValuable is expected here. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		m, diags := mapVal.ToMapValue(ctx)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for key := range m.Elements() {
			for nestedName, nestedAttr := range nestedAttribute.GetAttributes() {
				nestedAttrReq := tfsdk.ValidateAttributeRequest{
					AttributePath:           req.AttributePath.AtMapKey(key).AtName(nestedName),
					AttributePathExpression: req.AttributePathExpression.AtMapKey(key).AtName(nestedName),
					Config:                  req.Config,
				}
				nestedAttrResp := &tfsdk.ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				AttributeValidate(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case fwschema.NestingModeSingle:
		objectVal, ok := req.AttributeConfig.(types.ObjectValuable)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error Invalid Value Type",
				"A type that implements types.ObjectValuable is expected here. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		o, diags := objectVal.ToObjectValue(ctx)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if o.IsNull() || o.IsUnknown() {
			return
		}

		for nestedName, nestedAttr := range nestedAttribute.GetAttributes() {
			nestedAttrReq := tfsdk.ValidateAttributeRequest{
				AttributePath:           req.AttributePath.AtName(nestedName),
				AttributePathExpression: req.AttributePathExpression.AtName(nestedName),
				Config:                  req.Config,
			}
			nestedAttrResp := &tfsdk.ValidateAttributeResponse{
				Diagnostics: resp.Diagnostics,
			}

			AttributeValidate(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

			resp.Diagnostics = nestedAttrResp.Diagnostics
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
