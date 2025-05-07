package toproto6

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

func NewInvokeActionCallBackServer(protov6CallbackServer tfprotov6.InvokeActionCallBackServer) action.InvokeActionCallBackServer {
	return &InvokeActionCallBackServer{
		proto6CallbackServer: protov6CallbackServer,
	}
}

type InvokeActionCallBackServer struct {
	proto6CallbackServer tfprotov6.InvokeActionCallBackServer
}

func (i *InvokeActionCallBackServer) Send(ctx context.Context, event action.InvokeActionEvent) error {
	switch actionEvent := event.(type) {
	case *action.StartedActionEvent:
		tfprotov6Event := StartedActionEvent(ctx, actionEvent)
		return i.proto6CallbackServer.Send(ctx, tfprotov6Event)
	case *action.ProgressActionEvent:
		tfprotov6Event := ProgressActionEvent(ctx, actionEvent)
		return i.proto6CallbackServer.Send(ctx, tfprotov6Event)
	case *action.FinishedActionEvent:
		tfprotov6Event := FinishedActionEvent(ctx, actionEvent)
		return i.proto6CallbackServer.Send(ctx, tfprotov6Event)
	case *action.CancelledActionEvent:
		tfprotov6Event := CancelledActionEvent(ctx, actionEvent)
		return i.proto6CallbackServer.Send(ctx, tfprotov6Event)
	default:
		return fmt.Errorf("unknown InvokeActionEvent type: %T", actionEvent)
	}
}

func StartedActionEvent(ctx context.Context, in *action.StartedActionEvent) *tfprotov6.StartedActionEvent {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.StartedActionEvent{
		CancellationToken: in.CancellationToken,
		Diagnostics:       Diagnostics(ctx, in.Diagnostics),
	}

	return resp
}

func ProgressActionEvent(ctx context.Context, in *action.ProgressActionEvent) *tfprotov6.ProgressActionEvent {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.ProgressActionEvent{
		StdOut:      in.StdOut,
		StdErr:      in.StdErr,
		Diagnostics: Diagnostics(ctx, in.Diagnostics),
	}

	return resp
}

func FinishedActionEvent(ctx context.Context, in *action.FinishedActionEvent) *tfprotov6.FinishedActionEvent {
	if in == nil {
		return nil
	}

	tfprotoVal, diags := State(ctx, in.NewConfig)
	in.Diagnostics.Append(diags...)

	resp := &tfprotov6.FinishedActionEvent{
		Outputs:     map[string]*tfprotov6.DynamicValue{},
		NewConfig:   tfprotoVal,
		Diagnostics: Diagnostics(ctx, in.Diagnostics),
	}

	return resp
}

func CancelledActionEvent(ctx context.Context, in *action.CancelledActionEvent) *tfprotov6.CancelledActionEvent {
	if in == nil {
		return nil
	}

	resp := &tfprotov6.CancelledActionEvent{
		Diagnostics: Diagnostics(ctx, in.Diagnostics),
	}

	return resp
}
