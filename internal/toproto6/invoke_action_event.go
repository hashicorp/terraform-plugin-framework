// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func ProgressInvokeActionEventType(ctx context.Context, event fwserver.InvokeProgressEvent) tfprotov6.InvokeActionEvent {
	return tfprotov6.InvokeActionEvent{
		Type: tfprotov6.ProgressInvokeActionEventType{
			Message: event.Message,
		},
	}
}

func CompletedInvokeActionEventType(ctx context.Context, fw *fwserver.InvokeActionResponse) tfprotov6.InvokeActionEvent {
	completedEvent := tfprotov6.CompletedInvokeActionEventType{
		Diagnostics: Diagnostics(ctx, fw.Diagnostics),
	}

	completedEvent.LinkedResources = make([]*tfprotov6.NewLinkedResource, len(fw.LinkedResources))

	for i, linkedResource := range fw.LinkedResources {
		newState, diags := State(ctx, linkedResource.NewState)
		completedEvent.Diagnostics = append(completedEvent.Diagnostics, Diagnostics(ctx, diags)...)

		newIdentity, diags := ResourceIdentity(ctx, linkedResource.NewIdentity)
		completedEvent.Diagnostics = append(completedEvent.Diagnostics, Diagnostics(ctx, diags)...)

		completedEvent.LinkedResources[i] = &tfprotov6.NewLinkedResource{
			NewState:        newState,
			NewIdentity:     newIdentity,
			RequiresReplace: linkedResource.RequiresReplace,
		}
	}

	return tfprotov6.InvokeActionEvent{
		Type: completedEvent,
	}
}
