// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func ListResourceResult(ctx context.Context, result *fwserver.ListResult) tfprotov6.ListResourceResult {
	diags := result.Diagnostics
	if diags.HasError() {
		return tfprotov6.ListResourceResult{
			Diagnostics: Diagnostics(ctx, diags),
		}
	}

	resourceIdentity, d := ResourceIdentity(ctx, result.Identity)
	diags.Append(d...)

	return tfprotov6.ListResourceResult{
		DisplayName: result.DisplayName,
		Identity:    resourceIdentity,
		Diagnostics: Diagnostics(ctx, result.Diagnostics),
	}
}

func ListResourceResultWithResource(ctx context.Context, result *fwserver.ListResult) tfprotov6.ListResourceResult {
	diags := result.Diagnostics
	if diags.HasError() {
		return tfprotov6.ListResourceResult{
			Diagnostics: Diagnostics(ctx, diags),
		}
	}

	resourceIdentity, d := ResourceIdentity(ctx, result.Identity)
	diags.Append(d...)

	resource, d := Resource(ctx, result.Resource)
	diags.Append(d...)

	return tfprotov6.ListResourceResult{
		DisplayName: result.DisplayName,
		Identity:    resourceIdentity,
		Resource:    resource,
		Diagnostics: Diagnostics(ctx, result.Diagnostics),
	}
}
