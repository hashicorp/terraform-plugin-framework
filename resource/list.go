// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type List interface {
	Metadata(context.Context, MetadataRequest, *MetadataResponse)
	ListSchema(context.Context, SchemaRequest, SchemaResponse)
	ListResources(context.Context, ListRequest, ListResponse)
}

type ListWithConfigure interface {
	List
	Configure(context.Context, ConfigureRequest, *ConfigureResponse)
}

type ListWithValidateConfig interface {
	List
	ValidateListConfig(context.Context, ValidateListConfigRequest, *ValidateListConfigResponse)
}

type ListRequest struct {
	Config                tfsdk.Config
	IncludeResourceObject bool
}

type ListResponse struct {
	Results []ListResult // TODO: streamify
}

type ListResult struct {
	Identity    tfsdk.ResourceIdentity
	Resource    tfsdk.ResourceObject
	DisplayName string
	Diagnostics diag.Diagnostics
}

type ValidateListConfigRequest struct {
	Config tfsdk.Config
}

type ValidateListConfigResponse struct {
	Diagnostics diag.Diagnostics
}
