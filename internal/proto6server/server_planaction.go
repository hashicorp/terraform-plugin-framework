// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// PlanAction satisfies the tfprotov6.ProviderServer interface.
func (s *Server) PlanAction(ctx context.Context, proto6Req *tfprotov6.PlanActionRequest) (*tfprotov6.PlanActionResponse, error) {
	// TODO:Actions: Implement
	panic("unimplemented")
}
