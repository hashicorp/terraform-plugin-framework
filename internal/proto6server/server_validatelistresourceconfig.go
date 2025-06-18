// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func (s *Server) ValidateListResourceConfig(ctx context.Context, request *tfprotov6.ValidateListResourceConfigRequest) (*tfprotov6.ValidateListResourceConfigResponse, error) {
	return &tfprotov6.ValidateListResourceConfigResponse{}, nil
}
