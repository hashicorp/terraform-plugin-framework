package tfsdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

// AttributePlanModifier represents a modifier for an attribute at plan time.
// An AttributePlanModifier can only modify the planned value for the attribute
// on which it is defined. For plan-time modifications that modify the values of
// several attributes at once, please instead use the ResourceWithModifyPlan
// interface by defining a ModifyPlan function on the resource.
type AttributePlanModifier interface {
	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this modifier is,
	// what it's for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description(context.Context) string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this modifier is, what it's for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription(context.Context) string

	// Modify is called when the provider has an opportunity to modify
	// the plan: once during the plan phase when Terraform is determining
	// the diff that should be shown to the user for approval, and once
	// during the apply phase with any unknown values from configuration
	// filled in with their final values.
	//
	// The Modify function has access to the config, state, and plan for
	// both the attribute in question and the entire resource, but it can
	// only modify the value of the one attribute.
	//
	// Any returned errors will stop further execution of plan modifications
	// for this Attribute and any nested Attribute. Other Attribute at the same
	// or higher levels of the Schema will still execute any plan modifications
	// to ensure all warnings and errors across all root Attribute are
	// captured.
	//
	// Please see the documentation for ResourceWithModifyPlan#ModifyPlan
	// for further details.
	Modify(context.Context, ModifyAttributePlanRequest, *ModifyAttributePlanResponse)
}

// AttributePlanModifiers represents a sequence of AttributePlanModifiers, in
// order.
type AttributePlanModifiers []AttributePlanModifier

// RequiresReplace returns an AttributePlanModifier specifying the attribute as
// requiring replacement. This behaviour is identical to the ForceNew behaviour
// in terraform-plugin-sdk and will result in the resource being destroyed and
// recreated when the following conditions are met:
//
// 1. The resource's state is not null; a null state indicates that we're
// creating a resource, and we never need to destroy and recreate a resource
// when we're creating it.
//
// 2. The resource's plan is not null; a null plan indicates that we're
// deleting a resource, and we never need to destroy and recreate a resource
// when we're deleting it.
//
// 3. The attribute's config is not null or the attribute is not computed; a
// computed attribute with a null config almost always means that the provider
// is changing the value, and practitioners are usually unpleasantly surprised
// when a resource is destroyed and recreated when their configuration hasn't
// changed. This has the unfortunate side effect that removing a computed field
// from the config will not trigger a destroy and recreate cycle, even when
// that is warranted. To get around this, provider developer can implement
// their own AttributePlanModifier that handles that behavior in the way that
// most makes sense for their use case.
//
// 4. The attribute's value in the plan does not match the attribute's value in
// the state.
func RequiresReplace() AttributePlanModifier {
	return RequiresReplaceModifier{}
}

// RequiresReplaceModifier is an AttributePlanModifier that sets RequiresReplace
// on the attribute.
type RequiresReplaceModifier struct{}

// Modify fills the AttributePlanModifier interface. It sets RequiresReplace on
// the response to true if the following criteria are met:
//
// 1. The resource's state is not null; a null state indicates that we're
// creating a resource, and we never need to destroy and recreate a resource
// when we're creating it.
//
// 2. The resource's plan is not null; a null plan indicates that we're
// deleting a resource, and we never need to destroy and recreate a resource
// when we're deleting it.
//
// 3. The attribute's config is not null or the attribute is not computed; a
// computed attribute with a null config almost always means that the provider
// is changing the value, and practitioners are usually unpleasantly surprised
// when a resource is destroyed and recreated when their configuration hasn't
// changed. This has the unfortunate side effect that removing a computed field
// from the config will not trigger a destroy and recreate cycle, even when
// that is warranted. To get around this, provider developer can implement
// their own AttributePlanModifier that handles that behavior in the way that
// most makes sense for their use case.
//
// 4. The attribute's value in the plan does not match the attribute's value in
// the state.
func (r RequiresReplaceModifier) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	if req.AttributeConfig == nil || req.AttributePlan == nil || req.AttributeState == nil {
		// shouldn't happen, but let's not panic if it does
		return
	}

	if req.State.Raw.IsNull() {
		// if we're creating the resource, no need to delete and
		// recreate it
		return
	}

	if req.Plan.Raw.IsNull() {
		// if we're deleting the resource, no need to delete and
		// recreate it
		return
	}

	attrSchema, err := req.State.Schema.AttributeAtPath(req.AttributePath)

	// Path may lead to block instead of attribute. Blocks cannot be Computed.
	// If ErrPathIsBlock, attrSchema.Computed will still be false later.
	if err != nil && !errors.Is(err, ErrPathIsBlock) {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error finding attribute schema",
			fmt.Sprintf("An unexpected error was encountered retrieving the schema for this attribute. This is always a bug in the provider.\n\nError: %s", err),
		)
		return
	}

	configRaw, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error converting config value",
			fmt.Sprintf("An unexpected error was encountered converting a %s to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", req.AttributeConfig.Type(ctx), err),
		)
		return
	}
	if configRaw.IsNull() && attrSchema.Computed {
		// if the config is null and the attribute is computed, this
		// could be an out of band change, don't require replace
		return
	}

	if req.AttributePlan.Equal(req.AttributeState) {
		// if the plan and the state are in agreement, this attribute
		// isn't changing, don't require replace
		return
	}

	resp.RequiresReplace = true
}

