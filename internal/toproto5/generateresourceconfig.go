package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func GenerateResourceConfigResponse(ctx context.Context, fw *fwserver.GenerateResourceConfigResponse) *tfprotov5.GenerateResourceConfigResponse {
	if fw == nil {
		return nil
	}

	diags := Diagnostics(ctx, fw.Diagnostics)

	config, configDiags := Config(ctx, fw.Config)

	diags = append(diags, Diagnostics(ctx, configDiags)...)

	proto5 := &tfprotov5.GenerateResourceConfigResponse{
		Diagnostics: diags,
		Config:      config,
	}

	return proto5
}
