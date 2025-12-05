// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
)

var _ list.ListResource = &ListResourceWithValidateConfig{}
var _ list.ListResourceWithValidateConfig = &ListResourceWithValidateConfig{}

// Declarative list.ListResourceWithValidateConfig for unit testing.
type ListResourceWithValidateConfig struct {
	*ListResource

	// ListResourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, list.ValidateConfigRequest, *list.ValidateConfigResponse)
}

// ValidateConfig satisfies the list.ListResourceWithValidateConfig interface.
func (p *ListResourceWithValidateConfig) ValidateListResourceConfig(ctx context.Context, req list.ValidateConfigRequest, resp *list.ValidateConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
