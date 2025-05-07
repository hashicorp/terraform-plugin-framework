package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
)

func (s *Server) CancelAction(ctx context.Context, proto6Req *tfprotov6.CancelActionRequest) (*tfprotov6.CancelActionResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	// TODO: ask for typename in RPC request

	return &tfprotov6.CancelActionResponse{
		Diagnostics: make([]*tfprotov6.Diagnostic, 0),
	}, nil
}
