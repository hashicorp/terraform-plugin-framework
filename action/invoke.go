// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type InvokeRequest struct {
	Config tfsdk.Config
}

type InvokeResponse struct {
	CancellationToken string
	CallbackServer    InvokeActionCallBackServer
	Diagnostics       diag.Diagnostics
}

type InvokeActionCallBackServer interface {
	Send(ctx context.Context, event InvokeActionEvent) error
}
