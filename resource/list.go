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
	// Metadata should return the full name of the managed resource to be listed,
	// such as examplecloud_thing..
	Metadata(context.Context, MetadataRequest, *MetadataResponse)

	// ListConfigSchema should return the schema for list blocks.
	ListConfigSchema(context.Context, SchemaRequest, SchemaResponse)

	// ListResources is called when the provider must list instances of a
	// managed resource type that satisfy a user-provided request.
	ListResources(context.Context, ListRequest, ListResponse)
}

// ListWithConfigure is an interface type that extends List to include a method
// which the framework will automatically call so provider developers have the
// opportunity to setup any necessary provider-level data or clients.
type ListWithConfigure interface {
	List

	// Configure enables provider-level data or clients to be set.
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

	// ListConfigValidators returns a list of functions which will all be performed during validation.
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

	// ValidateListConfig performs the validation.
	ValidateListConfig(context.Context, ValidateListConfigRequest, *ValidateListConfigResponse)
}

// ListRequest represents a request for the provider to list instances of a
// managed resource type that satisfy a user-defined request. An instance of
// this rqeuest struct is passed as an argument to the provider's ListResources
// function implementation.
type ListRequest struct {
	// Config is the configuration the user supplied for listing resource
	// instances.
	Config tfsdk.Config

	// IncludeResourceObject indicates whether the provider should populate
	// the ResourceObject field in the ListResult struct.
	IncludeResourceObject bool

	// TODO: consider applicability of:
	//
	// Private            *privatestate.ProviderData
	// ProviderMeta       tfsdk.Config
	// ClientCapabilities ReadClientCapabilities
}

// ListResponse represents a response to a ListRequest. An instance of this
// response struct is supplied as an argument to the provider's ListResource
// function implementation function. The provider should set an iterator
// function on the response struct.
type ListResponse struct {
	// Results is a function that emits ListRequest values via its yield
	// function argument.
	Results iter.Seq[ListResult] // Speculative + exploratory use of Go 1.23 iterators
}

// ListResult represents a managed resource instance. A provider's ListResource
// function implementation will emit zero or more results for a user-provided
// request.
type ListResult struct {
	// Identity is the identity of the managed resource instance.
	//
	// A nil value will raise will raise a diagnostic.
	Identity *tfsdk.ResourceIdentity

	// ResourceObject is the provider's representation of all attributes of the
	// managed resource instance.
	//
	// If ListRequest.IncludeResourceObject is true, a nil value will raise
	// a warning diagnostic.
	ResourceObject *tfsdk.ResourceObject

	// DisplayName is a provider-defined human-readable description of the
	// managed resource instance, intended for CLI and browser UIs.
	DisplayName string

	// Diagnostics report errors or warnings related to listing the
	// resource. An empty slice indicates a successful operation with no
	// warnings or errors generated.
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
	// Diagnostics report errors or warnings related to validating the list
	// configuration. An empty slice indicates success, with no warnings
	// or errors generated.
	Diagnostics diag.Diagnostics
}
