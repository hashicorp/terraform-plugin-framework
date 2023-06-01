// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithValidateConfig{}
var _ resource.ResourceWithValidateConfig = &ResourceWithValidateConfig{}

// Declarative resource.ResourceWithValidateConfig for unit testing.
type ResourceWithValidateConfig struct {
	*Resource

	// ResourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse)
}

// ValidateConfig satisfies the resource.ResourceWithValidateConfig interface.
func (p *ResourceWithValidateConfig) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
