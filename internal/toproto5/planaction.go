// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
)

// PlanActionResponse returns the *tfprotov5.PlanActionResponse equivalent of a *fwserver.PlanActionResponse.
func PlanActionResponse(ctx context.Context, fw *fwserver.PlanActionResponse) *tfprotov5.PlanActionResponse {
	if fw == nil {
		return nil
	}

	proto5 := &tfprotov5.PlanActionResponse{
		Diagnostics: Diagnostics(ctx, fw.Diagnostics),
		Deferred:    ActionDeferred(fw.Deferred),
	}

	proto5.LinkedResources = make([]*tfprotov5.PlannedLinkedResource, len(fw.LinkedResources))

	for i, linkedResource := range fw.LinkedResources {
		plannedState, diags := State(ctx, linkedResource.PlannedState)
		proto5.Diagnostics = append(proto5.Diagnostics, Diagnostics(ctx, diags)...)

		plannedIdentity, diags := ResourceIdentity(ctx, linkedResource.PlannedIdentity)
		proto5.Diagnostics = append(proto5.Diagnostics, Diagnostics(ctx, diags)...)

		proto5.LinkedResources[i] = &tfprotov5.PlannedLinkedResource{
			PlannedState:    plannedState,
			PlannedIdentity: plannedIdentity,
		}
	}

	return proto5
}
