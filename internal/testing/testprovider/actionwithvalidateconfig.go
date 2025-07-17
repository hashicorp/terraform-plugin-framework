// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

var _ action.Action = &ActionWithValidateConfig{}
var _ action.ActionWithValidateConfig = &ActionWithValidateConfig{}

// Declarative action.ActionWithValidateConfig for unit testing.
type ActionWithValidateConfig struct {
	*Action

	// ActionWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, action.ValidateConfigRequest, *action.ValidateConfigResponse)
}

// ValidateConfig satisfies the action.ActionWithValidateConfig interface.
func (p *ActionWithValidateConfig) ValidateConfig(ctx context.Context, req action.ValidateConfigRequest, resp *action.ValidateConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
