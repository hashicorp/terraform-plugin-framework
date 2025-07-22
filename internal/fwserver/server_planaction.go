// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// PlanActionRequest is the framework server request for the PlanAction RPC.
type PlanActionRequest struct {
	ClientCapabilities action.ModifyPlanClientCapabilities
	ActionSchema       fwschema.Schema
	Action             action.Action
	Config             *tfsdk.Config

	// TODO:Actions: Should we introduce another layer on top of this? To protect against index-oob and prevent invalid setting of data? (depending on the action schema)
	//
	// Could just introduce a new tfsdk.State that is more restricted? tfsdk.LinkedResourceState?
	// Theoretically, we also need the action schema itself, since there are different rules for each.
	// Should we just let Terraform core handle all the validation themselves? That's how it's done today.
	LinkedResources []*PlanLinkedResourceRequest // TODO:Actions: Should this be a pointer?
}

type PlanLinkedResourceRequest struct {
	Config        *tfsdk.Config
	PlannedState  *tfsdk.Plan
	PriorState    *tfsdk.State
	PriorIdentity *tfsdk.ResourceIdentity
}

// PlanActionResponse is the framework server response for the PlanAction RPC.
type PlanActionResponse struct {
	Deferred    *action.Deferred
	Diagnostics diag.Diagnostics

	LinkedResources []*PlanLinkedResourceResponse // TODO:Actions: Should this be a pointer?
}

type PlanLinkedResourceResponse struct {
	PlannedState    *tfsdk.State
	PlannedIdentity *tfsdk.ResourceIdentity
}

// PlanAction implements the framework server PlanAction RPC.
func (s *Server) PlanAction(ctx context.Context, req *PlanActionRequest, resp *PlanActionResponse) {
	if req == nil {
		return
	}

	if s.deferred != nil {
		logging.FrameworkDebug(ctx, "Provider has deferred response configured, automatically returning deferred response.",
			map[string]interface{}{
				logging.KeyDeferredReason: s.deferred.Reason.String(),
			},
		)

		resp.Deferred = &action.Deferred{
			Reason: action.DeferredReason(s.deferred.Reason),
		}
		return
	}

	if actionWithConfigure, ok := req.Action.(action.ActionWithConfigure); ok {
		logging.FrameworkTrace(ctx, "Action implements ActionWithConfigure")

		configureReq := action.ConfigureRequest{
			ProviderData: s.ActionConfigureData,
		}
		configureResp := action.ConfigureResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined Action Configure")
		actionWithConfigure.Configure(ctx, configureReq, &configureResp)
		logging.FrameworkTrace(ctx, "Called provider defined Action Configure")

		resp.Diagnostics.Append(configureResp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	if req.Config == nil {
		req.Config = &tfsdk.Config{
			Raw:    tftypes.NewValue(req.ActionSchema.Type().TerraformType(ctx), nil),
			Schema: req.ActionSchema,
		}
	}

	// By default, copy over planned state and identity for each linked resource
	resp.LinkedResources = make([]*PlanLinkedResourceResponse, len(req.LinkedResources))
	for i, lr := range req.LinkedResources {
		if lr.PlannedState == nil {
			// TODO:Actions: I'm not 100% sure if this is valid enough to be a concern, PlanResourceChange populates this with a null
			// value of the resource schema type, but it'd be nice to not have to carry linked resource schemas this far
			// if we don't need them.
			//
			// My current thought is that this isn't needed (a similar check would need to be done on identity). Specifically because
			// actions should always be following a linked resource PlanResourceChange call. So this value should always be populated and
			// this would more be protecting future logic from panicking if a bug existing in Terraform core or Framework/SDKv2.
			resp.Diagnostics.AddError(
				"Invalid PlannedState for Linked Resource",
				"An unexpected error was encountered when planning an action with linked resources. "+
					fmt.Sprintf("Linked resource planned state was nil when received in the protocol, index: %d.\n\n", i)+
					"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
			)
			return
		}

		resp.LinkedResources[i] = &PlanLinkedResourceResponse{
			PlannedState: planToState(*lr.PlannedState),
		}

		if lr.PriorIdentity != nil {
			resp.LinkedResources[i].PlannedIdentity = &tfsdk.ResourceIdentity{
				Schema: lr.PriorIdentity.Schema,
				Raw:    lr.PriorIdentity.Raw.Copy(),
			}
		}
	}

	// TODO:Actions: Should we add support for schema plan modifiers? Technically you could re-use any framework plan modifier
	// implementations from the "resource/schema/planmodifier" package

	if actionWithModifyPlan, ok := req.Action.(action.ActionWithModifyPlan); ok {
		logging.FrameworkTrace(ctx, "Action implements ActionWithModifyPlan")

		modifyPlanReq := action.ModifyPlanRequest{
			ClientCapabilities: req.ClientCapabilities,
			Config:             *req.Config,
			LinkedResources:    make([]action.ModifyPlanRequestLinkedResource, len(req.LinkedResources)),
		}

		modifyPlanResp := action.ModifyPlanResponse{
			Diagnostics:     resp.Diagnostics,
			LinkedResources: make([]action.ModifyPlanResponseLinkedResource, len(req.LinkedResources)),
		}

		for i, linkedResource := range req.LinkedResources {
			modifyPlanReq.LinkedResources[i] = action.ModifyPlanRequestLinkedResource{
				Config: *linkedResource.Config,
				Plan:   stateToPlan(*resp.LinkedResources[i].PlannedState),
				State:  *linkedResource.PriorState,
			}
			modifyPlanResp.LinkedResources[i] = action.ModifyPlanResponseLinkedResource{
				Plan: modifyPlanReq.LinkedResources[i].Plan,
			}

			if resp.LinkedResources[i].PlannedIdentity != nil {
				modifyPlanReq.LinkedResources[i].Identity = &tfsdk.ResourceIdentity{
					Schema: resp.LinkedResources[i].PlannedIdentity.Schema,
					Raw:    resp.LinkedResources[i].PlannedIdentity.Raw.Copy(),
				}
				modifyPlanResp.LinkedResources[i].Identity = &tfsdk.ResourceIdentity{
					Schema: resp.LinkedResources[i].PlannedIdentity.Schema,
					Raw:    resp.LinkedResources[i].PlannedIdentity.Raw.Copy(),
				}
			}
		}

		logging.FrameworkTrace(ctx, "Calling provider defined Action ModifyPlan")
		actionWithModifyPlan.ModifyPlan(ctx, modifyPlanReq, &modifyPlanResp)
		logging.FrameworkTrace(ctx, "Called provider defined Action ModifyPlan")

		resp.Diagnostics = modifyPlanResp.Diagnostics
		resp.Deferred = modifyPlanResp.Deferred

		if len(resp.LinkedResources) != len(modifyPlanResp.LinkedResources) {
			resp.Diagnostics.AddError(
				"Invalid Linked Resource Plan",
				"An unexpected error was encountered when planning an action with linked resources. "+
					fmt.Sprintf(
						"The number of linked resources planned cannot change, expected: %d, got: %d\n\n",
						len(resp.LinkedResources),
						len(modifyPlanResp.LinkedResources),
					)+
					"This is always a problem with the provider and should be reported to the provider developer.",
			)
			return
		}

		for i, newLinkedResource := range modifyPlanResp.LinkedResources {
			resp.LinkedResources[i].PlannedState = planToState(newLinkedResource.Plan)
			resp.LinkedResources[i].PlannedIdentity = newLinkedResource.Identity
		}
	}
}
