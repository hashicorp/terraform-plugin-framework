// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

var _ ephemeral.EphemeralResource = &EphemeralResourceWithRenew{}
var _ ephemeral.EphemeralResourceWithRenew = &EphemeralResourceWithRenew{}

// Declarative ephemeral.EphemeralResourceWithRenew for unit testing.
type EphemeralResourceWithRenew struct {
	*EphemeralResource

	// EphemeralResourceWithRenew interface methods
	RenewMethod func(context.Context, ephemeral.RenewRequest, *ephemeral.RenewResponse)
}

// Renew satisfies the ephemeral.EphemeralResourceWithRenew interface.
func (p *EphemeralResourceWithRenew) Renew(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
	if p.RenewMethod == nil {
		return
	}

	p.RenewMethod(ctx, req, resp)
}
