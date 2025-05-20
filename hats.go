// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package
type ResourceWithList interface {
	Resource

	ListSchema(context.Context, SchemaRequest, SchemaResponse)
	List(context.Context, ListRequest, ListResponse)
}

type ResourceWithValidateListConfig interface {
	ResourceWithList

	ValidateListConfig(
		context.Context,
		ValidateListConfigRequest,
		ValidateListConfigResponse)
}

type ListRequest struct {
	Config                tfsdk.Config
	IncludeResourceObject bool
}

type ListResponse struct {
	Results []ListResult
}

type ListResult struct {
	Identity    tfsdk.ResourceIdentity
	Resource    tfsdk.State
	DisplayName string
	Diagnostics diag.Diagnostics
}

type ValidateListConfigRequest struct {
	Config tfsdk.Config
}

type ValidateListConfigResponse struct {
	Diagnostics diag.Diagnostics
}
