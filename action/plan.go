// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type PlanRequest struct {
	Config tfsdk.Config
	Schema schema.Schema
}

type PlanResponse struct {
	Diagnostics   diag.Diagnostics
	PlannedConfig tfsdk.Plan
}
