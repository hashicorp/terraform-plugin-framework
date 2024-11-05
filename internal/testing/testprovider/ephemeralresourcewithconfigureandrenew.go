// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

var _ ephemeral.EphemeralResource = &EphemeralResourceWithConfigureAndRenew{}
var _ ephemeral.EphemeralResourceWithConfigure = &EphemeralResourceWithConfigureAndRenew{}
var _ ephemeral.EphemeralResourceWithRenew = &EphemeralResourceWithConfigureAndRenew{}

// Declarative ephemeral.EphemeralResourceWithConfigureAndRenew for unit testing.
type EphemeralResourceWithConfigureAndRenew struct {
	*EphemeralResource

	// EphemeralResourceWithConfigure interface methods
	ConfigureMethod func(context.Context, ephemeral.ConfigureRequest, *ephemeral.ConfigureResponse)

	// EphemeralResourceWithRenew interface methods
	RenewMethod func(context.Context, ephemeral.RenewRequest, *ephemeral.RenewResponse)
}

// Configure satisfies the ephemeral.EphemeralResourceWithConfigure interface.
func (r *EphemeralResourceWithConfigureAndRenew) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}

// Renew satisfies the ephemeral.EphemeralResourceWithRenew interface.
func (r *EphemeralResourceWithConfigureAndRenew) Renew(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
	if r.RenewMethod == nil {
		return
	}

	r.RenewMethod(ctx, req, resp)
}
