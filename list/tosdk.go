// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package list

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r ListRequest) ToResource(ctx context.Context, val any) (*tfsdk.Resource, diag.Diagnostics) {
	attrValue, diags := reflect.FromValue(ctx, r.ResourceSchemaType, val, path.Empty())
	if diags.HasError() {
		return nil, diags
	}

	tfValue, err := attrValue.ToTerraformValue(ctx)

	if err != nil {
		diags.AddError(
			"Resource Write Error",
			"An unexpected error was encountered trying to write resource data. "+
				"This is always an error in the provider. Please report the "+
				"following to the provider developer:\n\n"+
				fmt.Sprintf("Error: Unable to run ToTerraformValue on new value: %s", err),
		)
		return nil, diags
	}

	return &tfsdk.Resource{Raw: tfValue}, diags
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
