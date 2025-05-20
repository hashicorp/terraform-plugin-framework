// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"
	"iter"

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

// ListWithConfigure is an interface type that extends List to include a method
// which the framework will automatically call so provider developers have the
// opportunity to setup any necessary provider-level data or clients.
type ListWithConfigure interface {
	List
	Configure(context.Context, ConfigureRequest, *ConfigureResponse)
}

// ListWithConfigValidators is an interface type that extends List to include
// declarative validations.
//
// Declaring validation using this methodology simplifies implementation of
// reusable functionality. These also include descriptions, which can be used
// for automating documentation.
//
// Validation will include ListConfigValidators and ValidateListConfig, if both
// are implemented, in addition to any Attribute or Type validation.
type ListWithConfigValidators interface {
	List
	ListConfigValidators(context.Context) []ListConfigValidator
}

// ListWithValidateConfig is an interface type that extends List to include
// imperative validation.
//
// Declaring validation using this methodology simplifies one-off
// functionality that typically applies to a single resource. Any documentation
// of this functionality must be manually added into schema descriptions.
//
// Validation will include ListConfigValidators and ValidateListConfig, if both
// are implemented, in addition to any Attribute or Type validation.
type ListWithValidateConfig interface {
	List
	ValidateListConfig(context.Context, ValidateListConfigRequest, *ValidateListConfigResponse)
}

// ListRequest represents a request for the provider to list instances of a
// managed resource type that satisfy a user-defined request. An instance of
// this rqeuest struct is passed as an argument to the provider's ListResources
// function implementation.
type ListRequest struct {
	Config                tfsdk.Config
	IncludeResourceObject bool
}

// ListResponse represents a response to a ListRequest. An instance of this
// response struct is supplied as an argument to the provider's ListResource
// function implementation function. The provider should set an iterator
// function on the response struct.
type ListResponse struct {
	Results iter.Seq[ListResult] // Speculative + exploratory use of Go 1.23 iterators
}

// ListResult represents a managed resource instance. A provider's ListResource
// function implementation will emit zero or more results for a user-provided
// request.
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
