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

// InvokeActionRequest is the framework server request for the InvokeAction RPC.
type InvokeActionRequest struct {
	Action          action.Action
	ActionSchema    fwschema.Schema
	Config          *tfsdk.Config
	LinkedResources []*InvokeActionRequestLinkedResource
}

type InvokeActionRequestLinkedResource struct {
	Config          *tfsdk.Config
	PlannedState    *tfsdk.Plan
	PriorState      *tfsdk.State
	PlannedIdentity *tfsdk.ResourceIdentity
}

// InvokeActionEventsStream is the framework server stream for the InvokeAction RPC.
type InvokeActionResponse struct {
	// ProgressEvents is a channel provided by the consuming proto{5/6}server implementation
	// that allows the provider developers to return progress events while the action is being invoked.
	ProgressEvents  chan InvokeProgressEvent
	Diagnostics     diag.Diagnostics
	LinkedResources []*InvokeActionResponseLinkedResource
}

type InvokeActionResponseLinkedResource struct {
	NewState        *tfsdk.State
	NewIdentity     *tfsdk.ResourceIdentity
	RequiresReplace bool
}

type InvokeProgressEvent struct {
	Message string
}

// SendProgress is injected into the action.InvokeResponse for use by the provider developer
func (r *InvokeActionResponse) SendProgress(event action.InvokeProgressEvent) {
	r.ProgressEvents <- InvokeProgressEvent{
		Message: event.Message,
	}
}

// InvokeAction implements the framework server InvokeAction RPC.
func (s *Server) InvokeAction(ctx context.Context, req *InvokeActionRequest, resp *InvokeActionResponse) {
	if req == nil {
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

	invokeReq := action.InvokeRequest{
		Config:          *req.Config,
		LinkedResources: make([]action.InvokeRequestLinkedResource, len(req.LinkedResources)),
	}
	invokeResp := action.InvokeResponse{
		SendProgress:    resp.SendProgress,
		LinkedResources: make([]action.InvokeResponseLinkedResource, len(req.LinkedResources)),
	}

	// Pass-through the new state and identity to the response for each linked resource
	resp.LinkedResources = make([]*InvokeActionResponseLinkedResource, len(req.LinkedResources))
	for i, lr := range req.LinkedResources {
		// Initialize new state as a null object
		newState := &tfsdk.State{
			Schema: lr.PlannedState.Schema,
			Raw:    tftypes.NewValue(lr.PlannedState.Schema.Type().TerraformType(ctx), nil),
		}

		// Depending on when the action is run, prior state will either be the last read of
		// the resource (which could be null, if creating) or the final new state from ApplyResourceChange.
		//
		// If we have a prior state, use that as the default new state.
		if lr.PriorState != nil {
			newState = lr.PriorState
		}

		// Copy new state, config, plan and identity
		resp.LinkedResources[i] = &InvokeActionResponseLinkedResource{
			NewState: newState,
		}
		invokeReq.LinkedResources[i] = action.InvokeRequestLinkedResource{
			Config: *lr.Config,
			State:  *newState,
			Plan:   *lr.PlannedState,
		}
		invokeResp.LinkedResources[i] = action.InvokeResponseLinkedResource{
			State: *newState,
		}

		if lr.PlannedIdentity != nil {
			resp.LinkedResources[i].NewIdentity = &tfsdk.ResourceIdentity{
				Schema: lr.PlannedIdentity.Schema,
				Raw:    lr.PlannedIdentity.Raw.Copy(),
			}
			invokeReq.LinkedResources[i].Identity = &tfsdk.ResourceIdentity{
				Schema: lr.PlannedIdentity.Schema,
				Raw:    lr.PlannedIdentity.Raw.Copy(),
			}
			invokeResp.LinkedResources[i].Identity = &tfsdk.ResourceIdentity{
				Schema: lr.PlannedIdentity.Schema,
				Raw:    lr.PlannedIdentity.Raw.Copy(),
			}
		}
	}

	logging.FrameworkTrace(ctx, "Calling provider defined Action Invoke")
	req.Action.Invoke(ctx, invokeReq, &invokeResp)
	logging.FrameworkTrace(ctx, "Called provider defined Action Invoke")

	resp.Diagnostics = invokeResp.Diagnostics

	if len(resp.LinkedResources) != len(invokeResp.LinkedResources) {
		resp.Diagnostics.AddError(
			"Invalid Linked Resource State",
			"An unexpected error was encountered when invoking an action with linked resources. "+
				fmt.Sprintf(
					"The number of linked resource states produced by the action invoke cannot change: %d linked resource(s) were planned, expected %d\n\n",
					len(invokeResp.LinkedResources),
					len(resp.LinkedResources),
				)+
				"This is always a problem with the provider and should be reported to the provider developer.",
		)
		return
	}

	processingDiags := make(diag.Diagnostics, 0)
	for i, newLinkedResource := range invokeResp.LinkedResources {
		resp.LinkedResources[i].NewState = &newLinkedResource.State
		resp.LinkedResources[i].NewIdentity = newLinkedResource.Identity
		resp.LinkedResources[i].RequiresReplace = newLinkedResource.RequiresReplace

		if !resp.Diagnostics.HasError() && resp.LinkedResources[i].RequiresReplace {
			processingDiags.AddError(
				"Invalid Linked Resource Replacement",
				"An unexpected error was encountered when invoking an action with linked resources. "+
					fmt.Sprintf("The Terraform Provider returned a linked resource (at index %d) that "+
						"indicates that it needs to be replaced, but no error diagnostics were returned.\n\n"+
						"This is always a problem with the provider and should be reported to the provider developer.", i),
			)

			// Continue processing the rest of the linked resources
			continue
		}
	}

	resp.Diagnostics.Append(processingDiags...)
}
