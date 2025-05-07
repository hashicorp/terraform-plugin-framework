// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// PlanActionResponse returns the *tfprotov6.PlanActionResponse
// equivalent of a *fwserver.PlanActionResponse.
func PlanActionResponse(ctx context.Context, fw *fwserver.PlanActionResponse) *tfprotov6.PlanActionResponse {
	if fw == nil {
		return nil
	}

	proto6 := &tfprotov6.PlanActionResponse{
		Diagnostics: Diagnostics(ctx, fw.Diagnostics),
	}

	newConfig, diags := Plan(ctx, fw.PlannedConfig)

	proto6.Diagnostics = append(proto6.Diagnostics, Diagnostics(ctx, diags)...)
	proto6.NewConfig = newConfig

	return proto6
}

// TODO: bikeshed naming
// State returns the *tfprotov6.DynamicValue for a *tfsdk.State.
func Plan(ctx context.Context, fw *tfsdk.Plan) (*tfprotov6.DynamicValue, diag.Diagnostics) {
	if fw == nil {
		return nil, nil
	}

	data := &fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionState,
		Schema:         fw.Schema,
		TerraformValue: fw.Raw,
	}

	return DynamicValue(ctx, data)
}
