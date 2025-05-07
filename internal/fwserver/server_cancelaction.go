package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
)

type CancelActionRequest struct {
	Action            action.Action
	CancellationToken string
	CancelType        action.CancelType
}

type CancelActionResponse struct {
	Diagnostics diag.Diagnostics
}

func (s *Server) CancelAction(ctx context.Context, req *CancelActionRequest, resp *CancelActionResponse) {
	if req == nil {
		return
	}

	cancelReq := action.CancelRequest{
		CancellationToken: req.CancellationToken,
		CancellationType:  req.CancelType,
	}

	cancelResp := action.CancelResponse{
		Diagnostics: resp.Diagnostics,
	}

	logging.FrameworkTrace(ctx, "Calling provider defined Action Cancel")
	req.Action.Cancel(ctx, cancelReq, &cancelResp)
	logging.FrameworkTrace(ctx, "Called provider defined Action Cancel")

	resp.Diagnostics = cancelResp.Diagnostics
}
