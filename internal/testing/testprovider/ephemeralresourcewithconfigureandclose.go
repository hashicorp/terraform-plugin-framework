// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

var _ ephemeral.EphemeralResource = &EphemeralResourceWithConfigureAndClose{}
var _ ephemeral.EphemeralResourceWithConfigure = &EphemeralResourceWithConfigureAndClose{}
var _ ephemeral.EphemeralResourceWithClose = &EphemeralResourceWithConfigureAndClose{}

// Declarative ephemeral.EphemeralResourceWithConfigureAndClose for unit testing.
type EphemeralResourceWithConfigureAndClose struct {
	*EphemeralResource

	// EphemeralResourceWithConfigure interface methods
	ConfigureMethod func(context.Context, ephemeral.ConfigureRequest, *ephemeral.ConfigureResponse)

	// EphemeralResourceWithClose interface methods
	CloseMethod func(context.Context, ephemeral.CloseRequest, *ephemeral.CloseResponse)
}

// Configure satisfies the ephemeral.EphemeralResourceWithConfigure interface.
func (r *EphemeralResourceWithConfigureAndClose) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}

// Close satisfies the ephemeral.EphemeralResourceWithClose interface.
func (r *EphemeralResourceWithConfigureAndClose) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if r.CloseMethod == nil {
		return
	}

	r.CloseMethod(ctx, req, resp)
}