// Description returns a human-readable description of the plan modifier.
func (r RequiresReplaceModifier) Description(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r RequiresReplaceModifier) MarkdownDescription(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

// RequiresReplaceIf returns an AttributePlanModifier that mimics
// RequiresReplace, but only when the passed function `f` returns true. The
// resource will be destroyed and recreated if `f` returns true and the
// following conditions are met:
//
// 1. The resource's state is not null; a null state indicates that we're
// creating a resource, and we never need to destroy and recreate a resource
// when we're creating it.
//
// 2. The resource's plan is not null; a null plan indicates that we're
// deleting a resource, and we never need to destroy and recreate a resource
// when we're deleting it.
//
// 3. The attribute's config is not null or the attribute is not computed; a
// computed attribute with a null config almost always means that the provider
// is changing the value, and practitioners are usually unpleasantly surprised
// when a resource is destroyed and recreated when their configuration hasn't
// changed. This has the unfortunate side effect that removing a computed field
// from the config will not trigger a destroy and recreate cycle, even when
// that is warranted. To get around this, provider developer can implement
// their own AttributePlanModifier that handles that behavior in the way that
// most makes sense for their use case.
//
// 4. The attribute's value in the plan does not match the attribute's value in
// the state.
//
// If `f` does not return true, RequiresReplaceIf will *not* override prior
// AttributePlanModifiers' determination of whether the resource needs to be
// recreated or not. This allows for multiple RequiresReplaceIf (or other
// modifiers that sometimes set RequiresReplace) to be used on a single
// attribute without the last one in the list always determining the outcome.
func RequiresReplaceIf(f RequiresReplaceIfFunc, description, markdownDescription string) AttributePlanModifier {
	return RequiresReplaceIfModifier{
		f:                   f,
		description:         description,
		markdownDescription: markdownDescription,
	}
}

// RequiresReplaceIfFunc is a conditional function used in the RequiresReplaceIf
// plan modifier to determine whether the attribute requires replacement.
type RequiresReplaceIfFunc func(ctx context.Context, state, config attr.Value, path *tftypes.AttributePath) (bool, diag.Diagnostics)

// RequiresReplaceIfModifier is an AttributePlanModifier that sets RequiresReplace
// on the attribute if the conditional function returns true.
type RequiresReplaceIfModifier struct {
	f                   RequiresReplaceIfFunc
	description         string
	markdownDescription string
}

// Modify fills the AttributePlanModifier interface. It sets RequiresReplace on
// the response to true if the following criteria are met:
//
// 1. `f` returns true. If `f` returns false, the response will not be modified
// at all.
//
// 2. The resource's state is not null; a null state indicates that we're
// creating a resource, and we never need to destroy and recreate a resource
// when we're creating it.
//
// 3. The resource's plan is not null; a null plan indicates that we're
// deleting a resource, and we never need to destroy and recreate a resource
// when we're deleting it.
//
// 4. The attribute's config is not null or the attribute is not computed; a
// computed attribute with a null config almost always means that the provider
// is changing the value, and practitioners are usually unpleasantly surprised
// when a resource is destroyed and recreated when their configuration hasn't
// changed. This has the unfortunate side effect that removing a computed field
// from the config will not trigger a destroy and recreate cycle, even when
// that is warranted. To get around this, provider developer can implement
// their own AttributePlanModifier that handles that behavior in the way that
// most makes sense for their use case.
//
// 5. The attribute's value in the plan does not match the attribute's value in
// the state.
func (r RequiresReplaceIfModifier) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	if req.AttributeConfig == nil || req.AttributePlan == nil || req.AttributeState == nil {
		// shouldn't happen, but let's not panic if it does
		return
	}

	if req.State.Raw.IsNull() {
		// if we're creating the resource, no need to delete and
		// recreate it
		return
	}

	if req.Plan.Raw.IsNull() {
		// if we're deleting the resource, no need to delete and
		// recreate it
		return
	}

	attrSchema, err := req.State.Schema.AttributeAtPath(req.AttributePath)

	// Path may lead to block instead of attribute. Blocks cannot be Computed.
	// If ErrPathIsBlock, attrSchema.Computed will still be false later.
	if err != nil && !errors.Is(err, ErrPathIsBlock) {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error finding attribute schema",
			fmt.Sprintf("An unexpected error was encountered retrieving the schema for this attribute. This is always a bug in the provider.\n\nError: %s", err),
		)
		return
	}

	configRaw, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error converting config value",
			fmt.Sprintf("An unexpected error was encountered converting a %s to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", req.AttributeConfig.Type(ctx), err),
		)
		return
	}
	if configRaw.IsNull() && attrSchema.Computed {
		// if the config is null and the attribute is computed, this
		// could be an out of band change, don't require replace
		return
	}

	if req.AttributePlan.Equal(req.AttributeState) {
		// if the plan and the state are in agreement, this attribute
		// isn't changing, don't require replace
		return
	}

	res, diags := r.f(ctx, req.AttributeState, req.AttributeConfig, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	// If the function says to require replacing, we require replacing.
	// If the function says not to, we don't change the value that prior
	// plan modifiers may have set.
	if res {
		resp.RequiresReplace = true
	} else if resp.RequiresReplace {
		tfsdklog.Debug(ctx, "Keeping previous attribute replacement requirement", map[string]interface{}{"attribute_path": req.AttributePath.String()})
	}
}

