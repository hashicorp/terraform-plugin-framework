// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package list

import (
	"context"
	"iter"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ListResource represents an implementation of listing instances of a managed resource
// This is the core interface for all list resource implementations.
//
// ListResource implementations can optionally implement these additional concepts:
//
//   - Configure: Include provider-level data or clients.
//   - Validation: Schema-based or entire configuration via
//     ListResourceWithConfigValidators or ListResourceWithValidateConfig.
type ListResource interface {
	// Metadata should return the full name of the list resource such as
	// examplecloud_thing. This name should match the full name of the managed
	// resource to be listed; otherwise, the GetMetadata RPC will return an
	// error diagnostic.
	//
	// The method signature is intended to be compatible with the Metadata
	// method signature in the Resource interface. One implementation of
	// Metadata can satisfy both interfaces.
	Metadata(context.Context, resource.MetadataRequest, *resource.MetadataResponse)

	// ListResourceConfigSchema should return the schema for list blocks.
	ListResourceConfigSchema(context.Context, ListResourceSchemaRequest, *ListResourceSchemaResponse)

	// ListResource is called when the provider must list instances of a
	// managed resource type that satisfy a user-provided request.
	ListResource(context.Context, ListResourceRequest, *ListResourceResponse)
}

// ListResourceWithConfigure is an interface type that extends ListResource to include a method
// which the framework will automatically call so provider developers have the
// opportunity to setup any necessary provider-level data or clients.
type ListResourceWithConfigure interface {
	ListResource

	// Configure enables provider-level data or clients to be set.  The method
	// signature is intended to be compatible with the Configure method
	// signature in the Resource interface. One implementation of Configure can
	// satisfy both interfaces.
	Configure(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)
}

// ListResourceWithConfigValidators is an interface type that extends
// ListResource to include declarative validations.
//
// Declaring validation using this methodology simplifies implementation of
// reusable functionality. These also include descriptions, which can be used
// for automating documentation.
//
// Validation will include ListResourceConfigValidators and
// ValidateListResourceConfig, if both are implemented, in addition to any
// Attribute or Type validation.
type ListResourceWithConfigValidators interface {
	ListResource

	// ListResourceConfigValidators returns a list of functions which will all be performed during validation.
	ListResourceConfigValidators(context.Context) []ConfigValidator
}

// ListResourceWithValidateConfig is an interface type that extends ListResource to include
// imperative validation.
//
// Declaring validation using this methodology simplifies one-off
// functionality that typically applies to a single resource. Any documentation
// of this functionality must be manually added into schema descriptions.
//
// Validation will include ListResourceConfigValidators and ValidateListResourceConfig, if both
// are implemented, in addition to any Attribute or Type validation.
type ListResourceWithValidateConfig interface {
	ListResource

	// ValidateListResourceConfig performs the validation.
	ValidateListResourceConfig(context.Context, ValidateConfigRequest, *ValidateConfigResponse)
}

// ListResourceRequest represents a request for the provider to list instances
// of a managed resource type that satisfy a user-defined request. An instance
// of this reqeuest struct is passed as an argument to the provider's
// ListResource function implementation.
type ListResourceRequest struct {
	// Config is the configuration the user supplied for listing resource
	// instances.
	Config tfsdk.Config

	// IncludeResourceObject indicates whether the provider should populate
	// the ResourceObject field in the ListResourceEvent struct.
	IncludeResourceObject bool

	// TODO: consider applicability of:
	//
	// Private            *privatestate.ProviderData
	// ProviderMeta       tfsdk.Config
	// ClientCapabilities ReadClientCapabilities
}

// ListResourceResponse represents a response to a ListResourceRequest. An
// instance of this response struct is supplied as an argument to the
// provider's ListResource function implementation function. The provider
// should set an iterator function on the response struct.
type ListResourceResponse struct {
	// Results is a function that emits ListResourceEvent values via its yield
	// function argument.
	Results iter.Seq[ListResourceEvent]
}

// ListResourceEvent represents a listed managed resource instance. A
// provider's ListResource function implementation will emit zero or more
// events for a user-provided request.
type ListResourceEvent struct {
	// Identity is the identity of the managed resource instance.
	//
	// A nil value will raise will raise a diagnostic.
	Identity *tfsdk.ResourceIdentity

	// ResourceObject is the provider's representation of the attributes of the
	// listed managed resource instance.
	//
	// If ListResourceRequest.IncludeResourceObject is true, a nil value will raise
	// a warning diagnostic.
	ResourceObject *tfsdk.ResourceObject

	// DisplayName is a provider-defined human-readable description of the
	// listed managed resource instance, intended for CLI and browser UIs.
	DisplayName string

	// Diagnostics report errors or warnings related to the listed managed
	// resource instance. An empty slice indicates a successful operation with
	// no warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// ValidateConfigRequest represents a request to validate the configuration of
// a list resource. An instance of this request struct is supplied as an
// argument to the ValidateListResourceConfig receiver method or automatically
// passed through to each ListResourceConfigValidator.
type ValidateConfigRequest struct {
	// Config is the configuration the user supplied for the resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config tfsdk.Config
}

// ValidateConfigResponse represents a response to a ValidateConfigRequest. An
// instance of this response struct is supplied as an argument to the
// list.ValidateListResourceConfig receiver method or automatically passed
// through to each ConfigValidator.
type ValidateConfigResponse struct {
	// Diagnostics report errors or warnings related to validating the list
	// configuration. An empty slice indicates success, with no warnings
	// or errors generated.
	Diagnostics diag.Diagnostics
}
