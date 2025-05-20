// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// List represents an implementation of listing instances of a managed resource
// This is the core interface for all list implementations.
//
// List implementations can optionally implement these additional concepts:
//
//   - Configure: Include provider-level data or clients.
//   - Validation: Schema-based or entire configuration via
//     ListWithConfigValidators or ListWithValidateConfig.
type List interface {
	Metadata(context.Context, MetadataRequest, *MetadataResponse)
	ListSchema(context.Context, SchemaRequest, SchemaResponse)
	ListResources(context.Context, ListRequest, ListResponse)
}

type ListWithConfigure interface {
	List
	Configure(context.Context, ConfigureRequest, *ConfigureResponse)
}

type ListWithConfigValidators interface {
	List
	ListConfigValidators(context.Context) []ListConfigValidator
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

// ValidateListConfigRequest represents a request to validate the
// configuration of a resource. An instance of this request struct is
// supplied as an argument to the Resource ValidateListConfig receiver method
// or automatically passed through to each ListConfigValidator.
type ValidateListConfigRequest struct {
	// Config is the configuration the user supplied for the resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config tfsdk.Config
}

// ValidateListConfigResponse represents a response to a
// ValidateListConfigRequest. An instance of this response struct is
// supplied as an argument to the Resource ValidateListConfig receiver method
// or automatically passed through to each ListConfigValidator.
type ValidateListConfigResponse struct {
	// Diagnostics report errors or warnings related to validating the resource
	// configuration. An empty slice indicates success, with no warnings or
	// errors generated.
	Diagnostics diag.Diagnostics
}
