// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ModifyPlanClientCapabilities allows Terraform to publish information
// regarding optionally supported protocol features for the PlanAction RPC,
// such as forward-compatible Terraform behavior changes.
type ModifyPlanClientCapabilities struct {
	// DeferralAllowed indicates whether the Terraform client initiating
	// the request allows a deferral response.
	//
	// NOTE: This functionality is related to deferred action support, which is currently experimental and is subject
	// to change or break without warning. It is not protected by version compatibility guarantees.
	DeferralAllowed bool
}

// ModifyPlanRequest represents a request for the provider to modify the
// planned new state that Terraform has generated for any linked resources.
type ModifyPlanRequest struct {
	// Config is the configuration the user supplied for the action.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config tfsdk.Config

	// LinkedResources contains the data of the managed resource types that are linked to this action.
	//
	//   - If the action schema type is Unlinked, this field will be empty.
	//   - If the action schema type is Lifecycle, this field will contain a single linked resource.
	//   - If the action schema type is Linked, this field will be one or more linked resources, which
	//     will be in the same order as the linked resource schemas are defined in the action schema.
	//
	// For Lifecycle actions, the provider may only change computed-only attributes of the linked resources.
	// For Linked actions, the provider may change any attributes of the linked resources.
	LinkedResources []ModifyPlanRequestLinkedResource

	// ClientCapabilities defines optionally supported protocol features for the
	// PlanAction RPC, such as forward-compatible Terraform behavior changes.
	ClientCapabilities ModifyPlanClientCapabilities
}

// ModifyPlanRequestLinkedResource represents linked resource data prior to the action plan.
type ModifyPlanRequestLinkedResource struct {
	// Config is the configuration the user supplied for the linked resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config tfsdk.Config

	// State is the current state of the linked resource.
	State tfsdk.State

	// Identity is the current identity of the linked resource. If the linked resource does not
	// support identity, this value will not be set.
	Identity *tfsdk.ResourceIdentity

	// Plan is the latest planned new state for the linked resource. This could
	// be the result of the linked resource plan or a plan from a predecessor action.
	Plan tfsdk.Plan
}

// ModifyPlanResponse represents a response to a
// ModifyPlanRequest. An instance of this response struct is supplied
// as an argument to the action's ModifyPlan function, in which the provider
// should modify the Plan of any linked resources as appropriate.
type ModifyPlanResponse struct {
	// Diagnostics report errors or warnings related to determining the
	// planned state of the requested action's linked resources. Returning an empty slice
	// indicates a successful plan modification with no warnings or errors
	// generated.
	Diagnostics diag.Diagnostics

	// LinkedResources contains the provider modified data of the managed resource types that are linked to this action.
	//
	// For Lifecycle actions, the provider may only change computed-only attributes of the linked resources.
	// For Linked actions, the provider may change any attributes of the linked resources.
	LinkedResources []ModifyPlanResponseLinkedResource

	// Deferred indicates that Terraform should defer planning this
	// action until a follow-up apply operation.
	//
	// This field can only be set if
	// `(action.ModifyPlanRequest).ClientCapabilities.DeferralAllowed` is true.
	//
	// NOTE: This functionality is related to deferred action support, which is currently experimental and is subject
	// to change or break without warning. It is not protected by version compatibility guarantees.
	Deferred *Deferred
}

// ModifyPlanResponseLinkedResource represents linked resource data that was planned by the action.
type ModifyPlanResponseLinkedResource struct {
	// Plan is the planned new state for the linked resource.
	//
	// For Lifecycle actions, the provider may only change computed-only attributes of the linked resources.
	// For Linked actions, the provider may change any attributes of the linked resources.
	Plan tfsdk.Plan

	// Identity is the planned new identity of the resource.
	// This field is pre-populated from ModifyPlanRequest.Identity.
	Identity *tfsdk.ResourceIdentity
}