// Description returns a human-readable description of the plan modifier.
func (r RequiresReplaceIfModifier) Description(ctx context.Context) string {
	return r.description
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r RequiresReplaceIfModifier) MarkdownDescription(ctx context.Context) string {
	return r.markdownDescription
}

// UseStateForUnknown returns a UseStateForUnknownModifier.
func UseStateForUnknown() AttributePlanModifier {
	return UseStateForUnknownModifier{}
}

// UseStateForUnknownModifier is an AttributePlanModifier that copies the prior state
// value for an attribute into that attribute's plan, if that state is non-null.
//
// Computed attributes without the UseStateForUnknown attribute plan modifier will
// have their value set to Unknown in the plan, so their value always will be
// displayed as "(known after apply)" in the CLI plan output.
// If this plan modifier is used, the prior state value will be displayed in
// the plan instead unless a prior plan modifier adjusts the value.
type UseStateForUnknownModifier struct{}

// Modify copies the attribute's prior state to the attribute plan if the prior
// state value is not null.
func (r UseStateForUnknownModifier) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	val, err := req.AttributeState.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error converting state value",
			fmt.Sprintf("An unexpected error was encountered converting a %s to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", req.AttributeState.Type(ctx), err),
		)
		return
	}

	// if we have no state value, there's nothing to preserve
	if val.IsNull() {
		return
	}

	val, err = resp.AttributePlan.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error converting plan value",
			fmt.Sprintf("An unexpected error was encountered converting a %s to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", resp.AttributePlan.Type(ctx), err),
		)
		return
	}

	// if it's not planned to be the unknown value, stick with
	// the concrete plan
	if val.IsKnown() {
		return
	}

	val, err = req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error converting config value",
			fmt.Sprintf("An unexpected error was encountered converting a %s to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", req.AttributeConfig.Type(ctx), err),
		)
		return
	}

	// if the config is the unknown value, use the unknown value
	// otherwise, interpolation gets messed up
	if !val.IsKnown() {
		return
	}

	resp.AttributePlan = req.AttributeState
}

// Description returns a human-readable description of the plan modifier.
func (r UseStateForUnknownModifier) Description(ctx context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r UseStateForUnknownModifier) MarkdownDescription(ctx context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// ModifyAttributePlanRequest represents a request for the provider to modify an
// attribute value, or mark it as requiring replacement, at plan time. An
// instance of this request struct is supplied as an argument to the Modify
// function of an attribute's plan modifier(s).
type ModifyAttributePlanRequest struct {
	// AttributePath is the path of the attribute.
	AttributePath *tftypes.AttributePath

	// Config is the configuration the user supplied for the resource.
	Config Config

	// State is the current state of the resource.
	State State

	// Plan is the planned new state for the resource.
	Plan Plan

	// AttributeConfig is the configuration the user supplied for the attribute.
	AttributeConfig attr.Value

	// AttributeState is the current state of the attribute.
	AttributeState attr.Value

	// AttributePlan is the planned new state for the attribute.
	AttributePlan attr.Value

	// ProviderMeta is metadata from the provider_meta block of the module.
	ProviderMeta Config
}

// ModifyAttributePlanResponse represents a response to a
// ModifyAttributePlanRequest. An instance of this response struct is supplied
// as an argument to the Modify function of an attribute's plan modifier(s).
type ModifyAttributePlanResponse struct {
	// AttributePlan is the planned new state for the attribute.
	AttributePlan attr.Value

	// RequiresReplace indicates whether a change in the attribute
	// requires replacement of the whole resource.
	RequiresReplace bool

	// Diagnostics report errors or warnings related to determining the
	// planned state of the requested resource. Returning an empty slice
	// indicates a successful validation with no warnings or errors
	// generated.
	Diagnostics diag.Diagnostics
}
