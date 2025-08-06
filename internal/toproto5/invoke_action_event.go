// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func ProgressInvokeActionEventType(ctx context.Context, event fwserver.InvokeProgressEvent) tfprotov5.InvokeActionEvent {
	return tfprotov5.InvokeActionEvent{
		Type: tfprotov5.ProgressInvokeActionEventType{
			Message: event.Message,
		},
	}
}

func CompletedInvokeActionEventType(ctx context.Context, fw *fwserver.InvokeActionResponse) tfprotov5.InvokeActionEvent {
	completedEvent := tfprotov5.CompletedInvokeActionEventType{
		Diagnostics: Diagnostics(ctx, fw.Diagnostics),
	}

	completedEvent.LinkedResources = make([]*tfprotov5.NewLinkedResource, len(fw.LinkedResources))

	for i, linkedResource := range fw.LinkedResources {
		newState, diags := State(ctx, linkedResource.NewState)
		completedEvent.Diagnostics = append(completedEvent.Diagnostics, Diagnostics(ctx, diags)...)

		newIdentity, diags := ResourceIdentity(ctx, linkedResource.NewIdentity)
		completedEvent.Diagnostics = append(completedEvent.Diagnostics, Diagnostics(ctx, diags)...)

		completedEvent.LinkedResources[i] = &tfprotov5.NewLinkedResource{
			NewState:        newState,
			NewIdentity:     newIdentity,
			RequiresReplace: linkedResource.RequiresReplace,
		}
	}

	return tfprotov5.InvokeActionEvent{
		Type: completedEvent,
	}
}
