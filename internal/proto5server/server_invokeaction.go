// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// InvokeAction satisfies the tfprotov5.ProviderServer interface.
func (s *Server) InvokeAction(ctx context.Context, proto5Req *tfprotov5.InvokeActionRequest) (*tfprotov5.InvokeActionServerStream, error) {
	// TODO:Actions: Implement
	panic("unimplemented")
}
