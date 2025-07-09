// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// InvokeAction satisfies the tfprotov6.ProviderServer interface.
func (s *Server) InvokeAction(ctx context.Context, proto6Req *tfprotov6.InvokeActionRequest) (*tfprotov6.InvokeActionServerStream, error) {
	// TODO:Actions: Implement
	panic("unimplemented")
}
