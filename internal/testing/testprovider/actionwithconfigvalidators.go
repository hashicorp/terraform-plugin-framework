// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

var _ action.Action = &ActionWithConfigValidators{}
var _ action.ActionWithConfigValidators = &ActionWithConfigValidators{}

// Declarative action.ActionWithConfigValidators for unit testing.
type ActionWithConfigValidators struct {
	*Action

	// ActionWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []action.ConfigValidator
}

// ConfigValidators satisfies the action.ActionWithConfigValidators interface.
func (p *ActionWithConfigValidators) ConfigValidators(ctx context.Context) []action.ConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
