package action

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Invoke Action Events
type InvokeActionEvent interface {
	// TODO: make this interface unfillable to restrict implementations
	isInvokeActionEvent()
}

var _ InvokeActionEvent = &StartedActionEvent{}

type StartedActionEvent struct {
	CancellationToken string
	Diagnostics       diag.Diagnostics
}

func (s *StartedActionEvent) isInvokeActionEvent() {}

type FinishedActionEvent struct {
	NewConfig   *tfsdk.State
	Diagnostics diag.Diagnostics
}

func (f *FinishedActionEvent) isInvokeActionEvent() {}

type CancelledActionEvent struct {
	Diagnostics diag.Diagnostics
}

func (c *CancelledActionEvent) isInvokeActionEvent() {}

type ProgressActionEvent struct {
	StdOut      []string
	StdErr      []string
	Diagnostics diag.Diagnostics
}

var _ InvokeActionEvent = &ProgressActionEvent{}

func (p *ProgressActionEvent) isInvokeActionEvent() {}
