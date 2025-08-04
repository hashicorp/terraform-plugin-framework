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
	LinkedResources    []*PlanActionLinkedResourceRequest
}

type PlanActionLinkedResourceRequest struct {
	Config        *tfsdk.Config
	PlannedState  *tfsdk.Plan
	PriorState    *tfsdk.State
	PriorIdentity *tfsdk.ResourceIdentity
}

// PlanActionResponse is the framework server response for the PlanAction RPC.
type PlanActionResponse struct {
	Deferred        *action.Deferred
	Diagnostics     diag.Diagnostics
	LinkedResources []*PlanActionLinkedResourceResponse
}

type PlanActionLinkedResourceResponse struct {
	PlannedState    *tfsdk.State
	PlannedIdentity *tfsdk.ResourceIdentity
}

// PlanAction implements the framework server PlanAction RPC.
func (s *Server) PlanAction(ctx context.Context, req *PlanActionRequest, resp *PlanActionResponse) {
	if req == nil {
		return
	}

	// Copy over planned state and identity to the response for each linked resource as a default plan
	resp.LinkedResources = make([]*PlanActionLinkedResourceResponse, len(req.LinkedResources))
	for i, lr := range req.LinkedResources {
		resp.LinkedResources[i] = &PlanActionLinkedResourceResponse{
			PlannedState: planToState(*lr.PlannedState),
		}

		if lr.PriorIdentity != nil {
			resp.LinkedResources[i].PlannedIdentity = &tfsdk.ResourceIdentity{
				Schema: lr.PriorIdentity.Schema,
				Raw:    lr.PriorIdentity.Raw.Copy(),
			}
		}
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
						"The number of linked resources produced by the action plan cannot change: %d linked resource(s) were produced in the plan, expected %d\n\n",
						len(modifyPlanResp.LinkedResources),
						len(resp.LinkedResources),
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
