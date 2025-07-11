// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
)

var _ list.ListResource = &ListResourceWithConfigure{}
var _ list.ListResourceWithConfigure = &ListResourceWithConfigure{}

// Declarative list.ListResourceWithConfigure for unit testing.
type ListResourceWithConfigure struct {
	*ListResource

	// ListResourceWithConfigure interface methods
	ConfigureMethod func(context.Context, list.ConfigureRequest, *list.ConfigureResponse)
}

// Configure satisfies the list.ListResourceWithConfigure interface.
func (d *ListResourceWithConfigure) Configure(ctx context.Context, req list.ConfigureRequest, resp *list.ConfigureResponse) {
	if d.ConfigureMethod == nil {
		return
	}

	d.ConfigureMethod(ctx, req, resp)
}
