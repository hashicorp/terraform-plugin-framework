// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

var _ action.Action = &ActionWithConfigure{}
var _ action.ActionWithConfigure = &ActionWithConfigure{}

// Declarative action.ActionWithConfigure for unit testing.
type ActionWithConfigure struct {
	*Action

	// ActionWithConfigure interface methods
	ConfigureMethod func(context.Context, action.ConfigureRequest, *action.ConfigureResponse)
}

// Configure satisfies the action.ActionWithConfigure interface.
func (r *ActionWithConfigure) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}
