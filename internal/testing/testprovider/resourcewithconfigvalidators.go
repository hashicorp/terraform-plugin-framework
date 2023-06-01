// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithConfigValidators{}
var _ resource.ResourceWithConfigValidators = &ResourceWithConfigValidators{}

// Declarative resource.ResourceWithConfigValidators for unit testing.
type ResourceWithConfigValidators struct {
	*Resource

	// ResourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []resource.ConfigValidator
}

// ConfigValidators satisfies the resource.ResourceWithConfigValidators interface.
func (p *ResourceWithConfigValidators) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
