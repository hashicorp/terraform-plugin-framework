// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
)

var _ list.ListResource = &ListResourceWithConfigValidators{}
var _ list.ListResourceWithConfigValidators = &ListResourceWithConfigValidators{}

// Declarative list.ListResourceWithConfigValidators for unit testing.
type ListResourceWithConfigValidators struct {
	*ListResource

	// ListResourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []list.ConfigValidator
}

// ConfigValidators satisfies the list.ListResourceWithConfigValidators interface.
func (p *ListResourceWithConfigValidators) ListResourceConfigValidators(ctx context.Context) []list.ConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
