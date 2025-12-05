// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

var _ ephemeral.EphemeralResource = &EphemeralResourceWithValidateConfig{}
var _ ephemeral.EphemeralResourceWithValidateConfig = &EphemeralResourceWithValidateConfig{}

// Declarative ephemeral.EphemeralResourceWithValidateConfig for unit testing.
type EphemeralResourceWithValidateConfig struct {
	*EphemeralResource

	// EphemeralResourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, ephemeral.ValidateConfigRequest, *ephemeral.ValidateConfigResponse)
}

// ValidateConfig satisfies the ephemeral.EphemeralResourceWithValidateConfig interface.
func (p *EphemeralResourceWithValidateConfig) ValidateConfig(ctx context.Context, req ephemeral.ValidateConfigRequest, resp *ephemeral.ValidateConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
