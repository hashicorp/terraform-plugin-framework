package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// This file contains temporary types until GetSchema and Metadata are required
// in Resource.

var _ resource.Resource = &ResourceWithConfigValidatorsAndGetSchemaAndMetadata{}
var _ resource.ResourceWithConfigValidators = &ResourceWithConfigValidatorsAndGetSchemaAndMetadata{}
var _ resource.ResourceWithGetSchema = &ResourceWithConfigValidatorsAndGetSchemaAndMetadata{}
var _ resource.ResourceWithMetadata = &ResourceWithConfigValidatorsAndGetSchemaAndMetadata{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and Metadata are required in Resource.
type ResourceWithConfigValidatorsAndGetSchemaAndMetadata struct {
	*Resource

	// ResourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []resource.ConfigValidator

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithMetadata interface methods
	MetadataMethod func(context.Context, resource.MetadataRequest, *resource.MetadataResponse)
}

// ConfigValidators satisfies the resource.ResourceWithConfigValidators interface.
func (r *ResourceWithConfigValidatorsAndGetSchemaAndMetadata) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if r.ConfigValidatorsMethod == nil {
		return nil
	}

	return r.ConfigValidatorsMethod(ctx)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithConfigValidatorsAndGetSchemaAndMetadata) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// Metadata satisfies the resource.ResourceWithMetadata interface.
func (r *ResourceWithConfigValidatorsAndGetSchemaAndMetadata) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}

var _ resource.Resource = &ResourceWithGetSchemaAndMetadata{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndMetadata{}
var _ resource.ResourceWithMetadata = &ResourceWithGetSchemaAndMetadata{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and Metadata are required in Resource.
type ResourceWithGetSchemaAndMetadata struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithMetadata interface methods
	MetadataMethod func(context.Context, resource.MetadataRequest, *resource.MetadataResponse)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndMetadata) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// Metadata satisfies the resource.ResourceWithMetadata interface.
func (r *ResourceWithGetSchemaAndMetadata) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}

var _ resource.Resource = &ResourceWithGetSchemaAndImportStateAndMetadata{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndImportStateAndMetadata{}
var _ resource.ResourceWithImportState = &ResourceWithGetSchemaAndImportStateAndMetadata{}
var _ resource.ResourceWithMetadata = &ResourceWithGetSchemaAndImportStateAndMetadata{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and Metadata are required in Resource.
type ResourceWithGetSchemaAndImportStateAndMetadata struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithImportState interface methods
	ImportStateMethod func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse)

	// ResourceWithMetadata interface methods
	MetadataMethod func(context.Context, resource.MetadataRequest, *resource.MetadataResponse)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndImportStateAndMetadata) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// ImportState satisfies the resource.ResourceWithImportState interface.
func (r *ResourceWithGetSchemaAndImportStateAndMetadata) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.ImportStateMethod == nil {
		return
	}

	r.ImportStateMethod(ctx, req, resp)
}

// Metadata satisfies the resource.ResourceWithMetadata interface.
func (r *ResourceWithGetSchemaAndImportStateAndMetadata) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}

var _ resource.Resource = &ResourceWithGetSchemaAndModifyPlanAndMetadata{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndModifyPlanAndMetadata{}
var _ resource.ResourceWithModifyPlan = &ResourceWithGetSchemaAndModifyPlanAndMetadata{}
var _ resource.ResourceWithMetadata = &ResourceWithGetSchemaAndModifyPlanAndMetadata{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and Metadata are required in Resource.
type ResourceWithGetSchemaAndModifyPlanAndMetadata struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithModifyPlan interface methods
	ModifyPlanMethod func(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)

	// ResourceWithMetadata interface methods
	MetadataMethod func(context.Context, resource.MetadataRequest, *resource.MetadataResponse)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndModifyPlanAndMetadata) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// ModifyPlan satisfies the resource.ResourceWithModifyPlan interface.
func (r *ResourceWithGetSchemaAndModifyPlanAndMetadata) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if r.ModifyPlanMethod == nil {
		return
	}

	r.ModifyPlanMethod(ctx, req, resp)
}

// Metadata satisfies the resource.ResourceWithMetadata interface.
func (r *ResourceWithGetSchemaAndModifyPlanAndMetadata) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}

var _ resource.Resource = &ResourceWithGetSchemaAndMetadataAndUpgradeState{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndMetadataAndUpgradeState{}
var _ resource.ResourceWithMetadata = &ResourceWithGetSchemaAndMetadataAndUpgradeState{}
var _ resource.ResourceWithUpgradeState = &ResourceWithGetSchemaAndMetadataAndUpgradeState{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and Metadata are required in Resource.
type ResourceWithGetSchemaAndMetadataAndUpgradeState struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithMetadata interface methods
	MetadataMethod func(context.Context, resource.MetadataRequest, *resource.MetadataResponse)

	// ResourceWithUpgradeState interface methods
	UpgradeStateMethod func(context.Context) map[int64]resource.StateUpgrader
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndMetadataAndUpgradeState) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// Metadata satisfies the resource.ResourceWithMetadata interface.
func (r *ResourceWithGetSchemaAndMetadataAndUpgradeState) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}

// UpgradeState satisfies the resource.ResourceWithUpgradeState interface.
func (r *ResourceWithGetSchemaAndMetadataAndUpgradeState) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if r.UpgradeStateMethod == nil {
		return nil
	}

	return r.UpgradeStateMethod(ctx)
}

var _ resource.Resource = &ResourceWithGetSchemaAndMetadataAndValidateConfig{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndMetadataAndValidateConfig{}
var _ resource.ResourceWithMetadata = &ResourceWithGetSchemaAndMetadataAndValidateConfig{}
var _ resource.ResourceWithValidateConfig = &ResourceWithGetSchemaAndMetadataAndValidateConfig{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and Metadata are required in Resource.
type ResourceWithGetSchemaAndMetadataAndValidateConfig struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithMetadata interface methods
	MetadataMethod func(context.Context, resource.MetadataRequest, *resource.MetadataResponse)

	// ResourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndMetadataAndValidateConfig) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// Metadata satisfies the resource.ResourceWithMetadata interface.
func (r *ResourceWithGetSchemaAndMetadataAndValidateConfig) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}

// ValidateConfig satisfies the resource.ResourceWithValidateConfig interface.
func (r *ResourceWithGetSchemaAndMetadataAndValidateConfig) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.ValidateConfigMethod == nil {
		return
	}

	r.ValidateConfigMethod(ctx, req, resp)
}
