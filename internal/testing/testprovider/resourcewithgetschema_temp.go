package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// This file contains temporary types until GetSchema and TypeName are required
// in Resource.

var _ resource.Resource = &ResourceWithConfigValidatorsAndGetSchemaAndTypeName{}
var _ resource.ResourceWithConfigValidators = &ResourceWithConfigValidatorsAndGetSchemaAndTypeName{}
var _ resource.ResourceWithGetSchema = &ResourceWithConfigValidatorsAndGetSchemaAndTypeName{}
var _ resource.ResourceWithTypeName = &ResourceWithConfigValidatorsAndGetSchemaAndTypeName{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and TypeName are required in Resource.
type ResourceWithConfigValidatorsAndGetSchemaAndTypeName struct {
	*Resource

	// ResourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []resource.ConfigValidator

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithTypeName interface methods
	TypeNameMethod func(context.Context, resource.TypeNameRequest, *resource.TypeNameResponse)
}

// ConfigValidators satisfies the resource.ResourceWithConfigValidators interface.
func (r *ResourceWithConfigValidatorsAndGetSchemaAndTypeName) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if r.ConfigValidatorsMethod == nil {
		return nil
	}

	return r.ConfigValidatorsMethod(ctx)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithConfigValidatorsAndGetSchemaAndTypeName) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// TypeName satisfies the resource.ResourceWithTypeName interface.
func (r *ResourceWithConfigValidatorsAndGetSchemaAndTypeName) TypeName(ctx context.Context, req resource.TypeNameRequest, resp *resource.TypeNameResponse) {
	if r.TypeNameMethod == nil {
		return
	}

	r.TypeNameMethod(ctx, req, resp)
}

var _ resource.Resource = &ResourceWithGetSchemaAndTypeName{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndTypeName{}
var _ resource.ResourceWithTypeName = &ResourceWithGetSchemaAndTypeName{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and TypeName are required in Resource.
type ResourceWithGetSchemaAndTypeName struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithTypeName interface methods
	TypeNameMethod func(context.Context, resource.TypeNameRequest, *resource.TypeNameResponse)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndTypeName) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// TypeName satisfies the resource.ResourceWithTypeName interface.
func (r *ResourceWithGetSchemaAndTypeName) TypeName(ctx context.Context, req resource.TypeNameRequest, resp *resource.TypeNameResponse) {
	if r.TypeNameMethod == nil {
		return
	}

	r.TypeNameMethod(ctx, req, resp)
}

var _ resource.Resource = &ResourceWithGetSchemaAndImportStateAndTypeName{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndImportStateAndTypeName{}
var _ resource.ResourceWithImportState = &ResourceWithGetSchemaAndImportStateAndTypeName{}
var _ resource.ResourceWithTypeName = &ResourceWithGetSchemaAndImportStateAndTypeName{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and TypeName are required in Resource.
type ResourceWithGetSchemaAndImportStateAndTypeName struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithImportState interface methods
	ImportStateMethod func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse)

	// ResourceWithTypeName interface methods
	TypeNameMethod func(context.Context, resource.TypeNameRequest, *resource.TypeNameResponse)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndImportStateAndTypeName) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// ImportState satisfies the resource.ResourceWithImportState interface.
func (r *ResourceWithGetSchemaAndImportStateAndTypeName) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.ImportStateMethod == nil {
		return
	}

	r.ImportStateMethod(ctx, req, resp)
}

// TypeName satisfies the resource.ResourceWithTypeName interface.
func (r *ResourceWithGetSchemaAndImportStateAndTypeName) TypeName(ctx context.Context, req resource.TypeNameRequest, resp *resource.TypeNameResponse) {
	if r.TypeNameMethod == nil {
		return
	}

	r.TypeNameMethod(ctx, req, resp)
}

