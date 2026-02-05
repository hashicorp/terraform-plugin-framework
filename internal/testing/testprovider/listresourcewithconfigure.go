// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/list"
)

var _ list.ListResource = &ListResourceWithConfigure{}
var _ list.ListResourceWithConfigure = &ListResourceWithConfigure{}

// Declarative list.ListResourceWithConfigure for unit testing.
type ListResourceWithConfigure struct {
	*ListResource

	// ListResourceWithConfigure interface methods
	ConfigureMethod func(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)
}

// Configure satisfies the list.ListResourceWithConfigure interface.
func (d *ListResourceWithConfigure) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if d.ConfigureMethod == nil {
		return
	}

	d.ConfigureMethod(ctx, req, resp)
}
