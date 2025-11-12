package resource

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type ModifyConfigRequest struct {
	TypeName string

	Config tfsdk.Config
}

type ModifyConfigResponse struct {
	Config tfsdk.Config

	Diagnostics diag.Diagnostics
}
