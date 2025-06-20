// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package list

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r ListRequest) ToResource(ctx context.Context, val any) (*tfsdk.Resource, diag.Diagnostics) {
	resource := &tfsdk.Resource{Schema: r.ResourceSchema}
	diags := resource.Set(ctx, val)
	return resource, diags
}

func (r ListRequest) ToIdentity(ctx context.Context, val any) (*tfsdk.ResourceIdentity, diag.Diagnostics) {
	identity := &tfsdk.ResourceIdentity{Schema: r.ResourceIdentitySchema}
	diags := identity.Set(ctx, val)

	return identity, diags
}

func (r ListRequest) ToListResult(ctx context.Context, identityVal any, resourceVal any, displayName string) ListResult {
	allDiags := diag.Diagnostics{}

	identity, diags := r.ToIdentity(ctx, identityVal)
	allDiags.Append(diags...)
	if diags.HasError() {
		return ListResult{Diagnostics: allDiags}
	}

	var resource *tfsdk.Resource
	if r.IncludeResource && resourceVal != nil {
		resource, diags = r.ToResource(ctx, resourceVal)
		allDiags.Append(diags...)
		if diags.HasError() {
			return ListResult{Diagnostics: allDiags}
		}
	}

	return ListResult{
		DisplayName: displayName,
		Resource:    resource,
		Identity:    identity,
		Diagnostics: allDiags,
	}
}
