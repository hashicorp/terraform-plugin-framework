// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
)

// PlanActionResponse returns the *tfprotov6.PlanActionResponse equivalent of a *fwserver.PlanActionResponse.
func PlanActionResponse(ctx context.Context, fw *fwserver.PlanActionResponse) *tfprotov6.PlanActionResponse {
	if fw == nil {
		return nil
	}

	proto6 := &tfprotov6.PlanActionResponse{
		Diagnostics: Diagnostics(ctx, fw.Diagnostics),
		Deferred:    ActionDeferred(fw.Deferred),
	}

	proto6.LinkedResources = make([]*tfprotov6.PlannedLinkedResource, len(fw.LinkedResources))

	for i, linkedResource := range fw.LinkedResources {
		plannedState, diags := State(ctx, linkedResource.PlannedState)
		proto6.Diagnostics = append(proto6.Diagnostics, Diagnostics(ctx, diags)...)

		plannedIdentity, diags := ResourceIdentity(ctx, linkedResource.PlannedIdentity)
		proto6.Diagnostics = append(proto6.Diagnostics, Diagnostics(ctx, diags)...)

		proto6.LinkedResources[i] = &tfprotov6.PlannedLinkedResource{
			PlannedState:    plannedState,
			PlannedIdentity: plannedIdentity,
		}
	}

	return proto6
}
