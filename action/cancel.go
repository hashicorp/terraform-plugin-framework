// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CancelRequest represents a request for the provider to cancel a
// resource. An instance of this request struct is supplied as an argument to
// the resource's Cancel function.
type CancelRequest struct {
	CancellationToken string
	// CancellationFunc cancels all contexts associated with the given cancellation token
	CancellationFunc CancellationFunc
}

// CancelResponse represents a response to a CancelRequest. An
// instance of this response struct is supplied as
// an argument to the resource's Cancel function, in which the provider
// should set values on the CancelResponse as appropriate.
type CancelResponse struct {
	// InvokeActionCallBackServer is the callback server associated with the
	// cancellation token. Use this to send ProgressActionEvent and CancelledActionEvent
	// associated with the rollback and cancellation.
	InvokeActionCallBackServer InvokeActionCallBackServer
	Diagnostics                diag.Diagnostics
}

type CancellationFunc interface {
	Cancel(ctx context.Context, CancellationToken string) error
}
