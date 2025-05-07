package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/totftypes"
)

func LinkedResources(ctx context.Context, fw action.LinkedResources) ([]*tfprotov6.LinkedResource, diag.Diagnostics) {
	if fw == nil {
		return nil, nil
	}
	var diags diag.Diagnostics

	resp := make([]*tfprotov6.LinkedResource, len(fw))

	for _, resource := range fw {
		path, pathDiags := totftypes.AttributePath(ctx, resource.AttributePath)
		diags.Append(pathDiags...)
		if diags.HasError() {
			return nil, diags
		}

		linkedResource := &tfprotov6.LinkedResource{
			TypeName:  resource.ResourceTypeName,
			Attribute: path,
		}
		valDiag := diag.Diagnostics{}
		resp = append(resp, linkedResource)
		diags.Append(valDiag...)
	}

	return resp, diags
}
