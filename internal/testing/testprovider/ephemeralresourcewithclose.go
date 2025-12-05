// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

var _ ephemeral.EphemeralResource = &EphemeralResourceWithClose{}
var _ ephemeral.EphemeralResourceWithClose = &EphemeralResourceWithClose{}

// Declarative ephemeral.EphemeralResourceWithClose for unit testing.
type EphemeralResourceWithClose struct {
	*EphemeralResource

	// EphemeralResourceWithClose interface methods
	CloseMethod func(context.Context, ephemeral.CloseRequest, *ephemeral.CloseResponse)
}

// Close satisfies the ephemeral.EphemeralResourceWithClose interface.
func (p *EphemeralResourceWithClose) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if p.CloseMethod == nil {
		return
	}

	p.CloseMethod(ctx, req, resp)
}
