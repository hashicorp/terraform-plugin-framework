package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
// in terraform-plugin-sdk.
func RequiresReplace() AttributePlanModifier {
	return RequiresReplaceModifier{}
}

// RequiresReplaceModifier is an AttributePlanModifier that sets RequiresReplace
// on the attribute.
type RequiresReplaceModifier struct{}

// Modify sets RequiresReplace on the response to true.
func (r RequiresReplaceModifier) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
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

// RequiresReplaceIf returns an AttributePlanModifier that runs the conditional
// function f: if it returns true, it specifies the attribute as requiring
// replacement.
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

// Modify sets RequiresReplace on the response to true if the conditional
// RequiresReplaceIfFunc returns true.
func (r RequiresReplaceIfModifier) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	res, diags := r.f(ctx, req.AttributeState, req.AttributeConfig, req.AttributePath)
	resp.Diagnostics.Append(diags...)
	resp.RequiresReplace = res
}

// Description returns a human-readable description of the plan modifier.
func (r RequiresReplaceIfModifier) Description(ctx context.Context) string {
	return r.description
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r RequiresReplaceIfModifier) MarkdownDescription(ctx context.Context) string {
	return r.markdownDescription
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

// AddWarning appends a warning diagnostic to the response. If the warning
// concerns a particular attribute, AddAttributeWarning should be used instead.
func (r *ModifyAttributePlanResponse) AddWarning(summary, detail string) {
	r.Diagnostics.AddWarning(summary, detail)
}

// AddAttributeWarning appends a warning diagnostic to the response and labels
// it with a specific attribute.
func (r *ModifyAttributePlanResponse) AddAttributeWarning(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeWarning(attributePath, summary, detail)
}

// AddError appends an error diagnostic to the response. If the error concerns a
// particular attribute, AddAttributeError should be used instead.
func (r *ModifyAttributePlanResponse) AddError(summary, detail string) {
	r.Diagnostics.AddError(summary, detail)
}

// AddAttributeError appends an error diagnostic to the response and labels it
// with a specific attribute.
func (r *ModifyAttributePlanResponse) AddAttributeError(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeError(attributePath, summary, detail)
}
