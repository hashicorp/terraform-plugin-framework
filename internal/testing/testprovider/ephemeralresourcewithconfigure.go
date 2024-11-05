// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

var _ ephemeral.EphemeralResource = &EphemeralResourceWithConfigure{}
var _ ephemeral.EphemeralResourceWithConfigure = &EphemeralResourceWithConfigure{}

// Declarative ephemeral.EphemeralResourceWithConfigure for unit testing.
type EphemeralResourceWithConfigure struct {
	*EphemeralResource

	// EphemeralResourceWithConfigure interface methods
	ConfigureMethod func(context.Context, ephemeral.ConfigureRequest, *ephemeral.ConfigureResponse)
}

// Configure satisfies the ephemeral.EphemeralResourceWithConfigure interface.
func (d *EphemeralResourceWithConfigure) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if d.ConfigureMethod == nil {
		return
	}

	d.ConfigureMethod(ctx, req, resp)
}
