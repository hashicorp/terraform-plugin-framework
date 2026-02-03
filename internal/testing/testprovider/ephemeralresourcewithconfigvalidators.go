// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

var _ ephemeral.EphemeralResource = &EphemeralResourceWithConfigValidators{}
var _ ephemeral.EphemeralResourceWithConfigValidators = &EphemeralResourceWithConfigValidators{}

// Declarative ephemeral.EphemeralResourceWithConfigValidators for unit testing.
type EphemeralResourceWithConfigValidators struct {
	*EphemeralResource

	// EphemeralResourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []ephemeral.ConfigValidator
}

// ConfigValidators satisfies the ephemeral.EphemeralResourceWithConfigValidators interface.
func (p *EphemeralResourceWithConfigValidators) ConfigValidators(ctx context.Context) []ephemeral.ConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