var _ resource.Resource = &ResourceWithGetSchemaAndModifyPlanAndTypeName{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndModifyPlanAndTypeName{}
var _ resource.ResourceWithModifyPlan = &ResourceWithGetSchemaAndModifyPlanAndTypeName{}
var _ resource.ResourceWithTypeName = &ResourceWithGetSchemaAndModifyPlanAndTypeName{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and TypeName are required in Resource.
type ResourceWithGetSchemaAndModifyPlanAndTypeName struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithModifyPlan interface methods
	ModifyPlanMethod func(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)

	// ResourceWithTypeName interface methods
	TypeNameMethod func(context.Context, resource.TypeNameRequest, *resource.TypeNameResponse)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndModifyPlanAndTypeName) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// ModifyPlan satisfies the resource.ResourceWithModifyPlan interface.
func (r *ResourceWithGetSchemaAndModifyPlanAndTypeName) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if r.ModifyPlanMethod == nil {
		return
	}

	r.ModifyPlanMethod(ctx, req, resp)
}

// TypeName satisfies the resource.ResourceWithTypeName interface.
func (r *ResourceWithGetSchemaAndModifyPlanAndTypeName) TypeName(ctx context.Context, req resource.TypeNameRequest, resp *resource.TypeNameResponse) {
	if r.TypeNameMethod == nil {
		return
	}

	r.TypeNameMethod(ctx, req, resp)
}

var _ resource.Resource = &ResourceWithGetSchemaAndTypeNameAndUpgradeState{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndTypeNameAndUpgradeState{}
var _ resource.ResourceWithTypeName = &ResourceWithGetSchemaAndTypeNameAndUpgradeState{}
var _ resource.ResourceWithUpgradeState = &ResourceWithGetSchemaAndTypeNameAndUpgradeState{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and TypeName are required in Resource.
type ResourceWithGetSchemaAndTypeNameAndUpgradeState struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithTypeName interface methods
	TypeNameMethod func(context.Context, resource.TypeNameRequest, *resource.TypeNameResponse)

	// ResourceWithUpgradeState interface methods
	UpgradeStateMethod func(context.Context) map[int64]resource.StateUpgrader
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndTypeNameAndUpgradeState) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// TypeName satisfies the resource.ResourceWithTypeName interface.
func (r *ResourceWithGetSchemaAndTypeNameAndUpgradeState) TypeName(ctx context.Context, req resource.TypeNameRequest, resp *resource.TypeNameResponse) {
	if r.TypeNameMethod == nil {
		return
	}

	r.TypeNameMethod(ctx, req, resp)
}

// UpgradeState satisfies the resource.ResourceWithUpgradeState interface.
func (r *ResourceWithGetSchemaAndTypeNameAndUpgradeState) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if r.UpgradeStateMethod == nil {
		return nil
	}

	return r.UpgradeStateMethod(ctx)
}

var _ resource.Resource = &ResourceWithGetSchemaAndTypeNameAndValidateConfig{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchemaAndTypeNameAndValidateConfig{}
var _ resource.ResourceWithTypeName = &ResourceWithGetSchemaAndTypeNameAndValidateConfig{}
var _ resource.ResourceWithValidateConfig = &ResourceWithGetSchemaAndTypeNameAndValidateConfig{}

// Declarative resource.ResourceWithGetSchema for unit testing. This type is
// temporary until GetSchema and TypeName are required in Resource.
type ResourceWithGetSchemaAndTypeNameAndValidateConfig struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ResourceWithTypeName interface methods
	TypeNameMethod func(context.Context, resource.TypeNameRequest, *resource.TypeNameResponse)

	// ResourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchemaAndTypeNameAndValidateConfig) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}

// TypeName satisfies the resource.ResourceWithTypeName interface.
func (r *ResourceWithGetSchemaAndTypeNameAndValidateConfig) TypeName(ctx context.Context, req resource.TypeNameRequest, resp *resource.TypeNameResponse) {
	if r.TypeNameMethod == nil {
		return
	}

	r.TypeNameMethod(ctx, req, resp)
}

// ValidateConfig satisfies the resource.ResourceWithValidateConfig interface.
func (r *ResourceWithGetSchemaAndTypeNameAndValidateConfig) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.ValidateConfigMethod == nil {
		return
	}

	r.ValidateConfigMethod(ctx, req, resp)
}
